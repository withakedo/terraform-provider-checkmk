package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/withakedo/terraform-provider-checkmk/internal/common"
)

var (
	_ resource.Resource                   = &ServiceDiscoveryResource{}
	_ resource.ResourceWithValidateConfig = &ServiceDiscoveryResource{}
)

func NewServiceDiscoveryResource() resource.Resource {
	return &ServiceDiscoveryResource{}
}

// ServiceDiscoveryResource triggers CheckMK's service discovery for a host.
// Like ActivationResource, this is an action/trigger resource rather than a
// persistent CheckMK object: there is nothing server-side to "own" and clean
// up, so Read is a point-in-time no-op and Delete does nothing.
type ServiceDiscoveryResource struct {
	providerData *common.ProviderData
}

// ServiceDiscoveryResourceModel describes the resource data model.
type ServiceDiscoveryResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	HostName            types.String `tfsdk:"host_name"`
	Mode                types.String `tfsdk:"mode"`
	Activate            types.String `tfsdk:"activate"`
	ForceForeignChanges types.Bool   `tfsdk:"force_foreign_changes"`

	// Computed outputs
	State        types.String `tfsdk:"state"`
	DiscoveredAt types.String `tfsdk:"discovered_at"`
}

func (r *ServiceDiscoveryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_discovery"
}

func (r *ServiceDiscoveryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Triggers CheckMK service discovery for a host, so newly added or removed
services are found without a manual step in the CheckMK UI.

Re-running ` + "`terraform apply`" + ` re-triggers discovery (e.g. after a schedule, or when you
change ` + "`mode`" + `). Modes that auto-apply changes (` + "`fix_all`" + `, ` + "`tabula_rasa`" + `,
` + "`new`" + `, ` + "`remove`" + `) modify the host's active service configuration and are tracked
like any other change for the provider's activation handling; ` + "`refresh`" + ` (the default) only
updates the discovery preview and does not require activation.
`,
		Attributes: map[string]schema.Attribute{
			"id": common.ComputedIDAttribute("Unique identifier for this service discovery resource instance."),
			"host_name": schema.StringAttribute{
				MarkdownDescription: "Name of the host to run service discovery against.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "Discovery mode (default: `refresh`). One of: " +
					"`new` (add newly found services), `remove` (remove vanished services), " +
					"`fix_all` (add new and remove vanished), `refresh` (rediscover everything, " +
					"preview only), `only_host_labels` (update host labels only), " +
					"`only_service_labels` (update service labels only), `tabula_rasa` (remove all " +
					"and rediscover from scratch).",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("refresh"),
			},
			"activate":              common.ActivateAttribute(),
			"force_foreign_changes": common.ForceForeignChangesAttribute(),

			"state": schema.StringAttribute{
				MarkdownDescription: "Final state of the discovery run as reported by CheckMK " +
					"(`finished`, `exception`, or `stopped`).",
				Computed: true,
			},
			"discovered_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp (RFC3339) when discovery last completed.",
				Computed:            true,
			},
		},
	}
}

func (r *ServiceDiscoveryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.providerData = common.ConfigureResource(req, resp)
}

func (r *ServiceDiscoveryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceDiscoveryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("discovery-%s", data.HostName.ValueString()))

	diags := r.runDiscovery(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceDiscoveryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceDiscoveryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Discovery is a point-in-time trigger, not a persistent object we can
	// re-read from the API - the recorded state reflects the last run.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceDiscoveryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServiceDiscoveryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.runDiscovery(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceDiscoveryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// There is no server-side object to clean up: discovery just triggers a
	// scan of the host's already-configured (or already-removed) services.
}

// runDiscovery triggers a discovery run for the host in data, waits for
// completion, and updates data with the resulting state. It tracks the
// change for the provider's pending-change/activation handling, since
// auto-applying discovery modes change the host's service configuration.
func (r *ServiceDiscoveryResource) runDiscovery(ctx context.Context, data *ServiceDiscoveryResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	hostName := data.HostName.ValueString()
	mode := data.Mode.ValueString()

	tflog.Info(ctx, "Triggering CheckMK service discovery", map[string]interface{}{
		"host_name": hostName,
		"mode":      mode,
	})

	extensions, err := r.providerData.Client.DiscoverServices(ctx, hostName, mode)
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to run service discovery for host %q: %s", hostName, err),
		)
		return diags
	}

	data.State = types.StringValue(extensions.State)
	data.DiscoveredAt = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	cfg := common.BuildBaseConfig(r.providerData, data.Activate, data.ForceForeignChanges, types.BoolNull())
	if err := common.TrackAndActivate(ctx, r.providerData, cfg, "service_discovery"); err != nil {
		diags.AddWarning(
			"Activation Warning",
			fmt.Sprintf("Service discovery completed but activation failed: %s", err),
		)
	}

	tflog.Info(ctx, "Service discovery completed", map[string]interface{}{
		"host_name": hostName,
		"state":     extensions.State,
	})

	return diags
}

// ValidateConfig validates the resource configuration against the generated
// OpenAPI types for the connected CheckMK version.
func (r *ServiceDiscoveryResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.providerData == nil {
		return
	}

	var data ServiceDiscoveryResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	validator := common.NewAttributeValidator(r.providerData)
	resp.Diagnostics.Append(validator.ValidateStringField("DiscoverServices", "mode", data.Mode, path.Root("mode"))...)
}
