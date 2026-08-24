// Package groups provides data sources for reading host and service groups.
package groups

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/withakedo/terraform_checkmk_provider/internal/client"
	"github.com/withakedo/terraform_checkmk_provider/internal/common"
)

// GroupDataSourceModel is the shared data model for all group data sources.
type GroupDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Alias types.String `tfsdk:"alias"`
}

// GroupDataSourceConfig configures a group data source type.
type GroupDataSourceConfig struct {
	// TypeName is the Terraform data source type suffix (e.g., "host_group", "service_group")
	TypeName string
	// Description is the data source's markdown description
	Description string
	// IDDescription is the description for the id attribute
	IDDescription string
	// NameDescription is the description for the name attribute
	NameDescription string
	// AliasDescription is the description for the alias attribute
	AliasDescription string

	// Client operations
	Get func(ctx context.Context, c *client.Client, name string) (id, title string, err error)
}

// BaseGroupDataSource is the generic group data source implementation.
type BaseGroupDataSource struct {
	config       GroupDataSourceConfig
	providerData *common.ProviderData
}

// NewBaseGroupDataSource creates a new group data source with the given configuration.
func NewBaseGroupDataSource(config GroupDataSourceConfig) datasource.DataSource {
	return &BaseGroupDataSource{config: config}
}

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &BaseGroupDataSource{}

func (d *BaseGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.TypeName
}

func (d *BaseGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: d.config.Description,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: d.config.IDDescription,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: d.config.NameDescription,
				Required:            true,
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: d.config.AliasDescription,
				Computed:            true,
			},
		},
	}
}

func (d *BaseGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BaseGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, title, err := d.config.Get(ctx, d.providerData.Client, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read %s: %s", d.config.TypeName, err))
		return
	}

	data.ID = types.StringValue(id)
	data.Name = types.StringValue(id)
	data.Alias = types.StringValue(title)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// =============================================================================
// Pre-configured Group Data Sources
// =============================================================================

// HostGroupDataSourceConfig returns the configuration for host group data sources.
var HostGroupDataSourceConfig = GroupDataSourceConfig{
	TypeName:         "host_group",
	Description:      "Use this data source to retrieve information about a CheckMK host group.",
	IDDescription:    "The unique identifier for the host group (same as name).",
	NameDescription:  "The name of the host group to retrieve.",
	AliasDescription: "Human-readable alias/title for the host group.",

	Get: func(ctx context.Context, c *client.Client, name string) (string, string, error) {
		result, err := c.GetHostGroup(ctx, name)
		if err != nil {
			return "", "", err
		}
		return result.ID, result.Title, nil
	},
}

// ServiceGroupDataSourceConfig returns the configuration for service group data sources.
var ServiceGroupDataSourceConfig = GroupDataSourceConfig{
	TypeName:         "service_group",
	Description:      "Use this data source to retrieve information about a CheckMK service group.",
	IDDescription:    "The unique identifier for the service group (same as name).",
	NameDescription:  "The name of the service group to retrieve.",
	AliasDescription: "Human-readable alias/title for the service group.",

	Get: func(ctx context.Context, c *client.Client, name string) (string, string, error) {
		result, err := c.GetServiceGroup(ctx, name)
		if err != nil {
			return "", "", err
		}
		return result.ID, result.Title, nil
	},
}

// NewHostGroupDataSource creates a new host group data source.
func NewHostGroupDataSource() datasource.DataSource {
	return NewBaseGroupDataSource(HostGroupDataSourceConfig)
}

// NewServiceGroupDataSource creates a new service group data source.
func NewServiceGroupDataSource() datasource.DataSource {
	return NewBaseGroupDataSource(ServiceGroupDataSourceConfig)
}
