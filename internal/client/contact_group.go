package client

import (
	"context"
)

// ContactGroup represents a CheckMK contact group
type ContactGroup struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Links []Link `json:"links,omitempty"`
}

// ContactGroupCreateRequest is the request body for creating a contact group
type ContactGroupCreateRequest struct {
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

// ContactGroupUpdateRequest is the request body for updating a contact group
type ContactGroupUpdateRequest struct {
	Alias string `json:"alias"`
}

// ContactGroupWithETag wraps a ContactGroup with its ETag for strict resource locking
type ContactGroupWithETag struct {
	ContactGroup *ContactGroup
	ETag         string
}

// CreateContactGroup creates a new contact group in CheckMK
func (c *Client) CreateContactGroup(ctx context.Context, req *ContactGroupCreateRequest) (*ContactGroup, error) {
	return CreateResource[ContactGroup](c, ctx, ContactGroupConfig, req, req.Name)
}

// GetContactGroup retrieves a contact group by name
func (c *Client) GetContactGroup(ctx context.Context, name string) (*ContactGroup, error) {
	return GetResource[ContactGroup](c, ctx, ContactGroupConfig, name)
}

// GetContactGroupWithETag retrieves a contact group by name and returns it with its ETag
func (c *Client) GetContactGroupWithETag(ctx context.Context, name string) (*ContactGroupWithETag, error) {
	result, err := GetResourceWithETag[ContactGroup](c, ctx, ContactGroupConfig, name)
	if err != nil {
		return nil, err
	}
	return &ContactGroupWithETag{
		ContactGroup: result.Resource,
		ETag:         result.ETag,
	}, nil
}

// UpdateContactGroup updates an existing contact group
func (c *Client) UpdateContactGroup(ctx context.Context, name string, req *ContactGroupUpdateRequest, etag string) (*ContactGroup, error) {
	return UpdateResource[ContactGroup](c, ctx, ContactGroupConfig, name, req, etag)
}

// DeleteContactGroup deletes a contact group
func (c *Client) DeleteContactGroup(ctx context.Context, name string, etag string) error {
	return DeleteResource(c, ctx, ContactGroupConfig, name, etag)
}

// ListContactGroups retrieves all contact groups
func (c *Client) ListContactGroups(ctx context.Context) ([]ContactGroup, error) {
	return ListResources[ContactGroup](c, ctx, ContactGroupConfig)
}
