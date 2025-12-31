package client

import (
	"context"
	"fmt"
)

// Host represents a CheckMK host
type Host struct {
	ID         string         `json:"id"`
	Title      string         `json:"title"`
	Extensions HostExtensions `json:"extensions"`
	Links      []Link         `json:"links,omitempty"`
}

// HostExtensions contains host extension data
type HostExtensions struct {
	Folder     string                 `json:"folder"`
	Attributes map[string]interface{} `json:"attributes"`
	Cluster    []string               `json:"cluster_nodes,omitempty"`
}

// HostCreateRequest is the request body for creating a host
type HostCreateRequest struct {
	HostName   string                 `json:"host_name"`
	Folder     string                 `json:"folder,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// HostUpdateRequest is the request body for updating a host
type HostUpdateRequest struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// HostWithETag wraps a Host with its ETag for strict resource locking
type HostWithETag struct {
	Host *Host
	ETag string
}

// CreateHost creates a new host in CheckMK
func (c *Client) CreateHost(ctx context.Context, req *HostCreateRequest) (*Host, error) {
	resp, err := c.request(ctx, "POST", "/domain-types/host_config/collections/all", req)
	if err != nil {
		return nil, err
	}

	if err := HandleConflict(resp, "Host", req.HostName); err != nil {
		return nil, err
	}

	var host Host
	if err := c.handleResponse(resp, &host); err != nil {
		return nil, err
	}

	return &host, nil
}

// GetHost retrieves a host by name
func (c *Client) GetHost(ctx context.Context, hostName string) (*Host, error) {
	result, err := c.GetHostWithETag(ctx, hostName)
	if err != nil {
		return nil, err
	}
	return result.Host, nil
}

// GetHostWithETag retrieves a host by name and returns it with its ETag
func (c *Client) GetHostWithETag(ctx context.Context, hostName string) (*HostWithETag, error) {
	path := fmt.Sprintf("/objects/host_config/%s", hostName)
	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if err := HandleNotFound(resp, "Host", hostName); err != nil {
		return nil, err
	}

	var host Host
	if err := c.handleResponse(resp, &host); err != nil {
		return nil, err
	}

	return &HostWithETag{
		Host: &host,
		ETag: resp.Header.Get("ETag"),
	}, nil
}

// UpdateHost updates an existing host
// If etag is provided (non-empty), it will be used in the If-Match header for strict resource locking
// If etag is empty, uses If-Match: * to bypass ETag validation
func (c *Client) UpdateHost(ctx context.Context, hostName string, req *HostUpdateRequest, etag string) (*Host, error) {
	path := fmt.Sprintf("/objects/host_config/%s", hostName)

	resp, err := c.requestWithHeaders(ctx, "PUT", path, req, BuildETagHeaders(etag))
	if err != nil {
		return nil, err
	}

	if err := HandlePreconditionFailed(resp, "Host", hostName); err != nil {
		return nil, err
	}

	// Handle 428 Precondition Required (drift detection)
	if resp.StatusCode == 428 {
		return nil, NewDriftError("Host", hostName)
	}

	if err := HandleNotFound(resp, "Host", hostName); err != nil {
		return nil, err
	}

	var host Host
	if err := c.handleResponse(resp, &host); err != nil {
		return nil, err
	}

	return &host, nil
}

// DeleteHost deletes a host
// If etag is provided (non-empty), it will be used in the If-Match header for strict resource locking
// If etag is empty, uses If-Match: * to bypass ETag validation
func (c *Client) DeleteHost(ctx context.Context, hostName string, etag string) error {
	path := fmt.Sprintf("/objects/host_config/%s", hostName)

	resp, err := c.requestWithHeaders(ctx, "DELETE", path, nil, BuildETagHeaders(etag))
	if err != nil {
		return err
	}

	if err := HandlePreconditionFailed(resp, "Host", hostName); err != nil {
		return err
	}

	// 404 is acceptable for delete (idempotent)
	if IsNotFoundResponse(resp) {
		return nil
	}

	return c.handleResponse(resp, nil)
}
