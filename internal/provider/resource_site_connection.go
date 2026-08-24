package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/withakedo/terraform-provider-checkmk/internal/common"
)

var (
	_ resource.Resource                = &SiteConnectionResource{}
	_ resource.ResourceWithImportState = &SiteConnectionResource{}
)

func NewSiteConnectionResource() resource.Resource {
	return &SiteConnectionResource{}
}

// SiteConnectionResource manages a CheckMK site connection for distributed
// monitoring (connecting this site to a remote monitoring site). Like
// checkmk_bi_aggregation, the connection's configuration body is
// deliberately handled as raw JSON rather than a typed nested schema - see
// config_raw below - since CheckMK's site connection schema is large,
// deeply nested, and varies across versions/editions.
//
// This resource manages the connection's configuration only. It does not
// perform the separate remote-site login/logout actions (which exchange
// admin credentials for the remote site and are not idempotent
// configuration in the same sense); log in via the CheckMK UI or API after
// creating the connection if replication is enabled.
type SiteConnectionResource struct {
	providerData *common.ProviderData
}

// SiteConnectionResourceModel describes the resource data model.
type SiteConnectionResourceModel struct {
	ID        types.String `tfsdk:"id"`
	SiteID    types.String `tfsdk:"site_id"`
	ConfigRaw types.String `tfsdk:"config_raw"`
	LoggedIn  types.Bool   `tfsdk:"logged_in"`
	Activate  types.String `tfsdk:"activate"`
}

func (r *SiteConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_connection"
}

func (r *SiteConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a CheckMK site connection for distributed monitoring - connecting this
site to a remote monitoring site so its hosts/services can be included in this site's overview.

The connection's configuration (` + "`basic_settings`" + `, ` + "`configuration_connection`" + `,
` + "`status_connection`" + `) is deliberately handled as raw JSON in ` + "`config_raw`" + ` rather
than a typed schema, mirroring ` + "`checkmk_rule`" + ` and ` + "`checkmk_bi_aggregation`" + `: the
underlying schema is large, deeply nested, and varies across CheckMK versions/editions. See the
CheckMK REST API documentation for the ` + "`SiteConnectionCreate`" + ` schema for available fields.

This resource manages connection configuration only - it does not perform the separate remote-site
login/logout actions (which exchange admin credentials for the remote site). Log in via the
CheckMK UI or API after creating the connection if replication is enabled; ` + "`logged_in`" + `
reflects that state but cannot be set through this resource.

**Known limitation:** on read, ` + "`config_raw`" + ` is refreshed from the API's response, which
may include CheckMK-filled default fields you didn't specify in your original JSON. If CheckMK's
defaults don't round-trip byte-for-byte through ` + "`jsonencode`" + `, this can show a perpetual
diff; if that happens, add the defaulted fields explicitly to your configuration.
`,
		Attributes: map[string]schema.Attribute{
			"id": common.ComputedIDAttribute("Unique identifier for this resource (same as site_id)."),
			"site_id": schema.StringAttribute{
				MarkdownDescription: "Unique id for the site connection. Used as the resource identity.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config_raw": schema.StringAttribute{
				MarkdownDescription: "The site connection configuration as a JSON-encoded string (typically built " +
					"with `jsonencode(...)`), covering `basic_settings`, `configuration_connection`, and " +
					"`status_connection`. `basic_settings.site_id` is set automatically from `site_id`.",
				Required: true,
			},
			"logged_in": schema.BoolAttribute{
				MarkdownDescription: "Whether the remote site is currently logged in. Read-only; log in via the " +
					"CheckMK UI or API, not through this resource.",
				Computed: true,
			},
			"activate": common.ActivateAttribute(),
		},
	}
}

func (r *SiteConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.providerData = common.ConfigureResource(req, resp)
}

func (r *SiteConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SiteConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildBaseConfig(r.providerData, data.Activate, types.BoolNull(), types.BoolNull())

	siteConfig, err := decodeSiteConfig(data.ConfigRaw.ValueString(), data.SiteID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("config_raw"), "Invalid config_raw", err.Error())
		return
	}

	tflog.Info(ctx, "Creating CheckMK site connection", map[string]interface{}{"site_id": data.SiteID.ValueString()})

	result, err := r.providerData.Client.CreateSiteConnection(ctx, map[string]interface{}{"site_config": siteConfig})
	if err != nil {
		common.AddClientError(resp, "create", "site connection", err)
		return
	}

	if err := common.TrackAndActivate(ctx, r.providerData, cfg, "site_connection"); err != nil {
		common.AddActivationWarning(resp, "Site connection", "created", err)
	}

	data.ID = types.StringValue(data.SiteID.ValueString())
	applySiteConnectionResult(&data, result)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SiteConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SiteConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.providerData.Client.GetSiteConnection(ctx, data.SiteID.ValueString())
	if err != nil {
		if common.HandleNotFoundOnRead(err, resp) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read site connection: %s", err))
		return
	}

	data.ID = types.StringValue(data.SiteID.ValueString())
	applySiteConnectionResult(&data, result)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SiteConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SiteConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildBaseConfig(r.providerData, data.Activate, types.BoolNull(), types.BoolNull())

	siteConfig, err := decodeSiteConfig(data.ConfigRaw.ValueString(), data.SiteID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("config_raw"), "Invalid config_raw", err.Error())
		return
	}

	tflog.Info(ctx, "Updating CheckMK site connection", map[string]interface{}{"site_id": data.SiteID.ValueString()})

	result, err := r.providerData.Client.UpdateSiteConnection(ctx, data.SiteID.ValueString(), map[string]interface{}{"site_config": siteConfig})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update site connection: %s", err))
		return
	}

	if err := common.TrackAndActivate(ctx, r.providerData, cfg, "site_connection"); err != nil {
		common.AddActivationWarning(resp, "Site connection", "updated", err)
	}

	data.ID = types.StringValue(data.SiteID.ValueString())
	applySiteConnectionResult(&data, result)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SiteConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SiteConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildBaseConfig(r.providerData, data.Activate, types.BoolNull(), types.BoolNull())

	if err := r.providerData.Client.DeleteSiteConnection(ctx, data.SiteID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete site connection: %s", err))
		return
	}

	if err := common.TrackAndActivate(ctx, r.providerData, cfg, "site_connection"); err != nil {
		common.AddActivationWarning(resp, "Site connection", "deleted", err)
	}
}

func (r *SiteConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("site_id"), req, resp)
}

// decodeSiteConfig parses config_raw and forces basic_settings.site_id to
// the resource's site_id, so there is a single source of truth for the
// connection's identity.
func decodeSiteConfig(raw, siteID string) (map[string]interface{}, error) {
	var siteConfig map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &siteConfig); err != nil {
		return nil, fmt.Errorf("must be valid JSON: %w", err)
	}

	basicSettings, ok := siteConfig["basic_settings"].(map[string]interface{})
	if !ok {
		basicSettings = make(map[string]interface{})
	}
	basicSettings["site_id"] = siteID
	siteConfig["basic_settings"] = basicSettings

	return siteConfig, nil
}

// applySiteConnectionResult updates data from an API response's extensions
// (status_connection, configuration_connection, basic_settings, logged_in).
func applySiteConnectionResult(data *SiteConnectionResourceModel, body map[string]interface{}) {
	extensions, ok := body["extensions"].(map[string]interface{})
	if !ok {
		extensions = body
	}

	if loggedIn, ok := extensions["logged_in"].(bool); ok {
		data.LoggedIn = types.BoolValue(loggedIn)
	} else {
		data.LoggedIn = types.BoolNull()
	}

	configRaw := make(map[string]interface{}, 3)
	for _, key := range []string{"basic_settings", "configuration_connection", "status_connection"} {
		if v, ok := extensions[key]; ok {
			configRaw[key] = v
		}
	}
	if raw, err := json.Marshal(configRaw); err == nil {
		data.ConfigRaw = types.StringValue(string(raw))
	}
}
