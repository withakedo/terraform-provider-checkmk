package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/withakedo/terraform-provider-checkmk/internal/client"
	"github.com/withakedo/terraform-provider-checkmk/internal/common"
)

var (
	_ resource.Resource                   = &DowntimeResource{}
	_ resource.ResourceWithValidateConfig = &DowntimeResource{}
)

func NewDowntimeResource() resource.Resource {
	return &DowntimeResource{}
}

// DowntimeResource schedules a CheckMK downtime (maintenance window) for a
// host, or for specific services on a host. Downtimes are monitoring
// operations rather than configuration: they take effect immediately and do
// not require - or affect - config activation.
//
// CheckMK's create-downtime endpoints respond 204 No Content with no
// server-generated id, so this resource identifies its downtime by the
// host/service parameters it was created with rather than by CheckMK's
// internal downtime id, and deletes by those same parameters. It also does
// not currently support hostgroup/servicegroup/query-based downtimes, or
// CheckMK's separate "modify downtime" endpoint - changing any attribute
// replaces the resource (cancels the old downtime, schedules a new one).
type DowntimeResource struct {
	providerData *common.ProviderData
}

// DowntimeResourceModel describes the resource data model.
type DowntimeResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	HostName            types.String `tfsdk:"host_name"`
	ServiceDescriptions types.List   `tfsdk:"service_descriptions"`
	StartTime           types.String `tfsdk:"start_time"`
	EndTime             types.String `tfsdk:"end_time"`
	Comment             types.String `tfsdk:"comment"`
	Duration            types.Int64  `tfsdk:"duration"`
	Recur               types.String `tfsdk:"recur"`
}

func (r *DowntimeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_downtime"
}

func (r *DowntimeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Schedules a CheckMK downtime (maintenance window) for a host, or for
specific services on a host.

Downtimes are monitoring operations, not configuration - they take effect immediately and do not
require ` + "`checkmk_activation`" + `. Every attribute forces replacement on change: there is no
in-place "modify downtime" support, so changing e.g. ` + "`end_time`" + ` cancels the existing
downtime and schedules a new one.
`,
		Attributes: map[string]schema.Attribute{
			"id": common.ComputedIDAttribute("Unique identifier for this downtime resource instance."),
			"host_name": schema.StringAttribute{
				MarkdownDescription: "Name of the host to schedule the downtime for.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_descriptions": schema.ListAttribute{
				MarkdownDescription: "Names of services on `host_name` to schedule the downtime for. " +
					"If omitted, the downtime applies to the whole host.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"start_time": schema.StringAttribute{
				MarkdownDescription: "Start of the downtime, as an ISO 8601 / RFC3339 timestamp (e.g. `2026-09-01T22:00:00Z`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"end_time": schema.StringAttribute{
				MarkdownDescription: "End of the downtime, as an ISO 8601 / RFC3339 timestamp.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment explaining the reason for the downtime.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"duration": schema.Int64Attribute{
				MarkdownDescription: "For flexible (non-`fixed`) downtimes, the actual downtime " +
					"duration in seconds once triggered by the first problem within the scheduled window.",
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"recur": schema.StringAttribute{
				MarkdownDescription: "Recurrence of the downtime (default: `fixed`). One of: `fixed`, " +
					"`hour`, `day`, `week`, `second_week`, `fourth_week`, `weekday_start`, " +
					"`weekday_end`, `day_of_month`. Recurring downtimes other than `fixed` require a " +
					"CheckMK commercial edition.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("fixed"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *DowntimeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.providerData = common.ConfigureResource(req, resp)
}

func (r *DowntimeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DowntimeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostName := data.HostName.ValueString()

	var serviceDescriptions []string
	if !data.ServiceDescriptions.IsNull() {
		resp.Diagnostics.Append(data.ServiceDescriptions.ElementsAs(ctx, &serviceDescriptions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(serviceDescriptions) > 0 {
		req := &client.CreateServiceDowntimeRequest{
			DowntimeType:        "service",
			HostName:            hostName,
			ServiceDescriptions: serviceDescriptions,
			StartTime:           data.StartTime.ValueString(),
			EndTime:             data.EndTime.ValueString(),
			Comment:             data.Comment.ValueString(),
			Duration:            data.Duration.ValueInt64(),
			Recur:               data.Recur.ValueString(),
		}
		tflog.Info(ctx, "Scheduling CheckMK service downtime", map[string]interface{}{
			"host_name":            hostName,
			"service_descriptions": serviceDescriptions,
		})
		if err := r.providerData.Client.CreateServiceDowntime(ctx, req); err != nil {
			common.AddClientError(resp, "schedule", "service downtime", err)
			return
		}
	} else {
		req := &client.CreateHostDowntimeRequest{
			DowntimeType: "host",
			HostName:     hostName,
			StartTime:    data.StartTime.ValueString(),
			EndTime:      data.EndTime.ValueString(),
			Comment:      data.Comment.ValueString(),
			Duration:     data.Duration.ValueInt64(),
			Recur:        data.Recur.ValueString(),
		}
		tflog.Info(ctx, "Scheduling CheckMK host downtime", map[string]interface{}{
			"host_name": hostName,
		})
		if err := r.providerData.Client.CreateHostDowntime(ctx, req); err != nil {
			common.AddClientError(resp, "schedule", "host downtime", err)
			return
		}
	}

	data.ID = types.StringValue(fmt.Sprintf("downtime-%s-%s", hostName, data.StartTime.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DowntimeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DowntimeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// CheckMK does not return a lookup-friendly id when a downtime is
	// created, so this resource does not attempt to re-verify the downtime
	// still exists on read; it will naturally expire in CheckMK once
	// end_time passes regardless of Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DowntimeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Every attribute is RequiresReplace, so Terraform handles changes via
	// destroy+create; this should be unreachable in practice.
	var data DowntimeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DowntimeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DowntimeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostName := data.HostName.ValueString()

	var serviceDescriptions []string
	if !data.ServiceDescriptions.IsNull() {
		resp.Diagnostics.Append(data.ServiceDescriptions.ElementsAs(ctx, &serviceDescriptions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Info(ctx, "Deleting CheckMK downtime", map[string]interface{}{
		"host_name":            hostName,
		"service_descriptions": serviceDescriptions,
	})

	if err := r.providerData.Client.DeleteDowntimeByParams(ctx, hostName, serviceDescriptions); err != nil {
		common.AddClientError(resp, "delete", "downtime", err)
		return
	}
}

// ValidateConfig validates the resource configuration against the generated
// OpenAPI types for the connected CheckMK version.
func (r *DowntimeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.providerData == nil {
		return
	}

	var data DowntimeResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaName := "CreateHostDowntime"
	if !data.ServiceDescriptions.IsNull() {
		var serviceDescriptions []string
		diags := data.ServiceDescriptions.ElementsAs(ctx, &serviceDescriptions, false)
		resp.Diagnostics.Append(diags...)
		if len(serviceDescriptions) > 0 {
			schemaName = "CreateServiceDowntime"
		}
	}

	var diags diag.Diagnostics
	validator := common.NewAttributeValidator(r.providerData)
	diags.Append(validator.ValidateStringField(schemaName, "recur", data.Recur, path.Root("recur"))...)
	resp.Diagnostics.Append(diags...)
}
