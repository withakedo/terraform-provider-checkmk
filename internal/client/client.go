package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Client is the CheckMK API client
type Client struct {
	BaseURL         *url.URL
	HTTPClient      *http.Client
	Username        string
	Password        string
	Version         *Version // CheckMK server version
	ProviderVersion string   // Terraform provider version (for User-Agent)
	MaxRetries      int      // Maximum number of retries (default 3)
	RetryWaitMin    time.Duration
	RetryWaitMax    time.Duration

	// LongOperationTimeout bounds requests to CheckMK's blocking "wait for
	// completion" endpoints (activation, service discovery), which can
	// legitimately run far longer than a normal CRUD call and must not be
	// cut short by the general-purpose request_timeout. Defaults to 30
	// minutes if unset.
	LongOperationTimeout time.Duration
}

// longPollClient returns an http.Client bounded by LongOperationTimeout
// instead of HTTPClient's normal request_timeout, sharing the same
// Transport (and therefore connection pool). Redirects are not
// auto-followed: CheckMK's wait-for-completion endpoints redirect to
// themselves while still running, and Go's default 10-redirect cap would
// otherwise cut off a long-running wait.
func (c *Client) longPollClient() *http.Client {
	timeout := c.LongOperationTimeout
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: c.HTTPClient.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// pollSelfRedirectingCompletion polls a CheckMK "wait for completion"-style
// endpoint that responds 302 (redirecting to itself) while the underlying
// operation is still running, and 204 once it's done - the pattern used by
// both activation and service discovery. It uses longPollClient rather than
// HTTPClient so a slow operation isn't cut short by request_timeout, and
// follows redirects manually so it isn't bound by Go's default 10-redirect
// cap either.
func (c *Client) pollSelfRedirectingCompletion(ctx context.Context, startURL string) error {
	client := c.longPollClient()
	url := startURL

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

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("wait for completion request failed: %w", err)
		}

		switch resp.StatusCode {
		case http.StatusNoContent:
			// Operation has completed.
			resp.Body.Close()
			return nil
		case http.StatusFound:
			// Still running; CheckMK redirects to itself to avoid timeouts.
			if location := resp.Header.Get("Location"); location != "" {
				url = location
			}
			resp.Body.Close()
			continue
		case http.StatusNotFound:
			// No running operation with this id - treat as already complete
			// (e.g. it finished and was cleaned up between trigger and the
			// first poll).
			resp.Body.Close()
			return nil
		default:
			return c.handleResponse(resp, nil)
		}
	}
}

// SetProviderVersion sets the provider version for User-Agent header
func (c *Client) SetProviderVersion(version string) {
	c.ProviderVersion = version
}

// ClientOptions contains optional settings for the CheckMK client
type ClientOptions struct {
	RequestTimeout     time.Duration
	MaxRetries         int
	InsecureSkipVerify bool

	// LongOperationTimeout bounds blocking "wait for completion" calls
	// (activation, service discovery) independently of RequestTimeout.
	// Defaults to 30 minutes if unset.
	LongOperationTimeout time.Duration
}

// NewClient creates a new CheckMK API client
func NewClient(baseURL, username, password string) (*Client, error) {
	return NewClientWithOptions(baseURL, username, password, nil)
}

// NewClientWithOptions creates a new CheckMK API client with custom options
func NewClientWithOptions(baseURL, username, password string, opts *ClientOptions) (*Client, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Set defaults
	timeout := 60 * time.Second
	maxRetries := 3
	insecureSkipVerify := false

	// Apply options if provided
	if opts != nil {
		if opts.RequestTimeout > 0 {
			timeout = opts.RequestTimeout
		}
		if opts.MaxRetries > 0 {
			maxRetries = opts.MaxRetries
		}
		insecureSkipVerify = opts.InsecureSkipVerify
	}

	// Create HTTP transport with TLS options. Terraform runs resource
	// operations concurrently (default parallelism of 10), all against the
	// same CheckMK host; the default transport only keeps 2 idle
	// connections per host, so without raising these limits most of that
	// concurrency was going to connection setup/teardown instead of reuse.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 25,
		IdleConnTimeout:     90 * time.Second,
	}

	longOperationTimeout := time.Duration(0)
	if opts != nil {
		longOperationTimeout = opts.LongOperationTimeout
	}

	client := &Client{
		BaseURL:  parsedURL,
		Username: username,
		Password: password,
		HTTPClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		MaxRetries:           maxRetries,
		RetryWaitMin:         1 * time.Second,
		RetryWaitMax:         30 * time.Second,
		LongOperationTimeout: longOperationTimeout,
	}

	// Detect version on initialization
	version, err := client.GetVersion(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to detect CheckMK version: %w", err)
	}
	client.Version = version

	return client, nil
}

// isRetryableStatusCode returns true if the HTTP status code should trigger a retry
func isRetryableStatusCode(statusCode int) bool {
	// Retry on rate limiting and server errors
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	}
	return false
}

// calculateBackoff calculates the backoff duration with jitter
func (c *Client) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: min * 2^attempt
	backoff := c.RetryWaitMin * time.Duration(1<<uint(attempt))
	if backoff > c.RetryWaitMax {
		backoff = c.RetryWaitMax
	}
	// Add jitter (random value between 0 and 25% of backoff)
	jitter := time.Duration(rand.Int63n(int64(backoff / 4)))
	return backoff + jitter
}

// request makes an HTTP request to the CheckMK API with automatic retry on transient failures
func (c *Client) request(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	// Marshal body once for potential retries
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	reqURL := fmt.Sprintf("%s/check_mk/api/1.0%s", c.BaseURL.String(), path)

	// User-Agent header for API debugging and analytics
	providerVersion := c.ProviderVersion
	if providerVersion == "" {
		providerVersion = "dev"
	}
	userAgent := fmt.Sprintf(
		"terraform-provider-checkmk/%s (+https://github.com/withakedo/terraform-provider-checkmk)",
		providerVersion,
	)

	var lastErr error

	maxRetries := c.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context before each attempt
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Create fresh reader for each attempt
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", userAgent)
		req.SetBasicAuth(c.Username, c.Password)

		// Debug logging for API requests
		tflog.Debug(ctx, "CheckMK API request", map[string]interface{}{
			"method":  method,
			"url":     reqURL,
			"attempt": attempt + 1,
		})

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			tflog.Warn(ctx, "CheckMK API request failed, will retry", map[string]interface{}{
				"method":  method,
				"url":     reqURL,
				"attempt": attempt + 1,
				"error":   err.Error(),
			})

			// Wait before retrying (unless this was the last attempt)
			if attempt < maxRetries {
				backoff := c.calculateBackoff(attempt)
				tflog.Debug(ctx, "Waiting before retry", map[string]interface{}{
					"backoff_ms": backoff.Milliseconds(),
				})
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoff):
				}
			}
			continue
		}

		tflog.Debug(ctx, "CheckMK API response", map[string]interface{}{
			"method":  method,
			"url":     reqURL,
			"status":  resp.StatusCode,
			"attempt": attempt + 1,
		})

		// Check if we should retry based on status code
		if isRetryableStatusCode(resp.StatusCode) && attempt < maxRetries {
			// Close the response body before retrying
			resp.Body.Close()

			tflog.Warn(ctx, "CheckMK API returned retryable status, will retry", map[string]interface{}{
				"method":  method,
				"url":     reqURL,
				"status":  resp.StatusCode,
				"attempt": attempt + 1,
			})

			backoff := c.calculateBackoff(attempt)
			tflog.Debug(ctx, "Waiting before retry", map[string]interface{}{
				"backoff_ms": backoff.Milliseconds(),
			})
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			continue
		}

		// Success or non-retryable error
		return resp, nil
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("request failed after %d retries", maxRetries+1)
}

// handleResponse processes the HTTP response and handles errors
func (c *Client) handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		var apiError APIError
		if err := json.Unmarshal(body, &apiError); err != nil {
			// If we can't parse the error, return raw response
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return &apiError
	}

	// Parse success response
	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
