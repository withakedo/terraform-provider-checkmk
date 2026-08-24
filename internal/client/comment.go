package client

import (
	"context"
	"net/http"
	"net/url"
)

// CommentInfo describes a single active comment, as returned by
// ListComments. Mirrors the "CommentAttributes" schema from the CheckMK
// REST API.
type CommentInfo struct {
	ID                 int64  `json:"id"`
	SiteID             string `json:"site_id"`
	HostName           string `json:"host_name"`
	Author             string `json:"author"`
	Comment            string `json:"comment"`
	Persistent         bool   `json:"persistent"`
	EntryTime          string `json:"entry_time"`
	IsService          bool   `json:"is_service"`
	ServiceDescription string `json:"service_description,omitempty"`
}

type commentObjectResponse struct {
	Extensions CommentInfo `json:"extensions"`
}

type commentCollectionResponse struct {
	Value []commentObjectResponse `json:"value"`
}

// ListComments lists active comments for a host (both host- and
// service-level comments on it).
func (c *Client) ListComments(ctx context.Context, hostName string) ([]CommentInfo, error) {
	q := url.Values{}
	q.Set("host_name", hostName)

	resp, err := c.request(ctx, "GET", "/domain-types/comment/collections/all?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	var result commentCollectionResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	infos := make([]CommentInfo, len(result.Value))
	for i, v := range result.Value {
		infos[i] = v.Extensions
	}
	return infos, nil
}

// CreateHostCommentRequest adds a comment to a single host. Mirrors the
// "CreateHostComment" schema (comment_type = "host") from the CheckMK REST
// API.
type CreateHostCommentRequest struct {
	CommentType string `json:"comment_type"`
	HostName    string `json:"host_name"`
	Comment     string `json:"comment"`
	Persistent  bool   `json:"persistent"`
}

// CreateServiceCommentRequest adds a comment to a service on a host. Mirrors
// the "CreateServiceComment" schema (comment_type = "service") from the
// CheckMK REST API.
type CreateServiceCommentRequest struct {
	CommentType        string `json:"comment_type"`
	HostName           string `json:"host_name"`
	ServiceDescription string `json:"service_description"`
	Comment            string `json:"comment"`
	Persistent         bool   `json:"persistent"`
}

// DeleteCommentsByParamsRequest deletes comment(s) by host (and optionally
// specific services) rather than by server-generated comment id. CheckMK's
// create-comment endpoints respond 204 No Content with no id, so matching by
// the original parameters avoids having to look up the generated comment_id
// afterward.
type DeleteCommentsByParamsRequest struct {
	DeleteType          string   `json:"delete_type"`
	HostName            string   `json:"host_name"`
	ServiceDescriptions []string `json:"service_descriptions,omitempty"`
}

// CreateHostComment adds a comment to a host.
func (c *Client) CreateHostComment(ctx context.Context, req *CreateHostCommentRequest) error {
	resp, err := c.request(ctx, "POST", "/domain-types/comment/collections/host", req)
	if err != nil {
		return err
	}
	if apiErr := HandleNotFound(resp, "Host", req.HostName); apiErr != nil {
		resp.Body.Close()
		return apiErr
	}
	return c.handleResponse(resp, nil)
}

// CreateServiceComment adds a comment to a service on a host.
func (c *Client) CreateServiceComment(ctx context.Context, req *CreateServiceCommentRequest) error {
	resp, err := c.request(ctx, "POST", "/domain-types/comment/collections/service", req)
	if err != nil {
		return err
	}
	if apiErr := HandleNotFound(resp, "Host", req.HostName); apiErr != nil {
		resp.Body.Close()
		return apiErr
	}
	return c.handleResponse(resp, nil)
}

// DeleteCommentsByParams removes the comment(s) matching a host and,
// optionally, a specific set of services. Deleting comments that no longer
// exist is treated as success, matching the idempotent-delete convention
// used elsewhere in this client.
func (c *Client) DeleteCommentsByParams(ctx context.Context, hostName string, serviceDescriptions []string) error {
	req := &DeleteCommentsByParamsRequest{
		DeleteType:          "params",
		HostName:            hostName,
		ServiceDescriptions: serviceDescriptions,
	}

	resp, err := c.request(ctx, "POST", "/domain-types/comment/actions/delete/invoke", req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil
	}
	return c.handleResponse(resp, nil)
}
