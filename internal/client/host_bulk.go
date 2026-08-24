package client

import "context"

// BulkCreateHostRequest creates many hosts in a single API call. Mirrors the
// "BulkCreateHost" schema from the CheckMK REST API.
type BulkCreateHostRequest struct {
	Entries []HostCreateRequest `json:"entries"`
}

// BulkUpdateHostEntry is a single host's update within a BulkUpdateHostRequest.
// Unlike the single-host UpdateHost endpoint (PUT /objects/host_config/{name}),
// the bulk endpoint identifies each entry by host_name in the body rather
// than in the URL. Mirrors the "UpdateHostEntry" schema.
type BulkUpdateHostEntry struct {
	HostName   string                 `json:"host_name"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// BulkUpdateHostRequest updates many hosts in a single API call. Mirrors the
// "BulkUpdateHost" schema.
type BulkUpdateHostRequest struct {
	Entries []BulkUpdateHostEntry `json:"entries"`
}

// BulkDeleteHostRequest deletes many hosts in a single API call, identified
// by name. Mirrors the "BulkDeleteHost" schema.
type BulkDeleteHostRequest struct {
	Entries []string `json:"entries"`
}

// HostConfigCollection is the response envelope for bulk-create and
// bulk-update, containing the resulting host objects.
type HostConfigCollection struct {
	Value []Host `json:"value"`
}

// BulkCreateHosts creates many hosts in a single API call. CheckMK's bulk
// endpoints are not transactional: if any entry fails validation, the whole
// call responds with an error (400) that lists which host names failed and
// which succeeded (in the error detail) - already-created hosts from a
// partially-successful call are NOT rolled back server-side. Callers should
// treat a non-nil error here as "state unknown, reconcile manually" rather
// than "nothing happened".
func (c *Client) BulkCreateHosts(ctx context.Context, req *BulkCreateHostRequest) (*HostConfigCollection, error) {
	resp, err := c.request(ctx, "POST", "/domain-types/host_config/actions/bulk-create/invoke", req)
	if err != nil {
		return nil, err
	}

	var result HostConfigCollection
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BulkUpdateHosts updates many hosts in a single API call. See
// BulkCreateHosts for the non-transactional failure behavior.
func (c *Client) BulkUpdateHosts(ctx context.Context, req *BulkUpdateHostRequest) (*HostConfigCollection, error) {
	resp, err := c.request(ctx, "PUT", "/domain-types/host_config/actions/bulk-update/invoke", req)
	if err != nil {
		return nil, err
	}

	var result HostConfigCollection
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BulkDeleteHosts deletes many hosts, identified by name, in a single API
// call.
func (c *Client) BulkDeleteHosts(ctx context.Context, hostNames []string) error {
	req := &BulkDeleteHostRequest{Entries: hostNames}

	resp, err := c.request(ctx, "POST", "/domain-types/host_config/actions/bulk-delete/invoke", req)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}
