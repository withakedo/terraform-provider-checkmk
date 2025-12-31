package client

import (
	"context"
)

// Password represents a CheckMK password
type Password struct {
	ID         string             `json:"id"`
	Title      string             `json:"title"`
	Extensions PasswordExtensions `json:"extensions"`
	Links      []Link             `json:"links,omitempty"`
}

// PasswordExtensions contains password extension data
type PasswordExtensions struct {
	Owner            string   `json:"owned_by,omitempty"`
	Shared           []string `json:"shared,omitempty"`
	CustomerID       string   `json:"customer,omitempty"`
	Comment          string   `json:"comment,omitempty"`
	DocumentationURL string   `json:"documentation_url,omitempty"`
	EditableBy       string   `json:"editable_by,omitempty"`
}

// PasswordCreateRequest is the request body for creating a password
type PasswordCreateRequest struct {
	Ident            string   `json:"ident"`
	Title            string   `json:"title"`
	Password         string   `json:"password"`
	Owner            string   `json:"owner,omitempty"`
	Shared           []string `json:"shared,omitempty"`
	CustomerID       string   `json:"customer,omitempty"`
	Comment          string   `json:"comment,omitempty"`
	DocumentationURL string   `json:"documentation_url,omitempty"`
	EditableBy       string   `json:"editable_by,omitempty"`
}

// PasswordUpdateRequest is the request body for updating a password
type PasswordUpdateRequest struct {
	Title            string   `json:"title,omitempty"`
	Password         string   `json:"password,omitempty"`
	Owner            string   `json:"owner,omitempty"`
	Shared           []string `json:"shared,omitempty"`
	CustomerID       string   `json:"customer,omitempty"`
	Comment          string   `json:"comment,omitempty"`
	DocumentationURL string   `json:"documentation_url,omitempty"`
	EditableBy       string   `json:"editable_by,omitempty"`
}

// PasswordWithETag wraps a Password with its ETag for strict resource locking
type PasswordWithETag struct {
	Password *Password
	ETag     string
}

// CreatePassword creates a new password in CheckMK
func (c *Client) CreatePassword(ctx context.Context, req *PasswordCreateRequest) (*Password, error) {
	return CreateResource[Password](c, ctx, PasswordConfig, req, req.Ident)
}

// GetPassword retrieves a password by ID
func (c *Client) GetPassword(ctx context.Context, passwordID string) (*Password, error) {
	return GetResource[Password](c, ctx, PasswordConfig, passwordID)
}

// GetPasswordWithETag retrieves a password by ID and returns it with its ETag
func (c *Client) GetPasswordWithETag(ctx context.Context, passwordID string) (*PasswordWithETag, error) {
	result, err := GetResourceWithETag[Password](c, ctx, PasswordConfig, passwordID)
	if err != nil {
		return nil, err
	}
	return &PasswordWithETag{
		Password: result.Resource,
		ETag:     result.ETag,
	}, nil
}

// UpdatePassword updates an existing password
func (c *Client) UpdatePassword(ctx context.Context, passwordID string, req *PasswordUpdateRequest, etag string) (*Password, error) {
	return UpdateResource[Password](c, ctx, PasswordConfig, passwordID, req, etag)
}

// DeletePassword deletes a password
func (c *Client) DeletePassword(ctx context.Context, passwordID string, etag string) error {
	return DeleteResource(c, ctx, PasswordConfig, passwordID, etag)
}

// ListPasswords retrieves all passwords
func (c *Client) ListPasswords(ctx context.Context) ([]Password, error) {
	return ListResources[Password](c, ctx, PasswordConfig)
}
