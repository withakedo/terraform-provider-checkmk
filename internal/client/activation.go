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
// waitTime is the number of seconds to wait, on top of the actual activation
// completing, for changes to finish propagating (e.g. eventual-consistency
// margin for other API reads). The activation itself is no longer covered by
// this fixed wait: instead, ActivateChanges polls CheckMK's "wait for
// completion" endpoint so it returns as soon as activation genuinely
// finishes, rather than guessing a fixed sleep duration that may be too
// short (stale reads) or too long (slow applies) for the actual activation
// time.
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

	// Activation triggered successfully - CheckMK processes it asynchronously.
	// Poll until it actually completes instead of guessing a fixed sleep.
	if err := c.waitForActivationCompletion(ctx, activation.ID); err != nil {
		return err
	}

	// Extra margin for eventual consistency in other API reads, on top of
	// activation genuinely having completed.
	if waitTime > 0 {
		select {
		case <-time.After(time.Duration(waitTime) * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// waitForActivationCompletion polls CheckMK's "wait for completion" endpoint
// for the given activation run. That endpoint responds 302, redirecting to
// itself, for as long as the activation is still running (CheckMK paces
// these redirects itself to avoid client-side timeouts), and 204 once the
// activation has finished. Redirects are followed manually (rather than via
// the shared HTTPClient's default redirect-following) so each hop can be
// distinguished from genuine completion.
func (c *Client) waitForActivationCompletion(ctx context.Context, activationID string) error {
	if activationID == "" {
		return nil
	}

	url := fmt.Sprintf("%s/check_mk/api/1.0/objects/activation_run/%s/actions/wait-for-completion/invoke", c.BaseURL.String(), activationID)

	noRedirectClient := &http.Client{
		Timeout:   c.HTTPClient.Timeout,
		Transport: c.HTTPClient.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.SetBasicAuth(c.Username, c.Password)

		resp, err := noRedirectClient.Do(req)
		if err != nil {
			return fmt.Errorf("wait for activation completion request failed: %w", err)
		}

		switch resp.StatusCode {
		case http.StatusNoContent:
			// Activation has completed.
			resp.Body.Close()
			return nil
		case http.StatusFound:
			// Still running; CheckMK redirects to itself to avoid timeouts.
			// Follow the redirect target if provided, otherwise re-poll the
			// same URL.
			if location := resp.Header.Get("Location"); location != "" {
				url = location
			}
			resp.Body.Close()
			continue
		case http.StatusNotFound:
			// No running activation with this id - treat as already
			// complete (e.g. it finished and was cleaned up between our
			// trigger and the first poll).
			resp.Body.Close()
			return nil
		default:
			return c.handleResponse(resp, nil)
		}
	}
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
