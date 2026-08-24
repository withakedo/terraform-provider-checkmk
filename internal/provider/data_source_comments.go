package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/withakedo/terraform-provider-checkmk/internal/common"
)

var _ datasource.DataSource = &CommentsDataSource{}

func NewCommentsDataSource() datasource.DataSource {
	return &CommentsDataSource{}
}

// CommentsDataSource lists the comments currently present on a host (both
// host-level and service-level comments), so out-of-band or
// externally-added comments can be observed from Terraform -
// checkmk_comment itself has no persistent server-side object to read back
// (see its docs), so this is the way to detect that kind of drift.
type CommentsDataSource struct {
	providerData *common.ProviderData
}

// CommentsDataSourceModel describes the data source's data model.
type CommentsDataSourceModel struct {
	ID       types.String        `tfsdk:"id"`
	HostName types.String        `tfsdk:"host_name"`
	Comments []CommentEntryModel `tfsdk:"comments"`
}

// CommentEntryModel describes a single comment within the list.
type CommentEntryModel struct {
	ID                 types.Int64  `tfsdk:"id"`
	SiteID             types.String `tfsdk:"site_id"`
	Author             types.String `tfsdk:"author"`
	Comment            types.String `tfsdk:"comment"`
	Persistent         types.Bool   `tfsdk:"persistent"`
	EntryTime          types.String `tfsdk:"entry_time"`
	IsService          types.Bool   `tfsdk:"is_service"`
	ServiceDescription types.String `tfsdk:"service_description"`
}

func (d *CommentsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_comments"
}

func (d *CommentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists comments currently present on a host, covering both host-level and " +
			"service-level comments. Useful for detecting comments added outside Terraform (in the CheckMK " +
			"UI, by another automation, or by `checkmk_comment` resources managed elsewhere).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Same as `host_name`.",
				Computed:            true,
			},
			"host_name": schema.StringAttribute{
				MarkdownDescription: "The host to list comments for.",
				Required:            true,
			},
			"comments": schema.ListNestedAttribute{
				MarkdownDescription: "The comments found for this host.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                  schema.Int64Attribute{Computed: true, MarkdownDescription: "CheckMK's internal id for the comment."},
						"site_id":             schema.StringAttribute{Computed: true, MarkdownDescription: "The site id of the comment."},
						"author":              schema.StringAttribute{Computed: true, MarkdownDescription: "Who wrote the comment."},
						"comment":             schema.StringAttribute{Computed: true, MarkdownDescription: "The comment text."},
						"persistent":          schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the comment persists a CheckMK restart."},
						"entry_time":          schema.StringAttribute{Computed: true, MarkdownDescription: "When the comment was created."},
						"is_service":          schema.BoolAttribute{Computed: true, MarkdownDescription: "True if this is a service comment, false if it's a host comment."},
						"service_description": schema.StringAttribute{Computed: true, MarkdownDescription: "The service name, if `is_service` is true."},
					},
				},
			},
		},
	}
}

func (d *CommentsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CommentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CommentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostName := data.HostName.ValueString()
	infos, err := d.providerData.Client.ListComments(ctx, hostName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list comments: %s", err))
		return
	}

	entries := make([]CommentEntryModel, len(infos))
	for i, info := range infos {
		entries[i] = CommentEntryModel{
			ID:                 types.Int64Value(info.ID),
			SiteID:             types.StringValue(info.SiteID),
			Author:             types.StringValue(info.Author),
			Comment:            types.StringValue(info.Comment),
			Persistent:         types.BoolValue(info.Persistent),
			EntryTime:          types.StringValue(info.EntryTime),
			IsService:          types.BoolValue(info.IsService),
			ServiceDescription: types.StringValue(info.ServiceDescription),
		}
	}

	data.ID = types.StringValue(hostName)
	data.Comments = entries

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
