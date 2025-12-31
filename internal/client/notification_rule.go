package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// NotificationRule represents a CheckMK notification rule
type NotificationRule struct {
	ID         string                     `json:"id"`
	Title      string                     `json:"title"`
	DomainType string                     `json:"domainType,omitempty"`
	Extensions NotificationRuleExtensions `json:"extensions"`
	Links      []Link                     `json:"links,omitempty"`
}

// NotificationRuleExtensions contains notification rule extension data
type NotificationRuleExtensions struct {
	RuleConfig json.RawMessage `json:"rule_config"`
}

// NotificationRuleCreateRequest is the request body for creating a notification rule
type NotificationRuleCreateRequest struct {
	RuleConfig json.RawMessage `json:"rule_config"`
}

// NotificationRuleUpdateRequest is the request body for updating a notification rule
type NotificationRuleUpdateRequest struct {
	RuleConfig json.RawMessage `json:"rule_config"`
}

// NotificationRuleWithETag wraps a NotificationRule with its ETag for strict resource locking
type NotificationRuleWithETag struct {
	Rule *NotificationRule
	ETag string
}

// CreateNotificationRule creates a new notification rule in CheckMK
func (c *Client) CreateNotificationRule(ctx context.Context, req *NotificationRuleCreateRequest) (*NotificationRule, error) {
	resp, err := c.request(ctx, "POST", "/domain-types/notification_rule/collections/all", req)
	if err != nil {
		return nil, err
	}

	if err := HandleConflict(resp, "NotificationRule", ""); err != nil {
		return nil, err
	}

	var rule NotificationRule
	if err := c.handleResponse(resp, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// GetNotificationRule retrieves a notification rule by ID
func (c *Client) GetNotificationRule(ctx context.Context, ruleID string) (*NotificationRule, error) {
	result, err := c.GetNotificationRuleWithETag(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	return result.Rule, nil
}

// GetNotificationRuleWithETag retrieves a notification rule by ID and returns it with its ETag
func (c *Client) GetNotificationRuleWithETag(ctx context.Context, ruleID string) (*NotificationRuleWithETag, error) {
	path := fmt.Sprintf("/objects/notification_rule/%s", ruleID)
	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if err := HandleNotFound(resp, "NotificationRule", ruleID); err != nil {
		return nil, err
	}

	var rule NotificationRule
	if err := c.handleResponse(resp, &rule); err != nil {
		return nil, err
	}

	return &NotificationRuleWithETag{
		Rule: &rule,
		ETag: resp.Header.Get("ETag"),
	}, nil
}

// ListNotificationRules retrieves all notification rules
func (c *Client) ListNotificationRules(ctx context.Context) ([]NotificationRule, error) {
	resp, err := c.request(ctx, "GET", "/domain-types/notification_rule/collections/all", nil)
	if err != nil {
		return nil, err
	}

	var result ListResponse[NotificationRule]
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Value, nil
}

// UpdateNotificationRule updates an existing notification rule
// If etag is provided (non-empty), it will be used in the If-Match header for strict resource locking
// If etag is empty, uses If-Match: * to bypass ETag validation
func (c *Client) UpdateNotificationRule(ctx context.Context, ruleID string, req *NotificationRuleUpdateRequest, etag string) (*NotificationRule, error) {
	path := fmt.Sprintf("/objects/notification_rule/%s", ruleID)

	resp, err := c.requestWithHeaders(ctx, "PUT", path, req, BuildETagHeaders(etag))
	if err != nil {
		return nil, err
	}

	if err := HandlePreconditionFailed(resp, "NotificationRule", ruleID); err != nil {
		return nil, err
	}

	// Handle 428 Precondition Required (drift detection)
	if resp.StatusCode == 428 {
		return nil, NewDriftError("NotificationRule", ruleID)
	}

	if err := HandleNotFound(resp, "NotificationRule", ruleID); err != nil {
		return nil, err
	}

	var rule NotificationRule
	if err := c.handleResponse(resp, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// DeleteNotificationRule deletes a notification rule
// CheckMK uses POST to /actions/delete/invoke for deletion instead of HTTP DELETE
// If etag is provided (non-empty), it will be used in the If-Match header for strict resource locking
// If etag is empty, uses If-Match: * to bypass ETag validation
func (c *Client) DeleteNotificationRule(ctx context.Context, ruleID string, etag string) error {
	// CheckMK uses a POST action endpoint for deletion
	path := fmt.Sprintf("/objects/notification_rule/%s/actions/delete/invoke", ruleID)

	resp, err := c.requestWithHeaders(ctx, "POST", path, nil, BuildETagHeaders(etag))
	if err != nil {
		return err
	}

	if err := HandlePreconditionFailed(resp, "NotificationRule", ruleID); err != nil {
		return err
	}

	// 404 is acceptable for delete (idempotent)
	if IsNotFoundResponse(resp) {
		return nil
	}

	return c.handleResponse(resp, nil)
}
