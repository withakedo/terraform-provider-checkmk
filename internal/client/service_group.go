package client

import (
	"context"
)

// ServiceGroup represents a CheckMK service group
type ServiceGroup struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Extensions ServiceGroupExtensions `json:"extensions"`
	Links      []Link                 `json:"links,omitempty"`
}

// ServiceGroupExtensions contains service group extension data
type ServiceGroupExtensions struct {
	Alias string `json:"alias"`
}

// ServiceGroupCreateRequest is the request body for creating a service group
type ServiceGroupCreateRequest struct {
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

// ServiceGroupUpdateRequest is the request body for updating a service group
type ServiceGroupUpdateRequest struct {
	Alias string `json:"alias"`
}

// ServiceGroupWithETag wraps a ServiceGroup with its ETag for strict resource locking
type ServiceGroupWithETag struct {
	ServiceGroup *ServiceGroup
	ETag         string
}

// CreateServiceGroup creates a new service group in CheckMK
func (c *Client) CreateServiceGroup(ctx context.Context, req *ServiceGroupCreateRequest) (*ServiceGroup, error) {
	return CreateResource[ServiceGroup](c, ctx, ServiceGroupConfig, req, req.Name)
}

// GetServiceGroup retrieves a service group by name
func (c *Client) GetServiceGroup(ctx context.Context, name string) (*ServiceGroup, error) {
	return GetResource[ServiceGroup](c, ctx, ServiceGroupConfig, name)
}

// GetServiceGroupWithETag retrieves a service group by name and returns it with its ETag
func (c *Client) GetServiceGroupWithETag(ctx context.Context, name string) (*ServiceGroupWithETag, error) {
	result, err := GetResourceWithETag[ServiceGroup](c, ctx, ServiceGroupConfig, name)
	if err != nil {
		return nil, err
	}
	return &ServiceGroupWithETag{
		ServiceGroup: result.Resource,
		ETag:         result.ETag,
	}, nil
}

// UpdateServiceGroup updates an existing service group
func (c *Client) UpdateServiceGroup(ctx context.Context, name string, req *ServiceGroupUpdateRequest, etag string) (*ServiceGroup, error) {
	return UpdateResource[ServiceGroup](c, ctx, ServiceGroupConfig, name, req, etag)
}

// DeleteServiceGroup deletes a service group
func (c *Client) DeleteServiceGroup(ctx context.Context, name string, etag string) error {
	return DeleteResource(c, ctx, ServiceGroupConfig, name, etag)
}

// ListServiceGroups retrieves all service groups
func (c *Client) ListServiceGroups(ctx context.Context) ([]ServiceGroup, error) {
	return ListResources[ServiceGroup](c, ctx, ServiceGroupConfig)
}
