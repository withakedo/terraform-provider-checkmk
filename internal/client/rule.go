package client

import (
	"context"
	"fmt"
	"net/url"
)

// Rule represents a CheckMK rule
type Rule struct {
	ID         string         `json:"id"`
	Title      string         `json:"title"`
	DomainType string         `json:"domainType,omitempty"`
	Extensions RuleExtensions `json:"extensions"`
	Links      []Link         `json:"links,omitempty"`
}

// RuleExtensions contains rule extension data
type RuleExtensions struct {
	Ruleset     string                 `json:"ruleset"`
	Folder      string                 `json:"folder"`
	FolderIndex int                    `json:"folder_index,omitempty"`
	Properties  RuleProperties         `json:"properties"`
	ValueRaw    string                 `json:"value_raw"`
	Conditions  map[string]interface{} `json:"conditions"`
}

// RuleProperties contains rule metadata
type RuleProperties struct {
	Description string `json:"description"`
	Comment     string `json:"comment,omitempty"`
	Disabled    bool   `json:"disabled"`
}

// RuleCreateRequest is the request body for creating a rule
type RuleCreateRequest struct {
	Ruleset    string                 `json:"ruleset"`
	Folder     string                 `json:"folder"`
	Properties RuleProperties         `json:"properties"`
	ValueRaw   string                 `json:"value_raw"`
	Conditions map[string]interface{} `json:"conditions"`
}

// RuleUpdateRequest is the request body for updating a rule
type RuleUpdateRequest struct {
	Properties RuleProperties         `json:"properties"`
	ValueRaw   string                 `json:"value_raw"`
	Conditions map[string]interface{} `json:"conditions"`
}

// RuleWithETag wraps a Rule with its ETag for strict resource locking
type RuleWithETag struct {
	Rule *Rule
	ETag string
}

// CreateRule creates a new rule in CheckMK
func (c *Client) CreateRule(ctx context.Context, req *RuleCreateRequest) (*Rule, error) {
	resp, err := c.request(ctx, "POST", "/domain-types/rule/collections/all", req)
	if err != nil {
		return nil, err
	}

	if err := HandleConflict(resp, "Rule", req.Properties.Description); err != nil {
		return nil, err
	}

	var rule Rule
	if err := c.handleResponse(resp, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// GetRule retrieves a rule by ID
func (c *Client) GetRule(ctx context.Context, ruleID string) (*Rule, error) {
	result, err := c.GetRuleWithETag(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	return result.Rule, nil
}

// GetRuleWithETag retrieves a rule by ID and returns it with its ETag
func (c *Client) GetRuleWithETag(ctx context.Context, ruleID string) (*RuleWithETag, error) {
	path := fmt.Sprintf("/objects/rule/%s", ruleID)
	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if err := HandleNotFound(resp, "Rule", ruleID); err != nil {
		return nil, err
	}

	var rule Rule
	if err := c.handleResponse(resp, &rule); err != nil {
		return nil, err
	}

	return &RuleWithETag{
		Rule: &rule,
		ETag: resp.Header.Get("ETag"),
	}, nil
}

// GetRulesByRuleset retrieves all rules in a specific ruleset
func (c *Client) GetRulesByRuleset(ctx context.Context, ruleset string) ([]Rule, error) {
	// URL encode the ruleset parameter
	path := fmt.Sprintf("/domain-types/rule/collections/all?ruleset_name=%s", url.QueryEscape(ruleset))

	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result ListResponse[Rule]
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Value, nil
}

// UpdateRule updates an existing rule
// If etag is provided (non-empty), it will be used in the If-Match header for strict resource locking
// If etag is empty, uses If-Match: * to bypass ETag validation
func (c *Client) UpdateRule(ctx context.Context, ruleID string, req *RuleUpdateRequest, etag string) (*Rule, error) {
	path := fmt.Sprintf("/objects/rule/%s", ruleID)

	resp, err := c.requestWithHeaders(ctx, "PUT", path, req, BuildETagHeaders(etag))
	if err != nil {
		return nil, err
	}

	if err := HandlePreconditionFailed(resp, "Rule", ruleID); err != nil {
		return nil, err
	}

	// Handle 428 Precondition Required (drift detection)
	if resp.StatusCode == 428 {
		return nil, NewDriftError("Rule", ruleID)
	}

	if err := HandleNotFound(resp, "Rule", ruleID); err != nil {
		return nil, err
	}

	var rule Rule
	if err := c.handleResponse(resp, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// DeleteRule deletes a rule
// If etag is provided (non-empty), it will be used in the If-Match header for strict resource locking
// If etag is empty, uses If-Match: * to bypass ETag validation
func (c *Client) DeleteRule(ctx context.Context, ruleID string, etag string) error {
	path := fmt.Sprintf("/objects/rule/%s", ruleID)

	resp, err := c.requestWithHeaders(ctx, "DELETE", path, nil, BuildETagHeaders(etag))
	if err != nil {
		return err
	}

	if err := HandlePreconditionFailed(resp, "Rule", ruleID); err != nil {
		return err
	}

	// 404 is acceptable for delete (idempotent)
	if IsNotFoundResponse(resp) {
		return nil
	}

	return c.handleResponse(resp, nil)
}
