package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ActivationRequest represents a request to activate pending changes
type ActivationRequest struct {
	Sites         []string `json:"sites,omitempty"`
	ForceActivate bool     `json:"force_foreign_changes,omitempty"`
}

// ActivationResponse represents the response from activating changes
type ActivationResponse struct {
	ID         string                 `json:"id"`
	DomainType string                 `json:"domainType"`
	Title      string                 `json:"title"`
	Extensions map[string]interface{} `json:"extensions"`
}

// ActivateChanges activates pending configuration changes
// forceForeignChanges controls whether to activate changes made by other users/tools (request body field)
// strictResourceLocking controls ETag validation for both activation and resource operations
//   - true: Requires proper ETag for activation endpoint (must GET pending changes first)
//   - false: Uses If-Match: * to bypass ETag validation (default, matches Python)
//
// waitTime is the number of seconds to wait after activation for changes to propagate
func (c *Client) ActivateChanges(ctx context.Context, sites []string, forceForeignChanges bool, strictResourceLocking bool, waitTime int) error {
	req := &ActivationRequest{
		Sites:         sites,
		ForceActivate: forceForeignChanges,
	}

	// Build headers based on strictResourceLocking setting
	headers := make(map[string]string)
	if !strictResourceLocking {
		// Use If-Match: * to bypass ETag requirements (Python approach, default)
		headers["If-Match"] = "*"
	}
	// If strictResourceLocking is true, caller must provide ETag
	// (not implemented yet - would need to GET pending changes first)

	resp, err := c.requestWithHeaders(ctx, "POST", "/domain-types/activation_run/actions/activate-changes/invoke", req, headers)
	if err != nil {
		return err
	}

	// Handle 422 Unprocessable Entity (no changes to activate)
	if resp.StatusCode == http.StatusUnprocessableEntity {
		// No changes to activate is not an error
		return nil
	}

	// Just check that the request succeeded
	var activation ActivationResponse
	if err := c.handleResponse(resp, &activation); err != nil {
		return err
	}

	// Activation triggered successfully - CheckMK processes it asynchronously
	// Wait for changes to propagate
	time.Sleep(time.Duration(waitTime) * time.Second)

	return nil
}

// requestWithHeaders makes an HTTP request with custom headers
func (c *Client) requestWithHeaders(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	url := fmt.Sprintf("%s/check_mk/api/1.0%s", c.BaseURL.String(), path)
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set standard headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Basic auth
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}
