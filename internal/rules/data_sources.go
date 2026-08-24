// Package rules provides data sources for reading CheckMK rules.
package rules

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/withakedo/terraform-provider-checkmk/internal/common"
)

// =============================================================================
// Rule Data Source
// =============================================================================

var _ datasource.DataSource = &RuleDataSource{}

func NewRuleDataSource() datasource.DataSource {
	return &RuleDataSource{}
}

type RuleDataSource struct {
	providerData *common.ProviderData
}

type RuleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	RuleID      types.String `tfsdk:"rule_id"`
	Ruleset     types.String `tfsdk:"ruleset"`
	Folder      types.String `tfsdk:"folder"`
	Description types.String `tfsdk:"description"`
	Comment     types.String `tfsdk:"comment"`
	Disabled    types.Bool   `tfsdk:"disabled"`
	ValueRaw    types.String `tfsdk:"value_raw"`
}

func (d *RuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rule"
}

func (d *RuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a CheckMK rule by its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the rule (same as rule_id).",
				Computed:            true,
			},
			"rule_id": schema.StringAttribute{
				MarkdownDescription: "The rule ID to retrieve (UUID assigned by CheckMK).",
				Required:            true,
			},
			"ruleset": schema.StringAttribute{
				MarkdownDescription: "The ruleset this rule belongs to.",
				Computed:            true,
			},
			"folder": schema.StringAttribute{
				MarkdownDescription: "The folder where this rule is defined.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The rule description.",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Optional comment for the rule.",
				Computed:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the rule is disabled.",
				Computed:            true,
			},
			"value_raw": schema.StringAttribute{
				MarkdownDescription: "The raw rule value as a Python literal string.",
				Computed:            true,
			},
		},
	}
}

func (d *RuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(*common.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.ProviderData, got: %T", req.ProviderData),
		)
		return
	}
	d.providerData = providerData
}

func (d *RuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.providerData.Client.GetRule(ctx, data.RuleID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read rule: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	data.RuleID = types.StringValue(result.ID)
	data.Ruleset = types.StringValue(result.Extensions.Ruleset)
	data.Folder = types.StringValue(result.Extensions.Folder)
	data.Description = types.StringValue(result.Extensions.Properties.Description)
	data.Comment = types.StringValue(result.Extensions.Properties.Comment)
	data.Disabled = types.BoolValue(result.Extensions.Properties.Disabled)
	data.ValueRaw = types.StringValue(result.Extensions.ValueRaw)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// =============================================================================
// Notification Rule Data Source
// =============================================================================

var _ datasource.DataSource = &NotificationRuleDataSource{}

func NewNotificationRuleDataSource() datasource.DataSource {
	return &NotificationRuleDataSource{}
}

type NotificationRuleDataSource struct {
	providerData *common.ProviderData
}

type NotificationRuleDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	RuleID     types.String `tfsdk:"rule_id"`
	Title      types.String `tfsdk:"title"`
	RuleConfig types.String `tfsdk:"rule_config"`
}

func (d *NotificationRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_rule"
}

func (d *NotificationRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a CheckMK notification rule by its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the notification rule (same as rule_id).",
				Computed:            true,
			},
			"rule_id": schema.StringAttribute{
				MarkdownDescription: "The notification rule ID to retrieve (UUID assigned by CheckMK).",
				Required:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "The notification rule title/description.",
				Computed:            true,
			},
			"rule_config": schema.StringAttribute{
				MarkdownDescription: "The raw notification rule configuration as JSON.",
				Computed:            true,
			},
		},
	}
}

func (d *NotificationRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(*common.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.ProviderData, got: %T", req.ProviderData),
		)
		return
	}
	d.providerData = providerData
}

func (d *NotificationRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NotificationRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.providerData.Client.GetNotificationRule(ctx, data.RuleID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notification_rule: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	data.RuleID = types.StringValue(result.ID)
	data.Title = types.StringValue(result.Title)
	data.RuleConfig = types.StringValue(string(result.Extensions.RuleConfig))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
