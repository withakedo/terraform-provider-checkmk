package client

import (
	"testing"
)

func TestGenerateRuleHash(t *testing.T) {
	tests := []struct {
		name        string
		ruleset     string
		description string
		conditions  interface{}
		wantSame    bool
		otherHash   string
	}{
		{
			name:        "simple rule with no conditions",
			ruleset:     "host_tags",
			description: "Test rule",
			conditions:  nil,
		},
		{
			name:        "rule with host_name condition",
			ruleset:     "host_tags",
			description: "Test host name rule",
			conditions: map[string]interface{}{
				"host_name": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"host1", "host2", "host3"},
				},
			},
		},
		{
			name:        "rule with host_tags",
			ruleset:     "inventory_df_rules",
			description: "Filesystem discovery",
			conditions: map[string]interface{}{
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "criticality",
						"operator": "is",
						"value":    "prod",
					},
					map[string]interface{}{
						"key":      "location",
						"operator": "is",
						"value":    "datacenter1",
					},
				},
			},
		},
		{
			name:        "rule with host_labels",
			ruleset:     "checkgroup_parameters:filesystem",
			description: "Filesystem thresholds",
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
			name:        "rule with service_description",
			ruleset:     "checkgroup_parameters:filesystem",
			description: "Filesystem patterns",
			conditions: map[string]interface{}{
				"service_description": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"/var", "/tmp", "/opt"},
				},
			},
		},
		{
			name:        "complex rule with multiple conditions",
			ruleset:     "checkgroup_parameters:filesystem",
			description: "Complex filesystem rule",
			conditions: map[string]interface{}{
				"host_name": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"server1", "server2"},
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
						"key":      "app",
						"operator": "is",
						"value":    "database",
					},
				},
				"service_description": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"/data"},
				},
			},
		},
		{
			name:        "rule with label groups",
			ruleset:     "checkgroup_parameters:filesystem",
			description: "Label group rule",
			conditions: map[string]interface{}{
				"host_label_groups": []interface{}{
					map[string]interface{}{
						"operator": "and",
						"label_group": []interface{}{
							map[string]interface{}{
								"key":      "tier",
								"operator": "is",
								"value":    "backend",
							},
							map[string]interface{}{
								"key":      "region",
								"operator": "is",
								"value":    "us-east",
							},
						},
					},
				},
			},
		},
		{
			name:        "rule with service label groups",
			ruleset:     "checkgroup_parameters:filesystem",
			description: "Service label group rule",
			conditions: map[string]interface{}{
				"service_label_groups": []interface{}{
					map[string]interface{}{
						"operator": "or",
						"label_group": []interface{}{
							map[string]interface{}{
								"key":      "type",
								"operator": "is",
								"value":    "database",
							},
							map[string]interface{}{
								"key":      "type",
								"operator": "is",
								"value":    "cache",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := GenerateRuleHash(tt.ruleset, tt.description, tt.conditions)

			// Verify hash is 32 hex characters (16 bytes)
			if len(hash) != 32 {
				t.Errorf("hash length = %d, want 32", len(hash))
			}

			// Verify hash contains only hex characters
			for _, c := range hash {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("hash contains non-hex character: %c", c)
				}
			}

			// Generate again to ensure deterministic
			hash2 := GenerateRuleHash(tt.ruleset, tt.description, tt.conditions)
			if hash != hash2 {
				t.Errorf("hash is not deterministic: %s != %s", hash, hash2)
			}
		})
	}
}

func TestGenerateRuleHash_OrderIndependence(t *testing.T) {
	tests := []struct {
		name        string
		ruleset     string
		description string
		conditions1 interface{}
		conditions2 interface{}
		shouldMatch bool
	}{
		{
			name:        "host_name match_on in different order",
			ruleset:     "host_tags",
			description: "Test rule",
			conditions1: map[string]interface{}{
				"host_name": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"host1", "host2", "host3"},
				},
			},
			conditions2: map[string]interface{}{
				"host_name": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"host3", "host1", "host2"},
				},
			},
			shouldMatch: true,
		},
		{
			name:        "host_tags in different order",
			ruleset:     "inventory_df_rules",
			description: "Filesystem discovery",
			conditions1: map[string]interface{}{
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "location",
						"operator": "is",
						"value":    "datacenter1",
					},
					map[string]interface{}{
						"key":      "criticality",
						"operator": "is",
						"value":    "prod",
					},
				},
			},
			conditions2: map[string]interface{}{
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "criticality",
						"operator": "is",
						"value":    "prod",
					},
					map[string]interface{}{
						"key":      "location",
						"operator": "is",
						"value":    "datacenter1",
					},
				},
			},
			shouldMatch: true,
		},
		{
			name:        "host_labels in different order",
			ruleset:     "checkgroup_parameters:filesystem",
			description: "Filesystem thresholds",
			conditions1: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "tier",
						"operator": "is",
						"value":    "backend",
					},
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
			},
			conditions2: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
					map[string]interface{}{
						"key":      "tier",
						"operator": "is",
						"value":    "backend",
					},
				},
			},
			shouldMatch: true,
		},
		{
			name:        "service_description match_on in different order",
			ruleset:     "checkgroup_parameters:filesystem",
			description: "Filesystem patterns",
			conditions1: map[string]interface{}{
				"service_description": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"/tmp", "/var", "/opt"},
				},
			},
			conditions2: map[string]interface{}{
				"service_description": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"/var", "/opt", "/tmp"},
				},
			},
			shouldMatch: true,
		},
		{
			name:        "label_group labels in different order",
			ruleset:     "checkgroup_parameters:filesystem",
			description: "Label group rule",
			conditions1: map[string]interface{}{
				"host_label_groups": []interface{}{
					map[string]interface{}{
						"operator": "and",
						"label_group": []interface{}{
							map[string]interface{}{
								"key":      "region",
								"operator": "is",
								"value":    "us-east",
							},
							map[string]interface{}{
								"key":      "tier",
								"operator": "is",
								"value":    "backend",
							},
						},
					},
				},
			},
			conditions2: map[string]interface{}{
				"host_label_groups": []interface{}{
					map[string]interface{}{
						"operator": "and",
						"label_group": []interface{}{
							map[string]interface{}{
								"key":      "tier",
								"operator": "is",
								"value":    "backend",
							},
							map[string]interface{}{
								"key":      "region",
								"operator": "is",
								"value":    "us-east",
							},
						},
					},
				},
			},
			shouldMatch: true,
		},
		{
			name:        "complex conditions with multiple arrays in different order",
			ruleset:     "checkgroup_parameters:filesystem",
			description: "Complex rule",
			conditions1: map[string]interface{}{
				"host_name": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"server2", "server1"},
				},
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "location",
						"operator": "is",
						"value":    "dc1",
					},
					map[string]interface{}{
						"key":      "criticality",
						"operator": "is",
						"value":    "prod",
					},
				},
			},
			conditions2: map[string]interface{}{
				"host_name": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"server1", "server2"},
				},
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "criticality",
						"operator": "is",
						"value":    "prod",
					},
					map[string]interface{}{
						"key":      "location",
						"operator": "is",
						"value":    "dc1",
					},
				},
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := GenerateRuleHash(tt.ruleset, tt.description, tt.conditions1)
			hash2 := GenerateRuleHash(tt.ruleset, tt.description, tt.conditions2)

			if tt.shouldMatch {
				if hash1 != hash2 {
					t.Errorf("hashes should match but don't:\n  hash1 = %s\n  hash2 = %s", hash1, hash2)
				}
			} else {
				if hash1 == hash2 {
					t.Errorf("hashes should not match but do: %s", hash1)
				}
			}
		})
	}
}

func TestGenerateRuleHash_Uniqueness(t *testing.T) {
	tests := []struct {
		name         string
		ruleset1     string
		description1 string
		conditions1  interface{}
		ruleset2     string
		description2 string
		conditions2  interface{}
	}{
		{
			name:         "different rulesets",
			ruleset1:     "host_tags",
			description1: "Test rule",
			conditions1:  nil,
			ruleset2:     "inventory_df_rules",
			description2: "Test rule",
			conditions2:  nil,
		},
		{
			name:         "different descriptions",
			ruleset1:     "host_tags",
			description1: "First rule",
			conditions1:  nil,
			ruleset2:     "host_tags",
			description2: "Second rule",
			conditions2:  nil,
		},
		{
			name:         "different host_name operators",
			ruleset1:     "host_tags",
			description1: "Test rule",
			conditions1: map[string]interface{}{
				"host_name": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"host1"},
				},
			},
			ruleset2:     "host_tags",
			description2: "Test rule",
			conditions2: map[string]interface{}{
				"host_name": map[string]interface{}{
					"operator": "none_of",
					"match_on": []interface{}{"host1"},
				},
			},
		},
		{
			name:         "different host_name match_on values",
			ruleset1:     "host_tags",
			description1: "Test rule",
			conditions1: map[string]interface{}{
				"host_name": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"host1"},
				},
			},
			ruleset2:     "host_tags",
			description2: "Test rule",
			conditions2: map[string]interface{}{
				"host_name": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"host2"},
				},
			},
		},
		{
			name:         "different tag keys",
			ruleset1:     "inventory_df_rules",
			description1: "Test rule",
			conditions1: map[string]interface{}{
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "criticality",
						"operator": "is",
						"value":    "prod",
					},
				},
			},
			ruleset2:     "inventory_df_rules",
			description2: "Test rule",
			conditions2: map[string]interface{}{
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "location",
						"operator": "is",
						"value":    "prod",
					},
				},
			},
		},
		{
			name:         "different tag values",
			ruleset1:     "inventory_df_rules",
			description1: "Test rule",
			conditions1: map[string]interface{}{
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "criticality",
						"operator": "is",
						"value":    "prod",
					},
				},
			},
			ruleset2:     "inventory_df_rules",
			description2: "Test rule",
			conditions2: map[string]interface{}{
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "criticality",
						"operator": "is",
						"value":    "dev",
					},
				},
			},
		},
		{
			name:         "different number of conditions",
			ruleset1:     "host_tags",
			description1: "Test rule",
			conditions1: map[string]interface{}{
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "criticality",
						"operator": "is",
						"value":    "prod",
					},
				},
			},
			ruleset2:     "host_tags",
			description2: "Test rule",
			conditions2: map[string]interface{}{
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "criticality",
						"operator": "is",
						"value":    "prod",
					},
					map[string]interface{}{
						"key":      "location",
						"operator": "is",
						"value":    "dc1",
					},
				},
			},
		},
		{
			name:         "label vs tag condition",
			ruleset1:     "checkgroup_parameters:filesystem",
			description1: "Test rule",
			conditions1: map[string]interface{}{
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
			},
			ruleset2:     "checkgroup_parameters:filesystem",
			description2: "Test rule",
			conditions2: map[string]interface{}{
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
			hash1 := GenerateRuleHash(tt.ruleset1, tt.description1, tt.conditions1)
			hash2 := GenerateRuleHash(tt.ruleset2, tt.description2, tt.conditions2)

			if hash1 == hash2 {
				t.Errorf("hashes should be different but are the same: %s", hash1)
			}
		})
	}
}

func TestNormalizeConditions(t *testing.T) {
	tests := []struct {
		name       string
		conditions interface{}
		want       CanonicalConditions
	}{
		{
			name:       "nil conditions",
			conditions: nil,
			want:       CanonicalConditions{},
		},
		{
			name:       "empty map",
			conditions: map[string]interface{}{},
			want:       CanonicalConditions{},
		},
		{
			name: "host_name condition",
			conditions: map[string]interface{}{
				"host_name": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"host3", "host1", "host2"},
				},
			},
			want: CanonicalConditions{
				HostName: &HostNameCondition{
					Operator: "one_of",
					MatchOn:  []string{"host1", "host2", "host3"},
				},
			},
		},
		{
			name: "host_tags sorted by key",
			conditions: map[string]interface{}{
				"host_tags": []interface{}{
					map[string]interface{}{
						"key":      "zebra",
						"operator": "is",
						"value":    "value1",
					},
					map[string]interface{}{
						"key":      "alpha",
						"operator": "is",
						"value":    "value2",
					},
				},
			},
			want: CanonicalConditions{
				HostTags: []TagCondition{
					{Key: "alpha", Operator: "is", Value: "value2"},
					{Key: "zebra", Operator: "is", Value: "value1"},
				},
			},
		},
		{
			name: "host_labels sorted",
			conditions: map[string]interface{}{
				"host_labels": []interface{}{
					map[string]interface{}{
						"key":      "tier",
						"operator": "is",
						"value":    "backend",
					},
					map[string]interface{}{
						"key":      "environment",
						"operator": "is",
						"value":    "production",
					},
				},
			},
			want: CanonicalConditions{
				HostLabels: []LabelCondition{
					{Key: "environment", Operator: "is", Value: "production"},
					{Key: "tier", Operator: "is", Value: "backend"},
				},
			},
		},
		{
			name: "service_description with sorted match_on",
			conditions: map[string]interface{}{
				"service_description": map[string]interface{}{
					"operator": "one_of",
					"match_on": []interface{}{"/var", "/opt", "/tmp"},
				},
			},
			want: CanonicalConditions{
				ServiceDescription: &ServiceDescCondition{
					Operator: "one_of",
					MatchOn:  []string{"/opt", "/tmp", "/var"},
				},
			},
		},
		{
			name: "label groups with sorted labels",
			conditions: map[string]interface{}{
				"host_label_groups": []interface{}{
					map[string]interface{}{
						"operator": "and",
						"label_group": []interface{}{
							map[string]interface{}{
								"key":      "region",
								"operator": "is",
								"value":    "us-east",
							},
							map[string]interface{}{
								"key":      "tier",
								"operator": "is",
								"value":    "backend",
							},
						},
					},
				},
			},
			want: CanonicalConditions{
				HostLabelGroups: []LabelGroupCondition{
					{
						Operator: "and",
						LabelGroup: []LabelCondition{
							{Key: "region", Operator: "is", Value: "us-east"},
							{Key: "tier", Operator: "is", Value: "backend"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeConditions(tt.conditions)

			// Compare host_name
			if (got.HostName == nil) != (tt.want.HostName == nil) {
				t.Errorf("HostName presence mismatch: got %v, want %v", got.HostName, tt.want.HostName)
			}
			if got.HostName != nil && tt.want.HostName != nil {
				if got.HostName.Operator != tt.want.HostName.Operator {
					t.Errorf("HostName.Operator = %v, want %v", got.HostName.Operator, tt.want.HostName.Operator)
				}
				if !stringSlicesEqual(got.HostName.MatchOn, tt.want.HostName.MatchOn) {
					t.Errorf("HostName.MatchOn = %v, want %v", got.HostName.MatchOn, tt.want.HostName.MatchOn)
				}
			}

			// Compare host_tags
			if !tagConditionsEqual(got.HostTags, tt.want.HostTags) {
				t.Errorf("HostTags = %v, want %v", got.HostTags, tt.want.HostTags)
			}

			// Compare host_labels
			if !labelConditionsEqual(got.HostLabels, tt.want.HostLabels) {
				t.Errorf("HostLabels = %v, want %v", got.HostLabels, tt.want.HostLabels)
			}

			// Compare service_description
			if (got.ServiceDescription == nil) != (tt.want.ServiceDescription == nil) {
				t.Errorf("ServiceDescription presence mismatch: got %v, want %v", got.ServiceDescription, tt.want.ServiceDescription)
			}
			if got.ServiceDescription != nil && tt.want.ServiceDescription != nil {
				if got.ServiceDescription.Operator != tt.want.ServiceDescription.Operator {
					t.Errorf("ServiceDescription.Operator = %v, want %v", got.ServiceDescription.Operator, tt.want.ServiceDescription.Operator)
				}
				if !stringSlicesEqual(got.ServiceDescription.MatchOn, tt.want.ServiceDescription.MatchOn) {
					t.Errorf("ServiceDescription.MatchOn = %v, want %v", got.ServiceDescription.MatchOn, tt.want.ServiceDescription.MatchOn)
				}
			}

			// Compare host_label_groups
			if !labelGroupConditionsEqual(got.HostLabelGroups, tt.want.HostLabelGroups) {
				t.Errorf("HostLabelGroups = %v, want %v", got.HostLabelGroups, tt.want.HostLabelGroups)
			}
		})
	}
}

// Helper functions for test comparisons

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func tagConditionsEqual(a, b []TagCondition) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Key != b[i].Key || a[i].Operator != b[i].Operator || a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}

func labelConditionsEqual(a, b []LabelCondition) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Key != b[i].Key || a[i].Operator != b[i].Operator || a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}

func labelGroupConditionsEqual(a, b []LabelGroupCondition) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Operator != b[i].Operator {
			return false
		}
		if !labelConditionsEqual(a[i].LabelGroup, b[i].LabelGroup) {
			return false
		}
	}
	return true
}

// Benchmark tests

func BenchmarkGenerateRuleHash_Simple(b *testing.B) {
	conditions := map[string]interface{}{
		"host_name": map[string]interface{}{
			"operator": "one_of",
			"match_on": []interface{}{"host1", "host2"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRuleHash("host_tags", "Test rule", conditions)
	}
}

func BenchmarkGenerateRuleHash_Complex(b *testing.B) {
	conditions := map[string]interface{}{
		"host_name": map[string]interface{}{
			"operator": "one_of",
			"match_on": []interface{}{"host1", "host2", "host3"},
		},
		"host_tags": []interface{}{
			map[string]interface{}{
				"key":      "criticality",
				"operator": "is",
				"value":    "prod",
			},
			map[string]interface{}{
				"key":      "location",
				"operator": "is",
				"value":    "datacenter1",
			},
		},
		"host_labels": []interface{}{
			map[string]interface{}{
				"key":      "environment",
				"operator": "is",
				"value":    "production",
			},
		},
		"service_description": map[string]interface{}{
			"operator": "one_of",
			"match_on": []interface{}{"/var", "/tmp", "/opt"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRuleHash("checkgroup_parameters:filesystem", "Complex rule", conditions)
	}
}

func BenchmarkNormalizeConditions(b *testing.B) {
	conditions := map[string]interface{}{
		"host_tags": []interface{}{
			map[string]interface{}{
				"key":      "zebra",
				"operator": "is",
				"value":    "value1",
			},
			map[string]interface{}{
				"key":      "alpha",
				"operator": "is",
				"value":    "value2",
			},
			map[string]interface{}{
				"key":      "middle",
				"operator": "is",
				"value":    "value3",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NormalizeConditions(conditions)
	}
}
