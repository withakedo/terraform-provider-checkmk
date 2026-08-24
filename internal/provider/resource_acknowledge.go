package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/withakedo/terraform-provider-checkmk/internal/client"
	"github.com/withakedo/terraform-provider-checkmk/internal/common"
)

var _ resource.Resource = &AcknowledgeResource{}

func NewAcknowledgeResource() resource.Resource {
	return &AcknowledgeResource{}
}

// AcknowledgeResource acknowledges the current problem state of a host, or
// of a single service on a host. Like checkmk_downtime, this is a monitoring
// operation rather than configuration: it takes effect immediately and does
// not require - or affect - config activation.
//
// CheckMK's acknowledge endpoints respond 204 No Content with no
// server-generated id, so this resource identifies its acknowledgement by
// the host/service parameters it was created with rather than by a CheckMK
// internal id, and removes the acknowledgement by those same parameters on
// delete. It does not currently support hostgroup/servicegroup/query-based
// acknowledgements.
type AcknowledgeResource struct {
	providerData *common.ProviderData
}

// AcknowledgeResourceModel describes the resource data model.
type AcknowledgeResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	HostName           types.String `tfsdk:"host_name"`
	ServiceDescription types.String `tfsdk:"service_description"`
	Comment            types.String `tfsdk:"comment"`
	Sticky             types.Bool   `tfsdk:"sticky"`
	Persistent         types.Bool   `tfsdk:"persistent"`
	Notify             types.Bool   `tfsdk:"notify"`
	ExpireOn           types.String `tfsdk:"expire_on"`
}

func (r *AcknowledgeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acknowledge"
}

func (r *AcknowledgeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Acknowledges the current problem state of a host, or of a single service on a host.

Acknowledgements are monitoring operations, not configuration - they take effect immediately and
do not require ` + "`checkmk_activation`" + `. Every attribute forces replacement on change: there is
no in-place "modify acknowledgement" support, so changing e.g. ` + "`comment`" + ` removes the
existing acknowledgement and creates a new one.
`,
		Attributes: map[string]schema.Attribute{
			"id": common.ComputedIDAttribute("Unique identifier for this acknowledgement resource instance."),
			"host_name": schema.StringAttribute{
				MarkdownDescription: "Name of the host to acknowledge the problem for.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_description": schema.StringAttribute{
				MarkdownDescription: "Name of the service on `host_name` to acknowledge the problem for. " +
					"If omitted, the acknowledgement applies to the host itself.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment explaining the acknowledgement.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sticky": schema.BoolAttribute{
				MarkdownDescription: "If true (default), only a state-change to UP/OK discards the " +
					"acknowledgement. If false, the acknowledgement is discarded on any state-change.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"persistent": schema.BoolAttribute{
				MarkdownDescription: "If true, the comment persists a CheckMK restart (default: false).",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"notify": schema.BoolAttribute{
				MarkdownDescription: "If true (default), notifications are sent out to the configured contacts.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"expire_on": schema.StringAttribute{
				MarkdownDescription: "If set, the acknowledgement expires at this time, as an ISO 8601 " +
					"/ RFC3339 timestamp (e.g. `2026-09-01T22:00:00Z`). Timezone defaults to UTC.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *AcknowledgeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.providerData = common.ConfigureResource(req, resp)
}

func (r *AcknowledgeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AcknowledgeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostName := data.HostName.ValueString()
	serviceDescription := data.ServiceDescription.ValueString()

	if serviceDescription != "" {
		req := &client.AcknowledgeServiceProblemRequest{
			AcknowledgeType:    "service",
			HostName:           hostName,
			ServiceDescription: serviceDescription,
			Comment:            data.Comment.ValueString(),
			Sticky:             data.Sticky.ValueBool(),
			Persistent:         data.Persistent.ValueBool(),
			Notify:             data.Notify.ValueBool(),
			ExpireOn:           data.ExpireOn.ValueString(),
		}
		tflog.Info(ctx, "Acknowledging CheckMK service problem", map[string]interface{}{
			"host_name":           hostName,
			"service_description": serviceDescription,
		})
		if err := r.providerData.Client.AcknowledgeServiceProblem(ctx, req); err != nil {
			common.AddClientError(resp, "acknowledge", "service problem", err)
			return
		}
	} else {
		req := &client.AcknowledgeHostProblemRequest{
			AcknowledgeType: "host",
			HostName:        hostName,
			Comment:         data.Comment.ValueString(),
			Sticky:          data.Sticky.ValueBool(),
			Persistent:      data.Persistent.ValueBool(),
			Notify:          data.Notify.ValueBool(),
			ExpireOn:        data.ExpireOn.ValueString(),
		}
		tflog.Info(ctx, "Acknowledging CheckMK host problem", map[string]interface{}{
			"host_name": hostName,
		})
		if err := r.providerData.Client.AcknowledgeHostProblem(ctx, req); err != nil {
			common.AddClientError(resp, "acknowledge", "host problem", err)
			return
		}
	}

	data.ID = types.StringValue(fmt.Sprintf("acknowledge-%s-%s", hostName, serviceDescription))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AcknowledgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AcknowledgeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// CheckMK does not return a lookup-friendly id when an acknowledgement is
	// created, so this resource does not attempt to re-verify the
	// acknowledgement still exists on read; it will naturally be discarded by
	// CheckMK on state-change (or expire_on) regardless of Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AcknowledgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Every attribute is RequiresReplace, so Terraform handles changes via
	// destroy+create; this should be unreachable in practice.
	var data AcknowledgeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AcknowledgeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AcknowledgeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostName := data.HostName.ValueString()
	serviceDescription := data.ServiceDescription.ValueString()

	tflog.Info(ctx, "Removing CheckMK acknowledgement", map[string]interface{}{
		"host_name":           hostName,
		"service_description": serviceDescription,
	})

	if err := r.providerData.Client.RemoveAcknowledgement(ctx, hostName, serviceDescription); err != nil {
		common.AddClientError(resp, "remove", "acknowledgement", err)
		return
	}
}
