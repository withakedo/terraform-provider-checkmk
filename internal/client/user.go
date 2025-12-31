package client

import (
	"context"
)

// User represents a CheckMK user
type User struct {
	ID         string         `json:"id"`
	Title      string         `json:"title"`
	Extensions UserExtensions `json:"extensions"`
	Links      []Link         `json:"links,omitempty"`
}

// UserExtensions contains user extension data
type UserExtensions struct {
	Fullname       string         `json:"fullname"`
	ContactOptions ContactOptions `json:"contact_options,omitempty"`
	ContactGroups  []string       `json:"contactgroups,omitempty"`
	Roles          []string       `json:"roles,omitempty"`
	DisableLogin   bool           `json:"disable_login,omitempty"`
	AuthOption     struct {
		AuthType string `json:"auth_type,omitempty"`
	} `json:"auth_option,omitempty"`
}

// ContactOptions contains contact-related settings
type ContactOptions struct {
	Email string `json:"email,omitempty"`
}

// UserCreateRequest is the request body for creating a user
type UserCreateRequest struct {
	Username       string          `json:"username"`
	Fullname       string          `json:"fullname"`
	ContactOptions *ContactOptions `json:"contact_options,omitempty"`
	ContactGroups  []string        `json:"contactgroups,omitempty"`
	Roles          []string        `json:"roles,omitempty"`
	DisableLogin   bool            `json:"disable_login,omitempty"`
	AuthOption     *AuthOption     `json:"auth_option,omitempty"`
}

// UserUpdateRequest is the request body for updating a user
type UserUpdateRequest struct {
	Fullname       string          `json:"fullname,omitempty"`
	ContactOptions *ContactOptions `json:"contact_options,omitempty"`
	ContactGroups  []string        `json:"contactgroups,omitempty"`
	Roles          []string        `json:"roles,omitempty"`
	DisableLogin   *bool           `json:"disable_login,omitempty"`
	AuthOption     *AuthOption     `json:"auth_option,omitempty"`
}

// AuthOption specifies authentication settings for a user
type AuthOption struct {
	AuthType         string `json:"auth_type"`                   // "password", "automation", "remove"
	Password         string `json:"password,omitempty"`          // For password auth
	AutomationSecret string `json:"automation_secret,omitempty"` // For automation auth
	EnforceChange    bool   `json:"enforce_password_change,omitempty"`
}

// UserWithETag wraps a User with its ETag for strict resource locking
type UserWithETag struct {
	User *User
	ETag string
}

// CreateUser creates a new user in CheckMK
func (c *Client) CreateUser(ctx context.Context, req *UserCreateRequest) (*User, error) {
	return CreateResource[User](c, ctx, UserConfig, req, req.Username)
}

// GetUser retrieves a user by username
func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	return GetResource[User](c, ctx, UserConfig, username)
}

// GetUserWithETag retrieves a user by username and returns it with its ETag
func (c *Client) GetUserWithETag(ctx context.Context, username string) (*UserWithETag, error) {
	result, err := GetResourceWithETag[User](c, ctx, UserConfig, username)
	if err != nil {
		return nil, err
	}
	return &UserWithETag{
		User: result.Resource,
		ETag: result.ETag,
	}, nil
}

// UpdateUser updates an existing user
func (c *Client) UpdateUser(ctx context.Context, username string, req *UserUpdateRequest, etag string) (*User, error) {
	return UpdateResource[User](c, ctx, UserConfig, username, req, etag)
}

// DeleteUser deletes a user
func (c *Client) DeleteUser(ctx context.Context, username string, etag string) error {
	return DeleteResource(c, ctx, UserConfig, username, etag)
}

// ListUsers retrieves all users
func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	return ListResources[User](c, ctx, UserConfig)
}
