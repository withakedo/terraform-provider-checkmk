// Package labels provides label rule resources (host_labels, service_labels).
// This file contains the generic base implementation for label resources.
package labels

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/terraform-provider-checkmk/internal/client"
	"github.com/terraform-provider-checkmk/internal/common"
	"github.com/terraform-provider-checkmk/internal/rules"
)

// LabelResourceModel is the shared data model for all label resources.
type LabelResourceModel struct {
	ID         types.String `tfsdk:"id"`
	APIID      types.String `tfsdk:"api_id"`
	Folder     types.String `tfsdk:"folder"`
	Labels     types.Map    `tfsdk:"labels"`
	Properties types.Object `tfsdk:"properties"`
	Conditions types.Object `tfsdk:"conditions"`
}

// LabelPropertiesModel describes rule properties for labels.
type LabelPropertiesModel struct {
	Description types.String `tfsdk:"description"`
	Comment     types.String `tfsdk:"comment"`
	Disabled    types.Bool   `tfsdk:"disabled"`
}

// LabelResourceConfig configures a label resource type.
type LabelResourceConfig struct {
	// TypeName is the Terraform resource type suffix (e.g., "host_labels", "service_labels")
	TypeName string
	// Ruleset is the CheckMK ruleset name (e.g., "host_label_rules", "service_label_rules")
	Ruleset string
	// Description is the resource's markdown description
	Description string
	// ResourceLabel is the label used in messages (e.g., "host", "service")
	ResourceLabel string
}

// BaseLabelResource is the generic label resource implementation.
type BaseLabelResource struct {
	config       LabelResourceConfig
	providerData *common.ProviderData
}

// NewBaseLabelResource creates a new label resource with the given configuration.
func NewBaseLabelResource(config LabelResourceConfig) resource.Resource {
	return &BaseLabelResource{config: config}
}

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &BaseLabelResource{}
var _ resource.ResourceWithImportState = &BaseLabelResource{}

func (r *BaseLabelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.TypeName
}

func (r *BaseLabelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: r.config.Description,
		Attributes: map[string]schema.Attribute{
			// Note: We don't use UseStateForUnknown here because the ID can change
			// when description or conditions change. This will show "(known after apply)"
			// in plan output but correctly handles ID changes.
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed hash based on description and conditions. Used for identity.",
				Computed:            true,
			},
			"api_id": schema.StringAttribute{
				MarkdownDescription: "CheckMK's internal UUID for the rule. Used for updates and deletes.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"folder": schema.StringAttribute{
				MarkdownDescription: "The folder path where the rule will be created. Default is '/' (root folder).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: fmt.Sprintf("Map of labels to assign to matching %ss. Keys and values are strings.", r.config.ResourceLabel),
				Required:            true,
				ElementType:         types.StringType,
			},
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Rule properties including description, comment, and disabled status.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						MarkdownDescription: "Human-readable description of the rule. Required and part of rule identity.",
						Required:            true,
					},
					"comment": schema.StringAttribute{
						MarkdownDescription: "Optional comment for the rule.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"disabled": schema.BoolAttribute{
						MarkdownDescription: "Whether the rule is disabled. Default is false.",
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"conditions": rules.RuleConditionsSchema(),
		},
	}
}

func (r *BaseLabelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.providerData = common.ConfigureResource(req, resp)
}

func (r *BaseLabelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LabelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildSimpleBaseConfig(r.providerData, types.BoolNull())

	// Extract properties
	var properties LabelPropertiesModel
	resp.Diagnostics.Append(data.Properties.As(ctx, &properties, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract conditions
	conditions, err := rules.ConvertConditionsToClient(ctx, data.Conditions)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert conditions: %s", err))
		return
	}

	// Ensure conditions is not nil (API requires at least empty object)
	if conditions == nil {
		conditions = make(map[string]interface{})
	}

	// Validate no circular dependencies
	validationErrors := rules.ValidateRulesetConditions(r.config.Ruleset, conditions)
	if len(validationErrors) > 0 {
		for _, validationErr := range validationErrors {
			resp.Diagnostics.AddError("Validation Error", validationErr)
		}
		return
	}

	// Convert labels map to value_raw JSON
	labelsMap := make(map[string]string)
	resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelsMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	valueRaw, err := json.Marshal(labelsMap)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal labels: %s", err))
		return
	}

	// Set default folder if not specified
	folder := "/"
	if !data.Folder.IsNull() && !data.Folder.IsUnknown() && data.Folder.ValueString() != "" {
		folder = data.Folder.ValueString()
	}

	// Create rule via API
	createReq := &client.RuleCreateRequest{
		Ruleset: r.config.Ruleset,
		Folder:  folder,
		Properties: client.RuleProperties{
			Description: properties.Description.ValueString(),
			Comment:     properties.Comment.ValueString(),
			Disabled:    properties.Disabled.ValueBool(),
		},
		ValueRaw:   string(valueRaw),
		Conditions: conditions,
	}

	rule, err := r.providerData.Client.CreateRule(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create %s rule: %s", r.config.TypeName, err))
		return
	}

	// Generate deterministic hash for ID
	hash := client.GenerateRuleHash(r.config.Ruleset, properties.Description.ValueString(), conditions)

	// Activate changes if configured
	if err := common.TrackAndActivate(ctx, r.providerData, cfg, r.config.TypeName); err != nil {
		common.AddActivationWarning(resp, r.config.TypeName, "created", err)
	}

	// Set computed values
	data.ID = types.StringValue(hash)
	data.APIID = types.StringValue(rule.ID)
	if data.Folder.IsNull() {
		data.Folder = types.StringValue(rule.Extensions.Folder)
	}

	// Update properties from API response
	properties.Comment = types.StringValue(rule.Extensions.Properties.Comment)
	properties.Disabled = types.BoolValue(rule.Extensions.Properties.Disabled)

	propertiesObj, diags := types.ObjectValueFrom(ctx, data.Properties.AttributeTypes(ctx), properties)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Properties = propertiesObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaseLabelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LabelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.APIID.IsNull() || data.APIID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing API ID",
			"Rule API ID is missing from state. This should not happen. Please report this issue.",
		)
		return
	}

	rule, err := r.providerData.Client.GetRule(ctx, data.APIID.ValueString())
	if err != nil {
		if common.HandleNotFoundOnRead(err, resp) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read %s rule: %s", r.config.TypeName, err))
		return
	}

	// Initialize properties struct (may be null during import)
	var properties LabelPropertiesModel
	if !data.Properties.IsNull() && !data.Properties.IsUnknown() {
		resp.Diagnostics.Append(data.Properties.As(ctx, &properties, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Preserve the ID from state unless this is an import (where ID might be a placeholder like "imported")
	// During import, generate a new hash from API data
	if data.ID.IsNull() || data.ID.IsUnknown() || data.ID.ValueString() == "imported" {
		hash := client.GenerateRuleHash(
			rule.Extensions.Ruleset,
			rule.Extensions.Properties.Description,
			rule.Extensions.Conditions,
		)
		data.ID = types.StringValue(hash)
	}
	// Otherwise, keep the existing ID from state

	// Update state with API data
	data.APIID = types.StringValue(rule.ID)
	data.Folder = types.StringValue(rule.Extensions.Folder)

	// Parse value_raw back to labels map
	labelsMap, err := parseLabelsFromValueRaw(rule.Extensions.ValueRaw)
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse labels from value_raw: %s", err))
		return
	}
	labelsMapValue, diags := types.MapValueFrom(ctx, types.StringType, labelsMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Labels = labelsMapValue

	// Update properties
	properties.Description = types.StringValue(rule.Extensions.Properties.Description)
	properties.Comment = types.StringValue(rule.Extensions.Properties.Comment)
	properties.Disabled = types.BoolValue(rule.Extensions.Properties.Disabled)

	// Define property attribute types (needed for import when data.Properties is null)
	propertiesAttrTypes := map[string]attr.Type{
		"description": types.StringType,
		"comment":     types.StringType,
		"disabled":    types.BoolType,
	}

	propertiesObj, diags := types.ObjectValueFrom(ctx, propertiesAttrTypes, properties)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Properties = propertiesObj

	// For import, initialize conditions as null object with correct type
	if data.Conditions.IsNull() || data.Conditions.IsUnknown() {
		data.Conditions = types.ObjectNull(rules.RuleConditionsAttrTypes())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaseLabelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LabelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildSimpleBaseConfig(r.providerData, types.BoolNull())

	// Extract properties
	var properties LabelPropertiesModel
	resp.Diagnostics.Append(data.Properties.As(ctx, &properties, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract conditions
	conditions, err := rules.ConvertConditionsToClient(ctx, data.Conditions)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert conditions: %s", err))
		return
	}

	// Ensure conditions is not nil (API requires at least empty object)
	if conditions == nil {
		conditions = make(map[string]interface{})
	}

	// Validate no circular dependencies
	validationErrors := rules.ValidateRulesetConditions(r.config.Ruleset, conditions)
	if len(validationErrors) > 0 {
		for _, validationErr := range validationErrors {
			resp.Diagnostics.AddError("Validation Error", validationErr)
		}
		return
	}

	// Convert labels map to value_raw JSON
	labelsMap := make(map[string]string)
	resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelsMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	valueRaw, err := json.Marshal(labelsMap)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal labels: %s", err))
		return
	}

	// Fetch ETag if strict locking is enabled
	etag, err := common.FetchETagIfStrict(ctx, cfg.StrictResourceLocking, func(ctx context.Context) (string, error) {
		result, err := r.providerData.Client.GetRuleWithETag(ctx, data.APIID.ValueString())
		if err != nil {
			return "", err
		}
		return result.ETag, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch rule ETag: %s", err))
		return
	}

	updateReq := &client.RuleUpdateRequest{
		Properties: client.RuleProperties{
			Description: properties.Description.ValueString(),
			Comment:     properties.Comment.ValueString(),
			Disabled:    properties.Disabled.ValueBool(),
		},
		ValueRaw:   string(valueRaw),
		Conditions: conditions,
	}

	rule, err := r.providerData.Client.UpdateRule(ctx, data.APIID.ValueString(), updateReq, etag)
	if err != nil {
		if !common.HandleDriftWarning(err, resp, data.APIID.ValueString()) {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update %s rule: %s", r.config.TypeName, err))
			return
		}
	}

	// Activate changes if configured
	if err := common.TrackAndActivate(ctx, r.providerData, cfg, r.config.TypeName); err != nil {
		common.AddActivationWarning(resp, r.config.TypeName, "updated", err)
	}

	// Update state with plan values and API response
	hash := client.GenerateRuleHash(r.config.Ruleset, properties.Description.ValueString(), conditions)
	data.ID = types.StringValue(hash)

	if rule != nil {
		data.APIID = types.StringValue(rule.ID)
		data.Folder = types.StringValue(rule.Extensions.Folder)

		properties.Comment = types.StringValue(rule.Extensions.Properties.Comment)
		properties.Disabled = types.BoolValue(rule.Extensions.Properties.Disabled)

		propertiesObj, diags := types.ObjectValueFrom(ctx, data.Properties.AttributeTypes(ctx), properties)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Properties = propertiesObj
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaseLabelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LabelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildSimpleBaseConfig(r.providerData, types.BoolNull())

	// Fetch ETag if strict locking is enabled
	etag, err := common.FetchETagIfStrict(ctx, cfg.StrictResourceLocking, func(ctx context.Context) (string, error) {
		result, err := r.providerData.Client.GetRuleWithETag(ctx, data.APIID.ValueString())
		if err != nil {
			if apiErr, ok := err.(*client.APIError); ok && apiErr.Status == 404 {
				return "", nil
			}
			return "", err
		}
		return result.ETag, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch rule ETag: %s", err))
		return
	}

	if err := r.providerData.Client.DeleteRule(ctx, data.APIID.ValueString(), etag); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete %s rule: %s", r.config.TypeName, err))
		return
	}

	// Activate changes if configured
	if err := common.TrackAndActivate(ctx, r.providerData, cfg, r.config.TypeName); err != nil {
		common.AddActivationWarning(resp, r.config.TypeName, "deleted", err)
	}
}

func (r *BaseLabelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("api_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "imported")...)
}

// =============================================================================
// Pre-configured Label Resources
// =============================================================================

// HostLabelsConfig returns the configuration for host labels resources.
var HostLabelsConfig = LabelResourceConfig{
	TypeName:      "host_labels",
	Ruleset:       "host_label_rules",
	ResourceLabel: "host",
	Description: "Assigns labels to hosts matching specified conditions. " +
		"This is a typed wrapper around the `host_label_rules` ruleset. " +
		"Requires activation.\n\n" +
		"**Note:** Host label rules cannot use host_labels or host_label_groups in conditions " +
		"(would create circular dependencies).",
}

// ServiceLabelsConfig returns the configuration for service labels resources.
var ServiceLabelsConfig = LabelResourceConfig{
	TypeName:      "service_labels",
	Ruleset:       "service_label_rules",
	ResourceLabel: "service",
	Description: "Assigns labels to services matching specified conditions. " +
		"This is a typed wrapper around the `service_label_rules` ruleset. " +
		"Requires activation.\n\n" +
		"**Note:** Service label rules cannot use service_labels or service_label_groups in conditions " +
		"(would create circular dependencies).",
}
