// Package client provides generic CRUD operations for CheckMK API resources.
// These functions reduce code duplication across resource-specific client methods.
package client

import (
	"context"
	"fmt"
)

// ResourceConfig defines the API paths for a CheckMK resource type.
// This configuration is used by the generic CRUD functions to determine
// which endpoints to call for each operation.
type ResourceConfig struct {
	// CollectionPath is the path for create/list operations (e.g., "/domain-types/host_group_config/collections/all")
	CollectionPath string
	// ObjectPathFmt is the format string for single-object operations (e.g., "/objects/host_group_config/%s")
	ObjectPathFmt string
	// TypeName is the human-readable resource type name for error messages (e.g., "Host group")
	TypeName string
}

// WithETag wraps a resource with its ETag for strict resource locking.
// This generic type replaces all the individual *WithETag types like
// HostGroupWithETag, ServiceGroupWithETag, etc.
type WithETag[T any] struct {
	Resource *T
	ETag     string
}

// IdentifierGetter is implemented by create request types to get the identifier for error messages.
// This is typically the "name" field used in conflict error reporting.
type IdentifierGetter interface {
	GetIdentifier() string
}

// =============================================================================
// Generic CRUD Functions
// =============================================================================

// CreateResource creates a new resource in CheckMK.
// The identifier parameter is used only for error messages (typically the name).
func CreateResource[T any, Req any](c *Client, ctx context.Context, cfg ResourceConfig, req *Req, identifier string) (*T, error) {
	resp, err := c.request(ctx, "POST", cfg.CollectionPath, req)
	if err != nil {
		return nil, err
	}

	if err := HandleConflict(resp, cfg.TypeName, identifier); err != nil {
		return nil, err
	}

	var result T
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetResource retrieves a resource by ID without ETag.
func GetResource[T any](c *Client, ctx context.Context, cfg ResourceConfig, id string) (*T, error) {
	result, err := GetResourceWithETag[T](c, ctx, cfg, id)
	if err != nil {
		return nil, err
	}
	return result.Resource, nil
}

// GetResourceWithETag retrieves a resource by ID and returns it with its ETag.
func GetResourceWithETag[T any](c *Client, ctx context.Context, cfg ResourceConfig, id string) (*WithETag[T], error) {
	path := fmt.Sprintf(cfg.ObjectPathFmt, id)
	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if err := HandleNotFound(resp, cfg.TypeName, id); err != nil {
		return nil, err
	}

	var result T
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &WithETag[T]{
		Resource: &result,
		ETag:     resp.Header.Get("ETag"),
	}, nil
}

// UpdateResource updates an existing resource.
// If etag is provided (non-empty), it will be used in the If-Match header for strict resource locking.
// If etag is empty, uses If-Match: * to bypass ETag validation.
func UpdateResource[T any, Req any](c *Client, ctx context.Context, cfg ResourceConfig, id string, req *Req, etag string) (*T, error) {
	path := fmt.Sprintf(cfg.ObjectPathFmt, id)

	resp, err := c.requestWithHeaders(ctx, "PUT", path, req, BuildETagHeaders(etag))
	if err != nil {
		return nil, err
	}

	if err := HandlePreconditionFailed(resp, cfg.TypeName, id); err != nil {
		return nil, err
	}

	if err := HandleNotFound(resp, cfg.TypeName, id); err != nil {
		return nil, err
	}

	var result T
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteResource deletes a resource using standard DELETE method.
// If etag is provided (non-empty), it will be used in the If-Match header for strict resource locking.
// If etag is empty, uses If-Match: * to bypass ETag validation.
// 404 responses are treated as success (idempotent delete).
func DeleteResource(c *Client, ctx context.Context, cfg ResourceConfig, id string, etag string) error {
	path := fmt.Sprintf(cfg.ObjectPathFmt, id)

	resp, err := c.requestWithHeaders(ctx, "DELETE", path, nil, BuildETagHeaders(etag))
	if err != nil {
		return err
	}

	if err := HandlePreconditionFailed(resp, cfg.TypeName, id); err != nil {
		return err
	}

	// 404 is acceptable for delete (idempotent)
	if IsNotFoundResponse(resp) {
		return nil
	}

	return c.handleResponse(resp, nil)
}

// DeleteResourceAction deletes a resource using action-based delete (POST to /actions/delete/invoke).
// Some CheckMK resources (like aux_tag) require this pattern instead of standard DELETE.
// If etag is provided (non-empty), it will be used in the If-Match header for strict resource locking.
// If etag is empty, uses If-Match: * to bypass ETag validation.
// 404 responses are treated as success (idempotent delete).
func DeleteResourceAction(c *Client, ctx context.Context, cfg ResourceConfig, id string, etag string) error {
	path := fmt.Sprintf(cfg.ObjectPathFmt+"/actions/delete/invoke", id)

	resp, err := c.requestWithHeaders(ctx, "POST", path, nil, BuildETagHeaders(etag))
	if err != nil {
		return err
	}

	if err := HandlePreconditionFailed(resp, cfg.TypeName, id); err != nil {
		return err
	}

	// 404 is acceptable for delete (idempotent)
	if IsNotFoundResponse(resp) {
		return nil
	}

	return c.handleResponse(resp, nil)
}

// ListResources retrieves all resources of a type.
func ListResources[T any](c *Client, ctx context.Context, cfg ResourceConfig) ([]T, error) {
	resp, err := c.request(ctx, "GET", cfg.CollectionPath, nil)
	if err != nil {
		return nil, err
	}

	var result ListResponse[T]
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Value, nil
}

// =============================================================================
// Pre-defined Resource Configurations
// =============================================================================
// These configs can be used with the generic functions to operate on specific resource types.

// HostGroupConfig is the resource configuration for host groups.
var HostGroupConfig = ResourceConfig{
	CollectionPath: "/domain-types/host_group_config/collections/all",
	ObjectPathFmt:  "/objects/host_group_config/%s",
	TypeName:       "Host group",
}

// ServiceGroupConfig is the resource configuration for service groups.
var ServiceGroupConfig = ResourceConfig{
	CollectionPath: "/domain-types/service_group_config/collections/all",
	ObjectPathFmt:  "/objects/service_group_config/%s",
	TypeName:       "Service group",
}

// ContactGroupConfig is the resource configuration for contact groups.
var ContactGroupConfig = ResourceConfig{
	CollectionPath: "/domain-types/contact_group_config/collections/all",
	ObjectPathFmt:  "/objects/contact_group_config/%s",
	TypeName:       "Contact group",
}

// TagGroupConfig is the resource configuration for tag groups.
var TagGroupConfig = ResourceConfig{
	CollectionPath: "/domain-types/host_tag_group/collections/all",
	ObjectPathFmt:  "/objects/host_tag_group/%s",
	TypeName:       "Tag group",
}

// AuxTagConfig is the resource configuration for auxiliary tags.
var AuxTagConfig = ResourceConfig{
	CollectionPath: "/domain-types/aux_tag/collections/all",
	ObjectPathFmt:  "/objects/aux_tag/%s",
	TypeName:       "Auxiliary tag",
}

// TimePeriodConfig is the resource configuration for time periods.
var TimePeriodConfig = ResourceConfig{
	CollectionPath: "/domain-types/time_period/collections/all",
	ObjectPathFmt:  "/objects/time_period/%s",
	TypeName:       "Time period",
}

// PasswordConfig is the resource configuration for passwords.
var PasswordConfig = ResourceConfig{
	CollectionPath: "/domain-types/password/collections/all",
	ObjectPathFmt:  "/objects/password/%s",
	TypeName:       "Password",
}

// UserConfig is the resource configuration for users.
var UserConfig = ResourceConfig{
	CollectionPath: "/domain-types/user_config/collections/all",
	ObjectPathFmt:  "/objects/user_config/%s",
	TypeName:       "User",
}

// FolderConfig is the resource configuration for folders.
var FolderConfig = ResourceConfig{
	CollectionPath: "/domain-types/folder_config/collections/all",
	ObjectPathFmt:  "/objects/folder_config/%s",
	TypeName:       "Folder",
}

// HostConfig is the resource configuration for hosts.
var HostConfig = ResourceConfig{
	CollectionPath: "/domain-types/host_config/collections/all",
	ObjectPathFmt:  "/objects/host_config/%s",
	TypeName:       "Host",
}

// RuleConfig is the resource configuration for rules.
var RuleConfig = ResourceConfig{
	CollectionPath: "/domain-types/rule/collections/all",
	ObjectPathFmt:  "/objects/rule/%s",
	TypeName:       "Rule",
}

// NotificationRuleConfig is the resource configuration for notification rules.
var NotificationRuleConfig = ResourceConfig{
	CollectionPath: "/domain-types/notification_rule/collections/all",
	ObjectPathFmt:  "/objects/notification_rule/%s",
	TypeName:       "Notification rule",
}
