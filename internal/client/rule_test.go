package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCreateRule(t *testing.T) {
	tests := []struct {
		name           string
		request        *RuleCreateRequest
		responseStatus int
		responseBody   string
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful creation",
			request: &RuleCreateRequest{
				Ruleset: "host_label_rules",
				Folder:  "/",
				Properties: RuleProperties{
					Description: "Test rule",
					Disabled:    false,
				},
				ValueRaw:   "{\"os\": \"linux\"}",
				Conditions: map[string]interface{}{},
			},
			responseStatus: http.StatusOK,
			responseBody: `{
				"id": "test-rule-id",
				"title": "Test rule",
				"domainType": "rule",
				"extensions": {
					"ruleset": "host_label_rules",
					"folder": "/",
					"properties": {
						"description": "Test rule",
						"disabled": false
					},
					"value_raw": "{'os': 'linux'}",
					"conditions": {}
				}
			}`,
			expectError: false,
		},
		{
			name: "conflict error",
			request: &RuleCreateRequest{
				Ruleset: "host_label_rules",
				Folder:  "/",
				Properties: RuleProperties{
					Description: "Duplicate rule",
					Disabled:    false,
				},
				ValueRaw:   "{}",
				Conditions: map[string]interface{}{},
			},
			responseStatus: http.StatusConflict,
			responseBody: `{
				"title": "Conflict",
				"status": 409,
				"detail": "Rule 'Duplicate rule' already exists"
			}`,
			expectError:   true,
			errorContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/check_mk/api/1.0/domain-types/rule/collections/all" {
					t.Errorf("Expected path /check_mk/api/1.0/domain-types/rule/collections/all, got %s", r.URL.Path)
				}

				// Verify request body
				var req RuleCreateRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Failed to decode request body: %v", err)
				}

				if req.Ruleset != tt.request.Ruleset {
					t.Errorf("Expected ruleset %s, got %s", tt.request.Ruleset, req.Ruleset)
				}

				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				HTTPClient: server.Client(),
				BaseURL:    mustParseURL(server.URL),
			}

			rule, err := client.CreateRule(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rule == nil {
					t.Error("Expected rule but got nil")
				} else {
					if rule.ID != "test-rule-id" {
						t.Errorf("Expected ID 'test-rule-id', got %s", rule.ID)
					}
					if rule.Extensions.Ruleset != "host_label_rules" {
						t.Errorf("Expected ruleset 'host_label_rules', got %s", rule.Extensions.Ruleset)
					}
				}
			}
		})
	}
}

func TestGetRule(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		responseStatus int
		responseBody   string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "successful get",
			ruleID:         "test-rule-id",
			responseStatus: http.StatusOK,
			responseBody: `{
				"id": "test-rule-id",
				"title": "Test rule",
				"domainType": "rule",
				"extensions": {
					"ruleset": "host_label_rules",
					"folder": "/",
					"properties": {
						"description": "Test rule",
						"disabled": false
					},
					"value_raw": "{'os': 'linux'}",
					"conditions": {}
				}
			}`,
			expectError: false,
		},
		{
			name:           "not found",
			ruleID:         "nonexistent-rule",
			responseStatus: http.StatusNotFound,
			responseBody: `{
				"title": "Not Found",
				"status": 404,
				"detail": "Rule 'nonexistent-rule' not found"
			}`,
			expectError:   true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				expectedPath := fmt.Sprintf("/check_mk/api/1.0/objects/rule/%s", tt.ruleID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Add ETag header
				w.Header().Set("ETag", `"test-etag"`)
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				HTTPClient: server.Client(),
				BaseURL:    mustParseURL(server.URL),
			}

			rule, err := client.GetRule(context.Background(), tt.ruleID)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rule == nil {
					t.Error("Expected rule but got nil")
				} else if rule.ID != tt.ruleID {
					t.Errorf("Expected ID %s, got %s", tt.ruleID, rule.ID)
				}
			}
		})
	}
}

func TestGetRuleWithETag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"test-etag-value"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "test-rule-id",
			"title": "Test rule",
			"domainType": "rule",
			"extensions": {
				"ruleset": "host_label_rules",
				"folder": "/",
				"properties": {
					"description": "Test rule",
					"disabled": false
				},
				"value_raw": "{'os': 'linux'}",
				"conditions": {}
			}
		}`))
	}))
	defer server.Close()

	client := &Client{
		HTTPClient: server.Client(),
		BaseURL:    mustParseURL(server.URL),
	}

	result, err := client.GetRuleWithETag(context.Background(), "test-rule-id")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.ETag != `"test-etag-value"` {
		t.Errorf("Expected ETag '\"test-etag-value\"', got %s", result.ETag)
	}
	if result.Rule.ID != "test-rule-id" {
		t.Errorf("Expected ID 'test-rule-id', got %s", result.Rule.ID)
	}
}

func TestGetRulesByRuleset(t *testing.T) {
	tests := []struct {
		name           string
		ruleset        string
		responseStatus int
		responseBody   string
		expectError    bool
		expectedCount  int
	}{
		{
			name:           "empty ruleset",
			ruleset:        "host_label_rules",
			responseStatus: http.StatusOK,
			responseBody: `{
				"id": "rule",
				"domainType": "rule",
				"value": [],
				"extensions": {
					"found_rules": 0
				}
			}`,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:           "ruleset with rules",
			ruleset:        "host_label_rules",
			responseStatus: http.StatusOK,
			responseBody: `{
				"id": "rule",
				"domainType": "rule",
				"value": [
					{
						"id": "rule-1",
						"title": "Rule 1",
						"domainType": "rule",
						"extensions": {
							"ruleset": "host_label_rules",
							"folder": "/",
							"properties": {
								"description": "Rule 1",
								"disabled": false
							},
							"value_raw": "{}",
							"conditions": {}
						}
					},
					{
						"id": "rule-2",
						"title": "Rule 2",
						"domainType": "rule",
						"extensions": {
							"ruleset": "host_label_rules",
							"folder": "/",
							"properties": {
								"description": "Rule 2",
								"disabled": false
							},
							"value_raw": "{}",
							"conditions": {}
						}
					}
				],
				"extensions": {
					"found_rules": 2
				}
			}`,
			expectError:   false,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				// Check query parameter
				rulesetParam := r.URL.Query().Get("ruleset_name")
				if rulesetParam != tt.ruleset {
					t.Errorf("Expected ruleset_name %s, got %s", tt.ruleset, rulesetParam)
				}

				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				HTTPClient: server.Client(),
				BaseURL:    mustParseURL(server.URL),
			}

			rules, err := client.GetRulesByRuleset(context.Background(), tt.ruleset)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(rules) != tt.expectedCount {
					t.Errorf("Expected %d rules, got %d", tt.expectedCount, len(rules))
				}
			}
		})
	}
}

func TestUpdateRule(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		request        *RuleUpdateRequest
		etag           string
		responseStatus int
		responseBody   string
		expectError    bool
		errorContains  string
	}{
		{
			name:   "successful update with ETag",
			ruleID: "test-rule-id",
			request: &RuleUpdateRequest{
				Properties: RuleProperties{
					Description: "Updated rule",
					Disabled:    false,
				},
				ValueRaw:   "{\"os\": \"windows\"}",
				Conditions: map[string]interface{}{},
			},
			etag:           `"valid-etag"`,
			responseStatus: http.StatusOK,
			responseBody: `{
				"id": "test-rule-id",
				"title": "Updated rule",
				"domainType": "rule",
				"extensions": {
					"ruleset": "host_label_rules",
					"folder": "/",
					"properties": {
						"description": "Updated rule",
						"disabled": false
					},
					"value_raw": "{'os': 'windows'}",
					"conditions": {}
				}
			}`,
			expectError: false,
		},
		{
			name:   "precondition failed",
			ruleID: "test-rule-id",
			request: &RuleUpdateRequest{
				Properties: RuleProperties{
					Description: "Updated rule",
					Disabled:    false,
				},
				ValueRaw:   "{}",
				Conditions: map[string]interface{}{},
			},
			etag:           `"invalid-etag"`,
			responseStatus: http.StatusPreconditionFailed,
			responseBody: `{
				"title": "Precondition Failed",
				"status": 412,
				"detail": "ETag mismatch"
			}`,
			expectError:   true,
			errorContains: "modified outside of Terraform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}

				// Check If-Match header
				ifMatch := r.Header.Get("If-Match")
				expectedIfMatch := tt.etag
				if expectedIfMatch == "" {
					expectedIfMatch = "*"
				}
				if ifMatch != expectedIfMatch {
					t.Errorf("Expected If-Match header %s, got %s", expectedIfMatch, ifMatch)
				}

				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				HTTPClient: server.Client(),
				BaseURL:    mustParseURL(server.URL),
			}

			rule, err := client.UpdateRule(context.Background(), tt.ruleID, tt.request, tt.etag)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rule == nil {
					t.Error("Expected rule but got nil")
				}
			}
		})
	}
}

func TestDeleteRule(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		etag           string
		responseStatus int
		expectError    bool
	}{
		{
			name:           "successful delete",
			ruleID:         "test-rule-id",
			etag:           `"valid-etag"`,
			responseStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:           "delete non-existent (idempotent)",
			ruleID:         "nonexistent-rule",
			etag:           "",
			responseStatus: http.StatusNotFound,
			expectError:    false,
		},
		{
			name:           "precondition failed",
			ruleID:         "test-rule-id",
			etag:           `"invalid-etag"`,
			responseStatus: http.StatusPreconditionFailed,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}

				w.WriteHeader(tt.responseStatus)
			}))
			defer server.Close()

			client := &Client{
				HTTPClient: server.Client(),
				BaseURL:    mustParseURL(server.URL),
			}

			err := client.DeleteRule(context.Background(), tt.ruleID, tt.etag)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper functions

func mustParseURL(urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}
	return u
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
