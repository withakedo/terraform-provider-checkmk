package client

import (
	"context"
)

// AuxTag represents a CheckMK auxiliary tag
type AuxTag struct {
	ID         string           `json:"id"`
	Title      string           `json:"title"`
	Extensions AuxTagExtensions `json:"extensions"`
	Links      []Link           `json:"links,omitempty"`
}

// AuxTagExtensions contains aux tag extension data
type AuxTagExtensions struct {
	Topic string `json:"topic,omitempty"`
	Help  string `json:"help,omitempty"`
}

// AuxTagCreateRequest is the request body for creating an aux tag
type AuxTagCreateRequest struct {
	AuxTagID string `json:"aux_tag_id"`
	Title    string `json:"title"`
	Topic    string `json:"topic"`
	Help     string `json:"help,omitempty"`
}

// AuxTagUpdateRequest is the request body for updating an aux tag
type AuxTagUpdateRequest struct {
	Title string `json:"title"`
	Topic string `json:"topic,omitempty"`
	Help  string `json:"help,omitempty"`
}

// AuxTagWithETag wraps an AuxTag with its ETag for strict resource locking
type AuxTagWithETag struct {
	AuxTag *AuxTag
	ETag   string
}

// CreateAuxTag creates a new auxiliary tag in CheckMK
func (c *Client) CreateAuxTag(ctx context.Context, req *AuxTagCreateRequest) (*AuxTag, error) {
	return CreateResource[AuxTag](c, ctx, AuxTagConfig, req, req.AuxTagID)
}

// GetAuxTag retrieves an auxiliary tag by ID
func (c *Client) GetAuxTag(ctx context.Context, id string) (*AuxTag, error) {
	return GetResource[AuxTag](c, ctx, AuxTagConfig, id)
}

// GetAuxTagWithETag retrieves an auxiliary tag by ID and returns it with its ETag
func (c *Client) GetAuxTagWithETag(ctx context.Context, id string) (*AuxTagWithETag, error) {
	result, err := GetResourceWithETag[AuxTag](c, ctx, AuxTagConfig, id)
	if err != nil {
		return nil, err
	}
	return &AuxTagWithETag{
		AuxTag: result.Resource,
		ETag:   result.ETag,
	}, nil
}

// UpdateAuxTag updates an existing auxiliary tag
func (c *Client) UpdateAuxTag(ctx context.Context, id string, req *AuxTagUpdateRequest, etag string) (*AuxTag, error) {
	return UpdateResource[AuxTag](c, ctx, AuxTagConfig, id, req, etag)
}

// DeleteAuxTag deletes an auxiliary tag using the action-based delete endpoint
// (POST to /actions/delete/invoke instead of standard DELETE)
func (c *Client) DeleteAuxTag(ctx context.Context, id string, etag string) error {
	return DeleteResourceAction(c, ctx, AuxTagConfig, id, etag)
}

// ListAuxTags retrieves all auxiliary tags
func (c *Client) ListAuxTags(ctx context.Context) ([]AuxTag, error) {
	return ListResources[AuxTag](c, ctx, AuxTagConfig)
}
