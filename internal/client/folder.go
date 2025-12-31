package client

import (
	"context"
	"fmt"
	"strings"
)

// Folder represents a CheckMK folder
type Folder struct {
	ID         string           `json:"id"`
	Title      string           `json:"title"`
	Extensions FolderExtensions `json:"extensions"`
	Links      []Link           `json:"links,omitempty"`
}

// FolderExtensions contains folder extension data
type FolderExtensions struct {
	Path       string                 `json:"path"`
	Attributes map[string]interface{} `json:"attributes"`
}

// FolderCreateRequest is the request body for creating a folder
type FolderCreateRequest struct {
	Name       string                 `json:"name"`
	Title      string                 `json:"title"`
	Parent     string                 `json:"parent"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// FolderUpdateRequest is the request body for updating a folder
type FolderUpdateRequest struct {
	Title      string                 `json:"title,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// FolderWithETag wraps a Folder with its ETag for strict resource locking
type FolderWithETag struct {
	Folder *Folder
	ETag   string
}

// PathToAPIFormat converts user-friendly path format to CheckMK API format
// User format: /IE/MUL/Server
// API format:  ~IE~MUL~Server
func PathToAPIFormat(path string) string {
	if path == "/" {
		return "~"
	}
	// Remove leading slash and replace / with ~
	path = strings.TrimPrefix(path, "/")
	return "~" + strings.ReplaceAll(path, "/", "~")
}

// PathFromAPIFormat converts CheckMK API format to user-friendly path format
// API format:  ~IE~MUL~Server
// User format: /IE/MUL/Server
func PathFromAPIFormat(apiPath string) string {
	if apiPath == "~" {
		return "/"
	}
	// Remove leading ~ and replace ~ with /
	path := strings.TrimPrefix(apiPath, "~")
	return "/" + strings.ReplaceAll(path, "~", "/")
}

// CreateFolder creates a new folder in CheckMK
func (c *Client) CreateFolder(ctx context.Context, req *FolderCreateRequest) (*Folder, error) {
	// Convert parent path to API format
	apiReq := &FolderCreateRequest{
		Name:       req.Name,
		Title:      req.Title,
		Parent:     PathToAPIFormat(req.Parent),
		Attributes: req.Attributes,
	}

	resp, err := c.request(ctx, "POST", "/domain-types/folder_config/collections/all", apiReq)
	if err != nil {
		return nil, err
	}

	if err := HandleConflict(resp, "Folder", fmt.Sprintf("%s/%s", req.Parent, req.Name)); err != nil {
		return nil, err
	}

	var folder Folder
	if err := c.handleResponse(resp, &folder); err != nil {
		return nil, err
	}

	return &folder, nil
}

// GetFolder retrieves a folder by path
func (c *Client) GetFolder(ctx context.Context, path string) (*Folder, error) {
	result, err := c.GetFolderWithETag(ctx, path)
	if err != nil {
		return nil, err
	}
	return result.Folder, nil
}

// GetFolderWithETag retrieves a folder by path and returns it with its ETag
func (c *Client) GetFolderWithETag(ctx context.Context, path string) (*FolderWithETag, error) {
	apiPath := fmt.Sprintf("/objects/folder_config/%s", PathToAPIFormat(path))
	resp, err := c.request(ctx, "GET", apiPath, nil)
	if err != nil {
		return nil, err
	}

	if err := HandleNotFound(resp, "Folder", path); err != nil {
		return nil, err
	}

	var folder Folder
	if err := c.handleResponse(resp, &folder); err != nil {
		return nil, err
	}

	return &FolderWithETag{
		Folder: &folder,
		ETag:   resp.Header.Get("ETag"),
	}, nil
}

// UpdateFolder updates an existing folder
// If etag is provided (non-empty), it will be used in the If-Match header for strict resource locking
// If etag is empty, uses If-Match: * to bypass ETag validation
func (c *Client) UpdateFolder(ctx context.Context, path string, req *FolderUpdateRequest, etag string) (*Folder, error) {
	apiPath := fmt.Sprintf("/objects/folder_config/%s", PathToAPIFormat(path))

	resp, err := c.requestWithHeaders(ctx, "PUT", apiPath, req, BuildETagHeaders(etag))
	if err != nil {
		return nil, err
	}

	if err := HandlePreconditionFailed(resp, "Folder", path); err != nil {
		return nil, err
	}

	if err := HandleNotFound(resp, "Folder", path); err != nil {
		return nil, err
	}

	var folder Folder
	if err := c.handleResponse(resp, &folder); err != nil {
		return nil, err
	}

	return &folder, nil
}

// DeleteFolder deletes a folder
// If etag is provided (non-empty), it will be used in the If-Match header for strict resource locking
// If etag is empty, uses If-Match: * to bypass ETag validation
func (c *Client) DeleteFolder(ctx context.Context, path string, etag string) error {
	apiPath := fmt.Sprintf("/objects/folder_config/%s", PathToAPIFormat(path))

	resp, err := c.requestWithHeaders(ctx, "DELETE", apiPath, nil, BuildETagHeaders(etag))
	if err != nil {
		return err
	}

	if err := HandlePreconditionFailed(resp, "Folder", path); err != nil {
		return err
	}

	// 404 is acceptable for delete (idempotent)
	if IsNotFoundResponse(resp) {
		return nil
	}

	return c.handleResponse(resp, nil)
}

// ListFolders retrieves all folders (optionally recursive)
func (c *Client) ListFolders(ctx context.Context, recursive bool) ([]Folder, error) {
	path := "/domain-types/folder_config/collections/all"
	if recursive {
		path += "?recursive=true"
	}

	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result ListResponse[Folder]
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Value, nil
}
