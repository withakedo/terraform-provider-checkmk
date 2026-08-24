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
	_ resource.Resource                = &BIAggregationResource{}
	_ resource.ResourceWithImportState = &BIAggregationResource{}
)

func NewBIAggregationResource() resource.Resource {
	return &BIAggregationResource{}
}

// BIAggregationResource manages a CheckMK Business Intelligence aggregation
// (a health-status rollup over a tree of hosts/services, e.g. "all services
// of the shop cluster are OK"). The aggregation's rule tree ("node") is
// deliberately handled as raw JSON rather than a typed nested schema - see
// definition_raw below - matching the same design choice checkmk_rule makes
// for arbitrary ruleset values.
type BIAggregationResource struct {
	providerData *common.ProviderData
}

// BIAggregationResourceModel describes the resource data model.
type BIAggregationResourceModel struct {
	ID            types.String `tfsdk:"id"`
	AggregationID types.String `tfsdk:"aggregation_id"`
	PackID        types.String `tfsdk:"pack_id"`
	DefinitionRaw types.String `tfsdk:"definition_raw"`
	Activate      types.String `tfsdk:"activate"`
}

func (r *BIAggregationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bi_aggregation"
}

func (r *BIAggregationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a CheckMK Business Intelligence (BI) aggregation - a health-status
rollup computed from a tree of hosts and/or services (e.g. "the shop is down if any of its
critical services are down").

BI aggregation rule trees (` + "`node`" + `) are recursive and highly variable - a node can be a
leaf that checks a host/service state, or a nested aggregation combining other nodes - so this
resource takes the whole aggregation body (everything except ` + "`id`" + ` and ` + "`pack_id`" + `,
which are set from ` + "`aggregation_id`" + ` and ` + "`pack_id`" + `) as raw JSON in
` + "`definition_raw`" + ` rather than a typed schema, mirroring how ` + "`checkmk_rule`" + ` handles
arbitrary ruleset values with ` + "`value_raw`" + `. See the CheckMK REST API documentation for the
` + "`BIAggregationEndpoint`" + ` schema for the fields available inside ` + "`definition_raw`" + `
(typically ` + "`comment`" + `, ` + "`groups`" + `, ` + "`node`" + `, ` + "`aggregation_visualization`" + `,
and ` + "`computation_options`" + `).

**Known limitation:** on read, ` + "`definition_raw`" + ` is refreshed from the API's response,
which may include CheckMK-filled default fields you didn't specify in your original JSON. If
CheckMK's defaults don't round-trip byte-for-byte through ` + "`jsonencode`" + `, this can show a
perpetual diff; if that happens, add the defaulted fields explicitly to your configuration.
`,
		Attributes: map[string]schema.Attribute{
			"id": common.ComputedIDAttribute("Unique identifier for this resource (same as aggregation_id)."),
			"aggregation_id": schema.StringAttribute{
				MarkdownDescription: "Unique id for the aggregation. Used as the resource identity.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pack_id": schema.StringAttribute{
				MarkdownDescription: "The BI pack this aggregation belongs to. Must already exist in CheckMK.",
				Required:            true,
			},
			"definition_raw": schema.StringAttribute{
				MarkdownDescription: "The aggregation definition as a JSON-encoded string (typically built with " +
					"`jsonencode(...)`), covering everything in the `BIAggregationEndpoint` schema except `id` " +
					"and `pack_id`. Common fields: `comment`, `groups`, `node`, `aggregation_visualization`, " +
					"`computation_options`.",
				Required: true,
			},
			"activate": common.ActivateAttribute(),
		},
	}
}

func (r *BIAggregationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.providerData = common.ConfigureResource(req, resp)
}

func (r *BIAggregationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BIAggregationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildBaseConfig(r.providerData, data.Activate, types.BoolNull(), types.BoolNull())

	body, err := decodeAggregationDefinition(data.DefinitionRaw.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("definition_raw"), "Invalid definition_raw", err.Error())
		return
	}
	aggregationID := data.AggregationID.ValueString()
	body["id"] = aggregationID
	body["pack_id"] = data.PackID.ValueString()

	tflog.Info(ctx, "Creating CheckMK BI aggregation", map[string]interface{}{"aggregation_id": aggregationID})

	result, err := r.providerData.Client.CreateBIAggregation(ctx, aggregationID, body)
	if err != nil {
		common.AddClientError(resp, "create", "BI aggregation", err)
		return
	}

	if err := common.TrackAndActivate(ctx, r.providerData, cfg, "bi_aggregation"); err != nil {
		common.AddActivationWarning(resp, "BI aggregation", "created", err)
	}

	data.ID = types.StringValue(aggregationID)
	if raw, err := encodeAggregationDefinition(result); err == nil {
		data.DefinitionRaw = types.StringValue(raw)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BIAggregationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BIAggregationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.providerData.Client.GetBIAggregation(ctx, data.AggregationID.ValueString())
	if err != nil {
		if common.HandleNotFoundOnRead(err, resp) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read BI aggregation: %s", err))
		return
	}

	data.ID = types.StringValue(data.AggregationID.ValueString())
	if packID, ok := extensionsPackID(result); ok {
		data.PackID = types.StringValue(packID)
	}
	if raw, err := encodeAggregationDefinition(result); err == nil {
		data.DefinitionRaw = types.StringValue(raw)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BIAggregationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BIAggregationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildBaseConfig(r.providerData, data.Activate, types.BoolNull(), types.BoolNull())

	body, err := decodeAggregationDefinition(data.DefinitionRaw.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("definition_raw"), "Invalid definition_raw", err.Error())
		return
	}
	aggregationID := data.AggregationID.ValueString()
	body["id"] = aggregationID
	body["pack_id"] = data.PackID.ValueString()

	tflog.Info(ctx, "Updating CheckMK BI aggregation", map[string]interface{}{"aggregation_id": aggregationID})

	result, err := r.providerData.Client.UpdateBIAggregation(ctx, aggregationID, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update BI aggregation: %s", err))
		return
	}

	if err := common.TrackAndActivate(ctx, r.providerData, cfg, "bi_aggregation"); err != nil {
		common.AddActivationWarning(resp, "BI aggregation", "updated", err)
	}

	data.ID = types.StringValue(aggregationID)
	if raw, err := encodeAggregationDefinition(result); err == nil {
		data.DefinitionRaw = types.StringValue(raw)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BIAggregationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BIAggregationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildBaseConfig(r.providerData, data.Activate, types.BoolNull(), types.BoolNull())

	if err := r.providerData.Client.DeleteBIAggregation(ctx, data.AggregationID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete BI aggregation: %s", err))
		return
	}

	if err := common.TrackAndActivate(ctx, r.providerData, cfg, "bi_aggregation"); err != nil {
		common.AddActivationWarning(resp, "BI aggregation", "deleted", err)
	}
}

func (r *BIAggregationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("aggregation_id"), req, resp)
}

// decodeAggregationDefinition parses definition_raw into a mutable map so
// id/pack_id can be injected before sending it to the API.
func decodeAggregationDefinition(raw string) (map[string]interface{}, error) {
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		return nil, fmt.Errorf("must be valid JSON: %w", err)
	}
	return body, nil
}

// encodeAggregationDefinition re-serializes an API response's extensions
// (minus id/pack_id, which are tracked as separate typed attributes) back
// into definition_raw for state / drift detection.
func encodeAggregationDefinition(body map[string]interface{}) (string, error) {
	extensions, ok := body["extensions"].(map[string]interface{})
	if !ok {
		extensions = body
	}
	filtered := make(map[string]interface{}, len(extensions))
	for k, v := range extensions {
		if k == "id" || k == "pack_id" {
			continue
		}
		filtered[k] = v
	}
	raw, err := json.Marshal(filtered)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func extensionsPackID(body map[string]interface{}) (string, bool) {
	extensions, ok := body["extensions"].(map[string]interface{})
	if !ok {
		extensions = body
	}
	packID, ok := extensions["pack_id"].(string)
	return packID, ok
}
