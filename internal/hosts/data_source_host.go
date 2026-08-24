// Package hosts provides data sources for reading CheckMK hosts.
package hosts

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/withakedo/terraform-provider-checkmk/internal/common"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &HostDataSource{}

// NewHostDataSource creates a new host data source.
func NewHostDataSource() datasource.DataSource {
	return &HostDataSource{}
}

// HostDataSource defines the data source implementation.
type HostDataSource struct {
	providerData *common.ProviderData
}

// HostDataSourceModel describes the data source data model.
type HostDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	HostName   types.String `tfsdk:"host_name"`
	Folder     types.String `tfsdk:"folder"`
	Attributes types.Map    `tfsdk:"attributes"`
}

func (d *HostDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (d *HostDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a CheckMK host.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the host (same as host_name).",
				Computed:            true,
			},
			"host_name": schema.StringAttribute{
				MarkdownDescription: "The hostname to retrieve.",
				Required:            true,
			},
			"folder": schema.StringAttribute{
				MarkdownDescription: "The folder path where the host resides.",
				Computed:            true,
			},
			"attributes": schema.MapAttribute{
				MarkdownDescription: "Host attributes (alias, ipaddress, tags, etc.).",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *HostDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *HostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HostDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.providerData.Client.GetHost(ctx, data.HostName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read host: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	data.HostName = types.StringValue(result.ID)
	data.Folder = types.StringValue(result.Extensions.Folder)

	// Convert attributes to map
	attrMap := make(map[string]string)
	for k, v := range result.Extensions.Attributes {
		if strVal, ok := v.(string); ok {
			attrMap[k] = strVal
		} else {
			attrMap[k] = fmt.Sprintf("%v", v)
		}
	}
	attrs, diags := types.MapValueFrom(ctx, types.StringType, attrMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Attributes = attrs

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
