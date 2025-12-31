// Package groups provides simple group resources (host_group, service_group).
// This file contains the generic base implementation for group resources.
package groups

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-provider-checkmk/internal/client"
	"github.com/terraform-provider-checkmk/internal/common"
)

// GroupResourceModel is the shared data model for all group resources.
type GroupResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Alias                 types.String `tfsdk:"alias"`
	StrictResourceLocking types.Bool   `tfsdk:"strict_resource_locking"`
}

// GroupResponse is a simple interface for group API responses.
type GroupResponse interface {
	GetID() string
	GetTitle() string
}

// GroupResourceConfig configures a group resource type.
type GroupResourceConfig struct {
	// TypeName is the Terraform resource type suffix (e.g., "host_group", "service_group")
	TypeName string
	// Description is the resource's markdown description
	Description string
	// IDDescription is the description for the id attribute
	IDDescription string
	// NameDescription is the description for the name attribute
	NameDescription string
	// AliasDescription is the description for the alias attribute
	AliasDescription string

	// Client operations
	Create      func(ctx context.Context, c *client.Client, name, alias string) (id, title string, err error)
	Get         func(ctx context.Context, c *client.Client, name string) (id, title string, err error)
	GetWithETag func(ctx context.Context, c *client.Client, name string) (id, title, etag string, err error)
	Update      func(ctx context.Context, c *client.Client, name, alias, etag string) (id, title string, err error)
	Delete      func(ctx context.Context, c *client.Client, name, etag string) error
}

// BaseGroupResource is the generic group resource implementation.
type BaseGroupResource struct {
	config       GroupResourceConfig
	providerData *common.ProviderData
}

// NewBaseGroupResource creates a new group resource with the given configuration.
func NewBaseGroupResource(config GroupResourceConfig) resource.Resource {
	return &BaseGroupResource{config: config}
}

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &BaseGroupResource{}
var _ resource.ResourceWithImportState = &BaseGroupResource{}

func (r *BaseGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.TypeName
}

func (r *BaseGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: r.config.Description,
		Attributes: map[string]schema.Attribute{
			"id":   common.ComputedIDAttribute(r.config.IDDescription),
			"name": common.RequiredIDAttribute(r.config.NameDescription),
			"alias": schema.StringAttribute{
				MarkdownDescription: r.config.AliasDescription,
				Required:            true,
			},
			"strict_resource_locking": common.StrictResourceLockingAttribute(),
		},
	}
}

func (r *BaseGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.providerData = common.ConfigureResource(req, resp)
}

func (r *BaseGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, title, err := r.config.Create(ctx, r.providerData.Client, data.Name.ValueString(), data.Alias.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create %s: %s", r.config.TypeName, err))
		return
	}

	data.ID = types.StringValue(id)
	data.Name = types.StringValue(id)
	data.Alias = types.StringValue(title)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaseGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, title, err := r.config.Get(ctx, r.providerData.Client, data.Name.ValueString())
	if err != nil {
		if common.HandleNotFoundOnRead(err, resp) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read %s: %s", r.config.TypeName, err))
		return
	}

	data.ID = types.StringValue(id)
	data.Name = types.StringValue(id)
	data.Alias = types.StringValue(title)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaseGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildSimpleBaseConfig(r.providerData, data.StrictResourceLocking)

	etag, err := common.FetchETagIfStrict(ctx, cfg.StrictResourceLocking, func(ctx context.Context) (string, error) {
		_, _, etag, err := r.config.GetWithETag(ctx, r.providerData.Client, data.Name.ValueString())
		return etag, err
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch %s ETag: %s", r.config.TypeName, err))
		return
	}

	id, title, err := r.config.Update(ctx, r.providerData.Client, data.Name.ValueString(), data.Alias.ValueString(), etag)
	if err != nil {
		if !common.HandleDriftWarning(err, resp, data.Name.ValueString()) {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update %s: %s", r.config.TypeName, err))
			return
		}
	}

	if id != "" {
		data.ID = types.StringValue(id)
		data.Name = types.StringValue(id)
		data.Alias = types.StringValue(title)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaseGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildSimpleBaseConfig(r.providerData, data.StrictResourceLocking)

	etag, err := common.FetchETagIfStrict(ctx, cfg.StrictResourceLocking, func(ctx context.Context) (string, error) {
		_, _, etag, err := r.config.GetWithETag(ctx, r.providerData.Client, data.Name.ValueString())
		if err != nil {
			if apiErr, ok := err.(*client.APIError); ok && apiErr.Status == 404 {
				return "", nil
			}
			return "", err
		}
		return etag, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch %s ETag: %s", r.config.TypeName, err))
		return
	}

	if err := r.config.Delete(ctx, r.providerData.Client, data.Name.ValueString(), etag); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete %s: %s", r.config.TypeName, err))
		return
	}
}

func (r *BaseGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// =============================================================================
// Pre-configured Group Resources
// =============================================================================

// HostGroupConfig returns the configuration for host group resources.
var HostGroupConfig = GroupResourceConfig{
	TypeName:         "host_group",
	Description:      "Manages a CheckMK host group. Host groups provide logical grouping of hosts independent of folder hierarchy. Does not require activation.",
	IDDescription:    "The unique identifier for the host group (same as name).",
	NameDescription:  "Unique name for the host group. Cannot be changed after creation.",
	AliasDescription: "Human-readable alias/title for the host group.",

	Create: func(ctx context.Context, c *client.Client, name, alias string) (string, string, error) {
		result, err := c.CreateHostGroup(ctx, &client.HostGroupCreateRequest{Name: name, Alias: alias})
		if err != nil {
			return "", "", err
		}
		return result.ID, result.Title, nil
	},
	Get: func(ctx context.Context, c *client.Client, name string) (string, string, error) {
		result, err := c.GetHostGroup(ctx, name)
		if err != nil {
			return "", "", err
		}
		return result.ID, result.Title, nil
	},
	GetWithETag: func(ctx context.Context, c *client.Client, name string) (string, string, string, error) {
		result, err := c.GetHostGroupWithETag(ctx, name)
		if err != nil {
			return "", "", "", err
		}
		return result.HostGroup.ID, result.HostGroup.Title, result.ETag, nil
	},
	Update: func(ctx context.Context, c *client.Client, name, alias, etag string) (string, string, error) {
		result, err := c.UpdateHostGroup(ctx, name, &client.HostGroupUpdateRequest{Alias: alias}, etag)
		if err != nil {
			return "", "", err
		}
		return result.ID, result.Title, nil
	},
	Delete: func(ctx context.Context, c *client.Client, name, etag string) error {
		return c.DeleteHostGroup(ctx, name, etag)
	},
}

// ServiceGroupConfig returns the configuration for service group resources.
var ServiceGroupConfig = GroupResourceConfig{
	TypeName:         "service_group",
	Description:      "Manages a CheckMK service group. Service groups provide logical grouping of services for rules and reporting. Does not require activation.",
	IDDescription:    "The unique identifier for the service group (same as name).",
	NameDescription:  "Unique name for the service group. Cannot be changed after creation.",
	AliasDescription: "Human-readable alias/title for the service group.",

	Create: func(ctx context.Context, c *client.Client, name, alias string) (string, string, error) {
		result, err := c.CreateServiceGroup(ctx, &client.ServiceGroupCreateRequest{Name: name, Alias: alias})
		if err != nil {
			return "", "", err
		}
		return result.ID, result.Title, nil
	},
	Get: func(ctx context.Context, c *client.Client, name string) (string, string, error) {
		result, err := c.GetServiceGroup(ctx, name)
		if err != nil {
			return "", "", err
		}
		return result.ID, result.Title, nil
	},
	GetWithETag: func(ctx context.Context, c *client.Client, name string) (string, string, string, error) {
		result, err := c.GetServiceGroupWithETag(ctx, name)
		if err != nil {
			return "", "", "", err
		}
		return result.ServiceGroup.ID, result.ServiceGroup.Title, result.ETag, nil
	},
	Update: func(ctx context.Context, c *client.Client, name, alias, etag string) (string, string, error) {
		result, err := c.UpdateServiceGroup(ctx, name, &client.ServiceGroupUpdateRequest{Alias: alias}, etag)
		if err != nil {
			return "", "", err
		}
		return result.ID, result.Title, nil
	},
	Delete: func(ctx context.Context, c *client.Client, name, etag string) error {
		return c.DeleteServiceGroup(ctx, name, etag)
	},
}
