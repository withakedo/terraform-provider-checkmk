package client

import (
	"context"
	"fmt"
)

// CreateBIAggregation creates a BI aggregation. body must be a JSON-shaped
// map matching CheckMK's "BIAggregationEndpoint" schema (id, pack_id,
// comment, customer, groups, node, aggregation_visualization,
// computation_options). This is deliberately untyped: the aggregation
// "node" is a recursive, highly variable rule tree (search/action nodes
// that can themselves nest further condition/aggregation nodes), which
// would require hand-rolling and maintaining a large, version-sensitive
// nested schema to type safely. checkmk_rule's value_raw takes the same
// approach for the analogous problem of arbitrary ruleset values.
func (c *Client) CreateBIAggregation(ctx context.Context, aggregationID string, body map[string]interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/objects/bi_aggregation/%s", aggregationID)
	resp, err := c.request(ctx, "POST", path, body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetBIAggregation retrieves a BI aggregation by id.
func (c *Client) GetBIAggregation(ctx context.Context, aggregationID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/objects/bi_aggregation/%s", aggregationID)
	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if err := HandleNotFound(resp, "BI aggregation", aggregationID); err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateBIAggregation updates an existing BI aggregation. Unlike host/rule
// updates, CheckMK does not require an If-Match/ETag header for this
// endpoint.
func (c *Client) UpdateBIAggregation(ctx context.Context, aggregationID string, body map[string]interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/objects/bi_aggregation/%s", aggregationID)
	resp, err := c.request(ctx, "PUT", path, body)
	if err != nil {
		return nil, err
	}

	if err := HandleNotFound(resp, "BI aggregation", aggregationID); err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteBIAggregation deletes a BI aggregation by id.
func (c *Client) DeleteBIAggregation(ctx context.Context, aggregationID string) error {
	path := fmt.Sprintf("/objects/bi_aggregation/%s", aggregationID)
	resp, err := c.request(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	// 404 is acceptable for delete (idempotent)
	if IsNotFoundResponse(resp) {
		resp.Body.Close()
		return nil
	}

	return c.handleResponse(resp, nil)
}
