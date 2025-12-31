// Package configuration provides data sources for reading CheckMK configuration objects.
package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-provider-checkmk/internal/common"
)

// =============================================================================
// Aux Tag Data Source
// =============================================================================

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AuxTagDataSource{}

// NewAuxTagDataSource creates a new aux_tag data source.
func NewAuxTagDataSource() datasource.DataSource {
	return &AuxTagDataSource{}
}

// AuxTagDataSource defines the data source implementation.
type AuxTagDataSource struct {
	providerData *common.ProviderData
}

// AuxTagDataSourceModel describes the data source data model.
type AuxTagDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Title types.String `tfsdk:"title"`
	Topic types.String `tfsdk:"topic"`
	Help  types.String `tfsdk:"help"`
}

func (d *AuxTagDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aux_tag"
}

func (d *AuxTagDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a CheckMK auxiliary tag.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the auxiliary tag.",
				Required:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Human-readable title for the auxiliary tag.",
				Computed:            true,
			},
			"topic": schema.StringAttribute{
				MarkdownDescription: "Topic for grouping auxiliary tags in the UI.",
				Computed:            true,
			},
			"help": schema.StringAttribute{
				MarkdownDescription: "Help text describing the auxiliary tag.",
				Computed:            true,
			},
		},
	}
}

func (d *AuxTagDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AuxTagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AuxTagDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.providerData.Client.GetAuxTag(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read aux_tag: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	data.Title = types.StringValue(result.Title)
	data.Topic = types.StringValue(result.Extensions.Topic)
	data.Help = types.StringValue(result.Extensions.Help)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// =============================================================================
// Folder Data Source
// =============================================================================

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &FolderDataSource{}

// NewFolderDataSource creates a new folder data source.
func NewFolderDataSource() datasource.DataSource {
	return &FolderDataSource{}
}

// FolderDataSource defines the data source implementation.
type FolderDataSource struct {
	providerData *common.ProviderData
}

// FolderDataSourceModel describes the data source data model.
type FolderDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Path       types.String `tfsdk:"path"`
	Title      types.String `tfsdk:"title"`
	Attributes types.Map    `tfsdk:"attributes"`
}

func (d *FolderDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folder"
}

func (d *FolderDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a CheckMK folder.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the folder (same as path).",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The path to the folder (e.g., '/', '/production', '/production/web').",
				Required:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Human-readable title for the folder.",
				Computed:            true,
			},
			"attributes": schema.MapAttribute{
				MarkdownDescription: "Folder attributes that are inherited by hosts in this folder.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *FolderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FolderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FolderDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.providerData.Client.GetFolder(ctx, data.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read folder: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	data.Path = types.StringValue(result.ID)
	data.Title = types.StringValue(result.Title)

	// Convert attributes to map
	attrMap := make(map[string]string)
	for k, v := range result.Extensions.Attributes {
		if strVal, ok := v.(string); ok {
			attrMap[k] = strVal
		} else {
			// Convert non-string values to string representation
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

// =============================================================================
// Tag Group Data Source
// =============================================================================

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &TagGroupDataSource{}

// NewTagGroupDataSource creates a new tag_group data source.
func NewTagGroupDataSource() datasource.DataSource {
	return &TagGroupDataSource{}
}

// TagGroupDataSource defines the data source implementation.
type TagGroupDataSource struct {
	providerData *common.ProviderData
}

// TagGroupDataSourceModel describes the data source data model.
type TagGroupDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Title types.String `tfsdk:"title"`
	Topic types.String `tfsdk:"topic"`
	Help  types.String `tfsdk:"help"`
	Tags  types.List   `tfsdk:"tags"`
}

// TagDataSourceModel describes a tag in the tag group.
type TagDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Title   types.String `tfsdk:"title"`
	AuxTags types.List   `tfsdk:"aux_tags"`
}

func (d *TagGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag_group"
}

func (d *TagGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a CheckMK tag group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the tag group.",
				Required:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Human-readable title for the tag group.",
				Computed:            true,
			},
			"topic": schema.StringAttribute{
				MarkdownDescription: "Topic for grouping tag groups in the UI.",
				Computed:            true,
			},
			"help": schema.StringAttribute{
				MarkdownDescription: "Help text describing the tag group.",
				Computed:            true,
			},
			"tags": schema.ListNestedAttribute{
				MarkdownDescription: "List of tags in the tag group.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier for the tag.",
							Computed:            true,
						},
						"title": schema.StringAttribute{
							MarkdownDescription: "Human-readable title for the tag.",
							Computed:            true,
						},
						"aux_tags": schema.ListAttribute{
							MarkdownDescription: "List of auxiliary tags associated with this tag.",
							ElementType:         types.StringType,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *TagGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TagGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TagGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.providerData.Client.GetTagGroup(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tag_group: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	data.Title = types.StringValue(result.Title)
	data.Topic = types.StringValue(result.Extensions.Topic)
	data.Help = types.StringValue(result.Extensions.Help)

	// Convert tags to list
	var tags []TagDataSourceModel
	for _, tag := range result.Extensions.Tags {
		auxTags, diags := types.ListValueFrom(ctx, types.StringType, tag.AuxTags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		tags = append(tags, TagDataSourceModel{
			ID:      types.StringValue(tag.ID),
			Title:   types.StringValue(tag.Title),
			AuxTags: auxTags,
		})
	}

	tagsList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":       types.StringType,
			"title":    types.StringType,
			"aux_tags": types.ListType{ElemType: types.StringType},
		},
	}, tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Tags = tagsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// =============================================================================
// Time Period Data Source
// =============================================================================

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &TimePeriodDataSource{}

// NewTimePeriodDataSource creates a new time_period data source.
func NewTimePeriodDataSource() datasource.DataSource {
	return &TimePeriodDataSource{}
}

// TimePeriodDataSource defines the data source implementation.
type TimePeriodDataSource struct {
	providerData *common.ProviderData
}

// TimePeriodDataSourceModel describes the data source data model.
type TimePeriodDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Alias types.String `tfsdk:"alias"`
}

func (d *TimePeriodDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_time_period"
}

func (d *TimePeriodDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a CheckMK time period.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the time period (same as name).",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the time period to retrieve.",
				Required:            true,
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "Human-readable alias for the time period.",
				Computed:            true,
			},
		},
	}
}

func (d *TimePeriodDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TimePeriodDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TimePeriodDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.providerData.Client.GetTimePeriod(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read time_period: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	data.Name = types.StringValue(result.ID)
	data.Alias = types.StringValue(result.Extensions.Alias)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
