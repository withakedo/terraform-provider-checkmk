// Package users provides data sources for reading CheckMK users and contact groups.
package users

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-provider-checkmk/internal/common"
)

// =============================================================================
// User Data Source
// =============================================================================

var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	providerData *common.ProviderData
}

type UserDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Username      types.String `tfsdk:"username"`
	Fullname      types.String `tfsdk:"fullname"`
	Email         types.String `tfsdk:"email"`
	ContactGroups types.List   `tfsdk:"contact_groups"`
	Roles         types.List   `tfsdk:"roles"`
	DisableLogin  types.Bool   `tfsdk:"disable_login"`
	AuthType      types.String `tfsdk:"auth_type"`
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a CheckMK user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the user (same as username).",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username to retrieve.",
				Required:            true,
			},
			"fullname": schema.StringAttribute{
				MarkdownDescription: "The full name of the user.",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address of the user.",
				Computed:            true,
			},
			"contact_groups": schema.ListAttribute{
				MarkdownDescription: "List of contact groups the user belongs to.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"roles": schema.ListAttribute{
				MarkdownDescription: "List of roles assigned to the user.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"disable_login": schema.BoolAttribute{
				MarkdownDescription: "Whether the user's login is disabled.",
				Computed:            true,
			},
			"auth_type": schema.StringAttribute{
				MarkdownDescription: "The authentication type (password, automation).",
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.providerData.Client.GetUser(ctx, data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	data.Username = types.StringValue(result.ID)
	data.Fullname = types.StringValue(result.Extensions.Fullname)
	data.Email = types.StringValue(result.Extensions.ContactOptions.Email)
	data.DisableLogin = types.BoolValue(result.Extensions.DisableLogin)
	data.AuthType = types.StringValue(result.Extensions.AuthOption.AuthType)

	contactGroups, diags := types.ListValueFrom(ctx, types.StringType, result.Extensions.ContactGroups)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ContactGroups = contactGroups

	roles, diags := types.ListValueFrom(ctx, types.StringType, result.Extensions.Roles)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Roles = roles

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// =============================================================================
// Contact Group Data Source
// =============================================================================

var _ datasource.DataSource = &ContactGroupDataSource{}

func NewContactGroupDataSource() datasource.DataSource {
	return &ContactGroupDataSource{}
}

type ContactGroupDataSource struct {
	providerData *common.ProviderData
}

type ContactGroupDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Alias types.String `tfsdk:"alias"`
}

func (d *ContactGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact_group"
}

func (d *ContactGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a CheckMK contact group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the contact group (same as name).",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the contact group to retrieve.",
				Required:            true,
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "Human-readable alias for the contact group.",
				Computed:            true,
			},
		},
	}
}

func (d *ContactGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ContactGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContactGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.providerData.Client.GetContactGroup(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read contact_group: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	data.Name = types.StringValue(result.ID)
	data.Alias = types.StringValue(result.Title)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// =============================================================================
// Password Data Source
// =============================================================================

var _ datasource.DataSource = &PasswordDataSource{}

func NewPasswordDataSource() datasource.DataSource {
	return &PasswordDataSource{}
}

type PasswordDataSource struct {
	providerData *common.ProviderData
}

type PasswordDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Title types.String `tfsdk:"title"`
	Owner types.String `tfsdk:"owner"`
}

func (d *PasswordDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password"
}

func (d *PasswordDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a CheckMK password store entry. " +
			"Note: The actual password value is never returned for security reasons.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the password.",
				Required:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Human-readable title for the password.",
				Computed:            true,
			},
			"owner": schema.StringAttribute{
				MarkdownDescription: "The owner of the password.",
				Computed:            true,
			},
		},
	}
}

func (d *PasswordDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PasswordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PasswordDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.providerData.Client.GetPassword(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read password: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	data.Title = types.StringValue(result.Title)
	data.Owner = types.StringValue(result.Extensions.Owner)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
