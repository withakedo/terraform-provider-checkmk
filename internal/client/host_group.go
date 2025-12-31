package client

import (
	"context"
)

// HostGroup represents a CheckMK host group
type HostGroup struct {
	ID         string              `json:"id"`
	Title      string              `json:"title"`
	Extensions HostGroupExtensions `json:"extensions"`
	Links      []Link              `json:"links,omitempty"`
}

// HostGroupExtensions contains host group extension data
type HostGroupExtensions struct {
	Alias string `json:"alias"`
}

// HostGroupCreateRequest is the request body for creating a host group
type HostGroupCreateRequest struct {
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

// HostGroupUpdateRequest is the request body for updating a host group
type HostGroupUpdateRequest struct {
	Alias string `json:"alias"`
}

// HostGroupWithETag wraps a HostGroup with its ETag for strict resource locking
type HostGroupWithETag struct {
	HostGroup *HostGroup
	ETag      string
}

// CreateHostGroup creates a new host group in CheckMK
func (c *Client) CreateHostGroup(ctx context.Context, req *HostGroupCreateRequest) (*HostGroup, error) {
	return CreateResource[HostGroup](c, ctx, HostGroupConfig, req, req.Name)
}

// GetHostGroup retrieves a host group by name
func (c *Client) GetHostGroup(ctx context.Context, name string) (*HostGroup, error) {
	return GetResource[HostGroup](c, ctx, HostGroupConfig, name)
}

// GetHostGroupWithETag retrieves a host group by name and returns it with its ETag
func (c *Client) GetHostGroupWithETag(ctx context.Context, name string) (*HostGroupWithETag, error) {
	result, err := GetResourceWithETag[HostGroup](c, ctx, HostGroupConfig, name)
	if err != nil {
		return nil, err
	}
	return &HostGroupWithETag{
		HostGroup: result.Resource,
		ETag:      result.ETag,
	}, nil
}

// UpdateHostGroup updates an existing host group
func (c *Client) UpdateHostGroup(ctx context.Context, name string, req *HostGroupUpdateRequest, etag string) (*HostGroup, error) {
	return UpdateResource[HostGroup](c, ctx, HostGroupConfig, name, req, etag)
}

// DeleteHostGroup deletes a host group
func (c *Client) DeleteHostGroup(ctx context.Context, name string, etag string) error {
	return DeleteResource(c, ctx, HostGroupConfig, name, etag)
}

// ListHostGroups retrieves all host groups
func (c *Client) ListHostGroups(ctx context.Context) ([]HostGroup, error) {
	return ListResources[HostGroup](c, ctx, HostGroupConfig)
}
