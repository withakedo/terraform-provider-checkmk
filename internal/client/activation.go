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
//   - true: fetches the current pending-changes ETag and sends it as If-Match,
//     so the activation fails with a precondition error (rather than silently
//     activating) if the set of pending changes shifted between planning and
//     applying.
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
//
// Calls are serialized within this process via activationMu (see its
// docstring) so that Terraform's concurrent resource operations under
// activate = "auto" don't collide on CheckMK's own activation lock. If
// CheckMK still reports 423 Locked - e.g. a human activated changes in the
// UI, or another terraform apply is running against the same site - this
// retries with backoff a few times before giving up, since that lock is
// typically held only for the duration of one activation cycle.
func (c *Client) ActivateChanges(ctx context.Context, sites []string, forceForeignChanges bool, strictResourceLocking bool, waitTime int) error {
	c.activationMu.Lock()
	defer c.activationMu.Unlock()

	const maxLockedRetries = 5
	lockedWait := 2 * time.Second

	for attempt := 0; ; attempt++ {
		activated, err := c.triggerAndWaitForActivation(ctx, sites, forceForeignChanges, strictResourceLocking)
		if err == nil {
			if activated && waitTime > 0 {
				// Extra margin for eventual consistency in other API reads,
				// on top of activation genuinely having completed.
				select {
				case <-time.After(time.Duration(waitTime) * time.Second):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		}

		apiErr, ok := err.(*APIError)
		if !ok || apiErr.Status != http.StatusLocked || attempt >= maxLockedRetries {
			return err
		}

		// Another activation (outside this process, or outside our lock -
		// e.g. triggered manually in the CheckMK UI) is still running.
		select {
		case <-time.After(lockedWait):
		case <-ctx.Done():
			return ctx.Err()
		}
		if lockedWait < 30*time.Second {
			lockedWait *= 2
		}
	}
}

// triggerAndWaitForActivation performs a single activate-changes attempt:
// trigger, then poll for completion. Returns activated=true if it actually
// triggered and waited out an activation (as opposed to finding nothing
// pending).
func (c *Client) triggerAndWaitForActivation(ctx context.Context, sites []string, forceForeignChanges bool, strictResourceLocking bool) (bool, error) {
	req := &ActivationRequest{
		Sites:         sites,
		ForceActivate: forceForeignChanges,
	}

	// Build headers based on strictResourceLocking setting
	etag := ""
	if strictResourceLocking {
		var err error
		etag, err = c.getPendingChangesETag(ctx)
		if err != nil {
			return false, fmt.Errorf("failed to fetch pending changes ETag for strict activation locking: %w", err)
		}
	}
	headers := BuildETagHeaders(etag)

	resp, err := c.requestWithHeaders(ctx, "POST", "/domain-types/activation_run/actions/activate-changes/invoke", req, headers)
	if err != nil {
		return false, err
	}

	// Handle 422 Unprocessable Entity (no changes to activate)
	if resp.StatusCode == http.StatusUnprocessableEntity {
		resp.Body.Close()
		// No changes to activate is not an error - most likely another
		// goroutine in this process already activated them while this one
		// waited for the lock.
		return false, nil
	}

	// Under strict locking, a 412 means the set of pending changes shifted
	// between fetching the ETag and activating (e.g. someone else made a
	// change concurrently) - surface that as drift rather than a generic
	// API error.
	if driftErr := HandlePreconditionFailed(resp, "Activation", "pending changes"); driftErr != nil {
		resp.Body.Close()
		return false, driftErr
	}

	// Just check that the request succeeded (this also surfaces 423 Locked
	// as an *APIError, which ActivateChanges retries on).
	var activation ActivationResponse
	if err := c.handleResponse(resp, &activation); err != nil {
		return false, err
	}

	// Activation triggered successfully - CheckMK processes it asynchronously.
	// Poll until it actually completes instead of guessing a fixed sleep.
	if err := c.waitForActivationCompletion(ctx, activation.ID); err != nil {
		return false, err
	}

	return true, nil
}

// getPendingChangesETag fetches the ETag identifying the current set of
// pending changes, for use as the If-Match header on activate-changes under
// strict resource locking. CheckMK validates that header against the live
// set of pending changes, so this must be fetched immediately before
// activating rather than cached.
func (c *Client) getPendingChangesETag(ctx context.Context) (string, error) {
	resp, err := c.request(ctx, "GET", "/domain-types/activation_run/collections/pending_changes", nil)
	if err != nil {
		return "", err
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return "", err
	}

	etag := resp.Header.Get("ETag")
	if etag == "" {
		return "", fmt.Errorf("CheckMK did not return an ETag for pending changes")
	}
	return etag, nil
}

// waitForActivationCompletion polls CheckMK's "wait for completion" endpoint
// for the given activation run until the activation genuinely finishes. See
// pollSelfRedirectingCompletion for how the endpoint's self-redirecting
// long-poll pattern is handled, and why it's bound by LongOperationTimeout
// rather than the general request_timeout.
func (c *Client) waitForActivationCompletion(ctx context.Context, activationID string) error {
	if activationID == "" {
		return nil
	}

	url := fmt.Sprintf("%s/check_mk/api/1.0/objects/activation_run/%s/actions/wait-for-completion/invoke", c.BaseURL.String(), activationID)
	return c.pollSelfRedirectingCompletion(ctx, url)
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
