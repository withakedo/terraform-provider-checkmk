package client

import (
	"context"
	"fmt"
)

// CreateSiteConnection creates a site connection for distributed
// monitoring. body must contain a "site_config" key matching CheckMK's
// "SiteConnectionCreate" schema (basic_settings, configuration_connection,
// status_connection sub-objects). This is deliberately untyped: that schema
// is large, deeply nested, and varies across CheckMK versions/editions -
// the same design choice checkmk_rule and checkmk_bi_aggregation make for
// similarly complex, highly variable bodies.
func (c *Client) CreateSiteConnection(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	resp, err := c.request(ctx, "POST", "/domain-types/site_connection/collections/all", body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetSiteConnection retrieves a site connection by site id.
func (c *Client) GetSiteConnection(ctx context.Context, siteID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/objects/site_connection/%s", siteID)
	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if err := HandleNotFound(resp, "Site connection", siteID); err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateSiteConnection updates an existing site connection. body must
// contain a "site_config" key. No If-Match/ETag header is required by this
// endpoint.
func (c *Client) UpdateSiteConnection(ctx context.Context, siteID string, body map[string]interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/objects/site_connection/%s", siteID)
	resp, err := c.request(ctx, "PUT", path, body)
	if err != nil {
		return nil, err
	}

	if err := HandleNotFound(resp, "Site connection", siteID); err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteSiteConnection deletes a site connection by site id.
func (c *Client) DeleteSiteConnection(ctx context.Context, siteID string) error {
	path := fmt.Sprintf("/objects/site_connection/%s/actions/delete/invoke", siteID)
	resp, err := c.request(ctx, "POST", path, nil)
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
