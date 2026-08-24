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

var _ resource.Resource = &CommentResource{}

func NewCommentResource() resource.Resource {
	return &CommentResource{}
}

// CommentResource adds a comment to a host, or to a single service on a
// host. Like checkmk_downtime and checkmk_acknowledge, this is a monitoring
// operation rather than configuration: it takes effect immediately and does
// not require - or affect - config activation.
//
// CheckMK's create-comment endpoints respond 204 No Content with no
// server-generated id, so this resource identifies its comment by the
// host/service parameters it was created with rather than by CheckMK's
// internal comment id, and deletes by those same parameters.
type CommentResource struct {
	providerData *common.ProviderData
}

// CommentResourceModel describes the resource data model.
type CommentResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	HostName           types.String `tfsdk:"host_name"`
	ServiceDescription types.String `tfsdk:"service_description"`
	Comment            types.String `tfsdk:"comment"`
	Persistent         types.Bool   `tfsdk:"persistent"`
}

func (r *CommentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_comment"
}

func (r *CommentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Adds a comment to a host, or to a single service on a host.

Comments are monitoring operations, not configuration - they take effect immediately and do not
require ` + "`checkmk_activation`" + `. Every attribute forces replacement on change: there is no
in-place "modify comment" support, so changing the comment text removes the existing comment and
creates a new one.
`,
		Attributes: map[string]schema.Attribute{
			"id": common.ComputedIDAttribute("Unique identifier for this comment resource instance."),
			"host_name": schema.StringAttribute{
				MarkdownDescription: "Name of the host to add the comment to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_description": schema.StringAttribute{
				MarkdownDescription: "Name of the service on `host_name` to add the comment to. " +
					"If omitted, the comment applies to the host itself.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "The comment text.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
		},
	}
}

func (r *CommentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.providerData = common.ConfigureResource(req, resp)
}

func (r *CommentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CommentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostName := data.HostName.ValueString()
	serviceDescription := data.ServiceDescription.ValueString()

	if serviceDescription != "" {
		req := &client.CreateServiceCommentRequest{
			CommentType:        "service",
			HostName:           hostName,
			ServiceDescription: serviceDescription,
			Comment:            data.Comment.ValueString(),
			Persistent:         data.Persistent.ValueBool(),
		}
		tflog.Info(ctx, "Adding CheckMK service comment", map[string]interface{}{
			"host_name":           hostName,
			"service_description": serviceDescription,
		})
		if err := r.providerData.Client.CreateServiceComment(ctx, req); err != nil {
			common.AddClientError(resp, "create", "service comment", err)
			return
		}
	} else {
		req := &client.CreateHostCommentRequest{
			CommentType: "host",
			HostName:    hostName,
			Comment:     data.Comment.ValueString(),
			Persistent:  data.Persistent.ValueBool(),
		}
		tflog.Info(ctx, "Adding CheckMK host comment", map[string]interface{}{
			"host_name": hostName,
		})
		if err := r.providerData.Client.CreateHostComment(ctx, req); err != nil {
			common.AddClientError(resp, "create", "host comment", err)
			return
		}
	}

	data.ID = types.StringValue(fmt.Sprintf("comment-%s-%s", hostName, serviceDescription))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CommentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CommentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// CheckMK does not return a lookup-friendly id when a comment is created,
	// so this resource does not attempt to re-verify the comment still
	// exists on read.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CommentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Every attribute is RequiresReplace, so Terraform handles changes via
	// destroy+create; this should be unreachable in practice.
	var data CommentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CommentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CommentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostName := data.HostName.ValueString()

	var serviceDescriptions []string
	if sd := data.ServiceDescription.ValueString(); sd != "" {
		serviceDescriptions = []string{sd}
	}

	tflog.Info(ctx, "Deleting CheckMK comment", map[string]interface{}{
		"host_name":            hostName,
		"service_descriptions": serviceDescriptions,
	})

	if err := r.providerData.Client.DeleteCommentsByParams(ctx, hostName, serviceDescriptions); err != nil {
		common.AddClientError(resp, "delete", "comment", err)
		return
	}
}
