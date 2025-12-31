package client

import (
	"context"
)

// TagGroup represents a CheckMK host tag group
type TagGroup struct {
	ID         string             `json:"id"`
	Title      string             `json:"title"`
	Extensions TagGroupExtensions `json:"extensions"`
	Links      []Link             `json:"links,omitempty"`
}

// TagGroupExtensions contains tag group extension data
type TagGroupExtensions struct {
	Topic string `json:"topic,omitempty"`
	Help  string `json:"help,omitempty"`
	Tags  []Tag  `json:"tags"`
}

// Tag represents an individual tag within a tag group
type Tag struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	AuxTags []string `json:"aux_tags,omitempty"`
}

// TagGroupCreateRequest is the request body for creating a tag group
type TagGroupCreateRequest struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Topic string `json:"topic,omitempty"`
	Help  string `json:"help,omitempty"`
	Tags  []Tag  `json:"tags"`
}

// TagGroupUpdateRequest is the request body for updating a tag group
type TagGroupUpdateRequest struct {
	Title string `json:"title"`
	Topic string `json:"topic,omitempty"`
	Help  string `json:"help,omitempty"`
	Tags  []Tag  `json:"tags"`
}

// TagGroupWithETag wraps a TagGroup with its ETag for strict resource locking
type TagGroupWithETag struct {
	TagGroup *TagGroup
	ETag     string
}

// CreateTagGroup creates a new tag group in CheckMK
func (c *Client) CreateTagGroup(ctx context.Context, req *TagGroupCreateRequest) (*TagGroup, error) {
	return CreateResource[TagGroup](c, ctx, TagGroupConfig, req, req.ID)
}

// GetTagGroup retrieves a tag group by ID
func (c *Client) GetTagGroup(ctx context.Context, id string) (*TagGroup, error) {
	return GetResource[TagGroup](c, ctx, TagGroupConfig, id)
}

// GetTagGroupWithETag retrieves a tag group by ID and returns it with its ETag
func (c *Client) GetTagGroupWithETag(ctx context.Context, id string) (*TagGroupWithETag, error) {
	result, err := GetResourceWithETag[TagGroup](c, ctx, TagGroupConfig, id)
	if err != nil {
		return nil, err
	}
	return &TagGroupWithETag{
		TagGroup: result.Resource,
		ETag:     result.ETag,
	}, nil
}

// UpdateTagGroup updates an existing tag group
func (c *Client) UpdateTagGroup(ctx context.Context, id string, req *TagGroupUpdateRequest, etag string) (*TagGroup, error) {
	return UpdateResource[TagGroup](c, ctx, TagGroupConfig, id, req, etag)
}

// DeleteTagGroup deletes a tag group
func (c *Client) DeleteTagGroup(ctx context.Context, id string, etag string) error {
	return DeleteResource(c, ctx, TagGroupConfig, id, etag)
}

// ListTagGroups retrieves all tag groups
func (c *Client) ListTagGroups(ctx context.Context) ([]TagGroup, error) {
	return ListResources[TagGroup](c, ctx, TagGroupConfig)
}
