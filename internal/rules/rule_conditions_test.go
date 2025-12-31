package rules

import (
	"context"
	"strings"
	"testing"
)

// TestValidateRulesetConditions_LabelRulesets tests that label conditions
// are properly rejected on label rulesets (circular dependency prevention)
func TestValidateRulesetConditions_LabelRulesets(t *testing.T) {
	tests := []struct {
		name       string
		ruleset    string
		conditions map[string]interface{}
		wantErrors int
		errorMatch string
	}{
		{
			name:    "host_label_rules with host_labels condition - should error",
			ruleset: "host_label_rules",
			conditions: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
			},
			wantErrors: 1,
			errorMatch: "circular dependency",
		},
		{
			name:    "host_label_rules with host_label_groups condition - should error",
			ruleset: "host_label_rules",
			conditions: map[string]interface{}{
				"host_label_groups": []interface{}{"prod_servers", "critical_systems"},
			},
			wantErrors: 1,
			errorMatch: "circular dependency",
		},
		{
			name:    "host_label_rules with both host_labels and host_label_groups - should error",
			ruleset: "host_label_rules",
			conditions: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
				"host_label_groups": []interface{}{"prod_servers"},
			},
			wantErrors: 1,
			errorMatch: "circular dependency",
		},
		{
			name:    "service_label_rules with service_labels condition - should error",
			ruleset: "service_label_rules",
			conditions: map[string]interface{}{
				"service_labels": []interface{}{
					map[string]interface{}{
						"key":      "tier",
						"operator": "is",
						"value":    "frontend",
					},
				},
			},
			wantErrors: 1,
			errorMatch: "circular dependency",
		},
		{
			name:    "service_label_rules with service_label_groups condition - should error",
			ruleset: "service_label_rules",
			conditions: map[string]interface{}{
				"service_label_groups": []interface{}{"web_services"},
			},
			wantErrors: 1,
			errorMatch: "circular dependency",
		},
		{
			name:    "host_label_rules with service labels - should allow (not circular)",
			ruleset: "host_label_rules",
			conditions: map[string]interface{}{
				"service_labels": []interface{}{
					map[string]interface{}{
						"key":      "tier",
						"operator": "is",
						"value":    "frontend",
					},
				},
			},
			wantErrors: 0,
		},
		{
			name:    "service_label_rules with host labels - should allow (not circular)",
			ruleset: "service_label_rules",
			conditions: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
			},
			wantErrors: 0,
		},
		{
			name:    "label ruleset with non-label conditions - should allow",
			ruleset: "host_label_rules",
			conditions: map[string]interface{}{
				"host_name": map[string]interface{}{
					"match_on": []interface{}{"web-*", "app-*"},
					"operator": "one_of",
				},
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "criticality",
						"operator": "is",
						"value":    "prod",
					},
				},
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateRulesetConditions(tt.ruleset, tt.conditions)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateRulesetConditions() returned %d errors, want %d errors: %v",
					len(errors), tt.wantErrors, errors)
				return
			}

			if tt.wantErrors > 0 && tt.errorMatch != "" {
				found := false
				for _, err := range errors {
					if strings.Contains(err, tt.errorMatch) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ValidateRulesetConditions() errors don't contain %q, got: %v",
						tt.errorMatch, errors)
				}
			}
		})
	}
}

// TestValidateRulesetConditions_NormalRulesets tests that normal rulesets
// accept all condition types including labels
func TestValidateRulesetConditions_NormalRulesets(t *testing.T) {
	tests := []struct {
		name       string
		ruleset    string
		conditions map[string]interface{}
	}{
		{
			name:    "extra_host_conf with host_labels - should allow",
			ruleset: "extra_host_conf:check_interval",
			conditions: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
			},
		},
		{
			name:    "extra_service_conf with service_labels - should allow",
			ruleset: "extra_service_conf:check_interval",
			conditions: map[string]interface{}{
				"service_labels": []interface{}{
					map[string]interface{}{
						"key":      "tier",
						"operator": "is",
						"value":    "frontend",
					},
				},
			},
		},
		{
			name:    "custom_checks with all label conditions - should allow",
			ruleset: "custom_checks",
			conditions: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
				"host_label_groups": []interface{}{"prod_servers"},
				"service_labels": []interface{}{
					map[string]interface{}{
						"key":      "tier",
						"operator": "is",
						"value":    "frontend",
					},
				},
				"service_label_groups": []interface{}{"web_services"},
			},
		},
		{
			name:    "notification_rule with labels - should allow",
			ruleset: "notification_rule",
			conditions: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateRulesetConditions(tt.ruleset, tt.conditions)

			if len(errors) != 0 {
				t.Errorf("ValidateRulesetConditions() returned unexpected errors for normal ruleset: %v", errors)
			}
		})
	}
}

// TestIsLabelRuleset tests the label ruleset detection
func TestIsLabelRuleset(t *testing.T) {
	tests := []struct {
		name    string
		ruleset string
		want    bool
	}{
		{
			name:    "host_label_rules - is label ruleset",
			ruleset: "host_label_rules",
			want:    true,
		},
		{
			name:    "service_label_rules - is label ruleset",
			ruleset: "service_label_rules",
			want:    true,
		},
		{
			name:    "extra_host_conf:check_interval - not label ruleset",
			ruleset: "extra_host_conf:check_interval",
			want:    false,
		},
		{
			name:    "custom_checks - not label ruleset",
			ruleset: "custom_checks",
			want:    false,
		},
		{
			name:    "notification_rule - not label ruleset",
			ruleset: "notification_rule",
			want:    false,
		},
		{
			name:    "empty string - not label ruleset",
			ruleset: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLabelRuleset(tt.ruleset)
			if got != tt.want {
				t.Errorf("isLabelRuleset(%q) = %v, want %v", tt.ruleset, got, tt.want)
			}
		})
	}
}

// TestHasHostLabelConditions tests detection of host label conditions
func TestHasHostLabelConditions(t *testing.T) {
	tests := []struct {
		name       string
		conditions map[string]interface{}
		want       bool
	}{
		{
			name: "has host_labels as slice - should detect",
			conditions: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
			},
			want: true,
		},
		{
			name: "has host_labels as map - should detect",
			conditions: map[string]interface{}{
				"host_labels": map[string]interface{}{
					"environment": "production",
				},
			},
			want: true,
		},
		{
			name: "has host_label_groups as slice - should detect",
			conditions: map[string]interface{}{
				"host_label_groups": []interface{}{"prod_servers", "critical_systems"},
			},
			want: true,
		},
		{
			name: "has host_label_groups as map - should detect",
			conditions: map[string]interface{}{
				"host_label_groups": map[string]interface{}{
					"group1": "value1",
				},
			},
			want: true,
		},
		{
			name: "has both host_labels and host_label_groups - should detect",
			conditions: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
				"host_label_groups": []interface{}{"prod_servers"},
			},
			want: true,
		},
		{
			name: "has empty host_labels slice - should not detect",
			conditions: map[string]interface{}{
				"host_labels": []interface{}{},
			},
			want: false,
		},
		{
			name: "has empty host_labels map - should not detect",
			conditions: map[string]interface{}{
				"host_labels": map[string]interface{}{},
			},
			want: false,
		},
		{
			name: "has only service labels - should not detect",
			conditions: map[string]interface{}{
				"service_labels": []interface{}{
					map[string]interface{}{
						"key":      "tier",
						"operator": "is",
						"value":    "frontend",
					},
				},
			},
			want: false,
		},
		{
			name:       "nil conditions - should not detect",
			conditions: nil,
			want:       false,
		},
		{
			name:       "empty conditions - should not detect",
			conditions: map[string]interface{}{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasHostLabelConditions(tt.conditions)
			if got != tt.want {
				t.Errorf("hasHostLabelConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHasServiceLabelConditions tests detection of service label conditions
func TestHasServiceLabelConditions(t *testing.T) {
	tests := []struct {
		name       string
		conditions map[string]interface{}
		want       bool
	}{
		{
			name: "has service_labels as slice - should detect",
			conditions: map[string]interface{}{
				"service_labels": []interface{}{
					map[string]interface{}{
						"key":      "tier",
						"operator": "is",
						"value":    "frontend",
					},
				},
			},
			want: true,
		},
		{
			name: "has service_labels as map - should detect",
			conditions: map[string]interface{}{
				"service_labels": map[string]interface{}{
					"tier": "frontend",
				},
			},
			want: true,
		},
		{
			name: "has service_label_groups as slice - should detect",
			conditions: map[string]interface{}{
				"service_label_groups": []interface{}{"web_services", "api_services"},
			},
			want: true,
		},
		{
			name: "has service_label_groups as map - should detect",
			conditions: map[string]interface{}{
				"service_label_groups": map[string]interface{}{
					"group1": "value1",
				},
			},
			want: true,
		},
		{
			name: "has both service_labels and service_label_groups - should detect",
			conditions: map[string]interface{}{
				"service_labels": []interface{}{
					map[string]interface{}{
						"key":      "tier",
						"operator": "is",
						"value":    "frontend",
					},
				},
				"service_label_groups": []interface{}{"web_services"},
			},
			want: true,
		},
		{
			name: "has empty service_labels slice - should not detect",
			conditions: map[string]interface{}{
				"service_labels": []interface{}{},
			},
			want: false,
		},
		{
			name: "has empty service_labels map - should not detect",
			conditions: map[string]interface{}{
				"service_labels": map[string]interface{}{},
			},
			want: false,
		},
		{
			name: "has only host labels - should not detect",
			conditions: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
			},
			want: false,
		},
		{
			name:       "nil conditions - should not detect",
			conditions: nil,
			want:       false,
		},
		{
			name:       "empty conditions - should not detect",
			conditions: map[string]interface{}{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasServiceLabelConditions(tt.conditions)
			if got != tt.want {
				t.Errorf("hasServiceLabelConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidateRulesetConditions_EdgeCases tests edge cases and boundary conditions
func TestValidateRulesetConditions_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		ruleset    string
		conditions map[string]interface{}
		wantErrors int
	}{
		{
			name:       "nil conditions - should not error",
			ruleset:    "host_label_rules",
			conditions: nil,
			wantErrors: 0,
		},
		{
			name:       "empty conditions - should not error",
			ruleset:    "host_label_rules",
			conditions: map[string]interface{}{},
			wantErrors: 0,
		},
		{
			name:    "label ruleset with empty label arrays - should not error",
			ruleset: "host_label_rules",
			conditions: map[string]interface{}{
				"host_labels":       []interface{}{},
				"host_label_groups": []interface{}{},
			},
			wantErrors: 0,
		},
		{
			name:    "label ruleset with other conditions - should not error",
			ruleset: "service_label_rules",
			conditions: map[string]interface{}{
				"host_name": map[string]interface{}{
					"match_on": []interface{}{"web-*"},
					"operator": "one_of",
				},
				"service_description": map[string]interface{}{
					"match_on": []interface{}{"CPU.*"},
					"operator": "one_of",
				},
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateRulesetConditions(tt.ruleset, tt.conditions)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateRulesetConditions() returned %d errors, want %d errors: %v",
					len(errors), tt.wantErrors, errors)
			}
		})
	}
}

// TestRuleConditionsSchema tests that the schema is properly constructed
func TestRuleConditionsSchema(t *testing.T) {
	schema := RuleConditionsSchema()

	if schema.MarkdownDescription == "" {
		t.Error("ConditionsSchema() missing description")
	}

	if !schema.Optional {
		t.Error("ConditionsSchema() should be optional")
	}

	// Check that key condition attributes exist
	requiredAttrs := []string{
		"host_name",
		"host_tags",
		"host_labels",
		"host_label_groups",
		"service_description",
		"service_labels",
		"service_label_groups",
	}

	for _, attrName := range requiredAttrs {
		if _, ok := schema.Attributes[attrName]; !ok {
			t.Errorf("ConditionsSchema() missing attribute %q", attrName)
		}
	}
}

// TestConvertConditionsFromClient_Empty tests conversion of empty conditions
func TestConvertConditionsFromClient_Empty(t *testing.T) {
	tests := []struct {
		name       string
		conditions map[string]interface{}
	}{
		{
			name:       "nil conditions",
			conditions: nil,
		},
		{
			name:       "empty map",
			conditions: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertConditionsFromClient(context.TODO(), tt.conditions)
			if err != nil {
				t.Errorf("ConvertConditionsFromClient() error = %v", err)
				return
			}

			if !result.IsNull() {
				t.Error("ConvertConditionsFromClient() expected null object for empty conditions")
			}
		})
	}
}

// TestConvertConditionsFromClient_HostName tests conversion of host_name conditions
func TestConvertConditionsFromClient_HostName(t *testing.T) {
	conditions := map[string]interface{}{
		"host_name": map[string]interface{}{
			"match_on": []interface{}{"web-*", "app-*"},
			"operator": "one_of",
		},
	}

	result, err := ConvertConditionsFromClient(context.TODO(), conditions)
	if err != nil {
		t.Fatalf("ConvertConditionsFromClient() error = %v", err)
	}

	if result.IsNull() {
		t.Fatal("ConvertConditionsFromClient() returned null, expected object")
	}

	// Verify the result has the host_name attribute
	attrs := result.Attributes()
	hostNameAttr, ok := attrs["host_name"]
	if !ok {
		t.Fatal("ConvertConditionsFromClient() missing host_name attribute")
	}

	if hostNameAttr.IsNull() {
		t.Fatal("ConvertConditionsFromClient() host_name should not be null")
	}
}

// TestConvertConditionsFromClient_HostTags tests conversion of host_tags conditions
func TestConvertConditionsFromClient_HostTags(t *testing.T) {
	conditions := map[string]interface{}{
		"host_tags": []interface{}{
			map[string]interface{}{
				"key":      "criticality",
				"operator": "is",
				"value":    "prod",
			},
			map[string]interface{}{
				"key":      "site",
				"operator": "is_not",
				"value":    "dr",
			},
		},
	}

	result, err := ConvertConditionsFromClient(context.TODO(), conditions)
	if err != nil {
		t.Fatalf("ConvertConditionsFromClient() error = %v", err)
	}

	if result.IsNull() {
		t.Fatal("ConvertConditionsFromClient() returned null, expected object")
	}

	attrs := result.Attributes()
	hostTagsAttr, ok := attrs["host_tags"]
	if !ok {
		t.Fatal("ConvertConditionsFromClient() missing host_tags attribute")
	}

	if hostTagsAttr.IsNull() {
		t.Fatal("ConvertConditionsFromClient() host_tags should not be null")
	}
}

// TestConvertConditionsFromClient_AllConditions tests conversion with all condition types
func TestConvertConditionsFromClient_AllConditions(t *testing.T) {
	conditions := map[string]interface{}{
		"host_name": map[string]interface{}{
			"match_on": []interface{}{"web-*"},
			"operator": "one_of",
		},
		"host_tags": []interface{}{
			map[string]interface{}{
				"key":      "criticality",
				"operator": "is",
				"value":    "prod",
			},
		},
		"host_labels": []interface{}{
			map[string]interface{}{
				"key":      "env",
				"operator": "is",
				"value":    "production",
			},
		},
		"host_label_groups": []interface{}{"prod_servers"},
		"service_description": map[string]interface{}{
			"match_on": []interface{}{"CPU*", "Memory"},
			"operator": "one_of",
		},
		"service_labels": []interface{}{
			map[string]interface{}{
				"key":      "tier",
				"operator": "is",
				"value":    "frontend",
			},
		},
		"service_label_groups": []interface{}{"web_services"},
	}

	result, err := ConvertConditionsFromClient(context.TODO(), conditions)
	if err != nil {
		t.Fatalf("ConvertConditionsFromClient() error = %v", err)
	}

	if result.IsNull() {
		t.Fatal("ConvertConditionsFromClient() returned null, expected object")
	}

	attrs := result.Attributes()

	// Check all attributes are present and not null
	checkAttrs := []string{
		"host_name",
		"host_tags",
		"host_labels",
		"host_label_groups",
		"service_description",
		"service_labels",
		"service_label_groups",
	}

	for _, attrName := range checkAttrs {
		attr, ok := attrs[attrName]
		if !ok {
			t.Errorf("ConvertConditionsFromClient() missing attribute %q", attrName)
			continue
		}
		if attr.IsNull() {
			t.Errorf("ConvertConditionsFromClient() attribute %q should not be null", attrName)
		}
	}
}

// TestConvertConditionsRoundTrip tests that ToClient -> FromClient preserves data
func TestConvertConditionsRoundTrip(t *testing.T) {
	// Start with API format conditions
	apiConditions := map[string]interface{}{
		"host_name": map[string]interface{}{
			"match_on": []interface{}{"web-*", "app-*"},
			"operator": "one_of",
		},
		"host_tags": []interface{}{
			map[string]interface{}{
				"key":      "criticality",
				"operator": "is",
				"value":    "prod",
			},
		},
	}

	// Convert from API to Terraform types
	tfConditions, err := ConvertConditionsFromClient(context.TODO(), apiConditions)
	if err != nil {
		t.Fatalf("ConvertConditionsFromClient() error = %v", err)
	}

	// Convert back from Terraform types to API format
	resultConditions, err := ConvertConditionsToClient(context.TODO(), tfConditions)
	if err != nil {
		t.Fatalf("ConvertConditionsToClient() error = %v", err)
	}

	// Verify host_name was preserved
	hostName, ok := resultConditions["host_name"].(map[string]interface{})
	if !ok {
		t.Fatal("Round trip lost host_name condition")
	}

	matchOn, ok := hostName["match_on"].([]string)
	if !ok || len(matchOn) != 2 {
		t.Errorf("Round trip changed host_name.match_on, got %v", hostName["match_on"])
	}

	operator, ok := hostName["operator"].(string)
	if !ok || operator != "one_of" {
		t.Errorf("Round trip changed host_name.operator, got %v", hostName["operator"])
	}

	// Verify host_tags was preserved
	hostTags, ok := resultConditions["host_tags"].([]map[string]interface{})
	if !ok || len(hostTags) != 1 {
		t.Errorf("Round trip changed host_tags, got %v", resultConditions["host_tags"])
	}
}
