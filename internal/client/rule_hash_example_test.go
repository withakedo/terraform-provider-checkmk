package client_test

import (
	"fmt"

	"github.com/terraform-provider-checkmk/internal/client"
)

// Example demonstrates basic rule hashing
func ExampleGenerateRuleHash() {
	// Simple rule with no conditions
	hash1 := client.GenerateRuleHash("host_tags", "Test rule", nil)
	fmt.Printf("Hash length: %d\n", len(hash1))

	// Rule with host_name condition
	conditions := map[string]interface{}{
		"host_name": map[string]interface{}{
			"operator": "one_of",
			"match_on": []interface{}{"host1", "host2"},
		},
	}
	hash2 := client.GenerateRuleHash("host_tags", "Test rule", conditions)
	fmt.Printf("With conditions: %s\n", hash2)

	// Output:
	// Hash length: 32
	// With conditions: 95ebb0921b2cb6f0e2fe7bde19c65150
}

// Example_orderIndependence demonstrates that hash is order-independent
func Example_orderIndependence() {
	// Same conditions in different order should produce the same hash
	conditions1 := map[string]interface{}{
		"host_tags": []interface{}{
			map[string]interface{}{"key": "criticality", "operator": "is", "value": "prod"},
			map[string]interface{}{"key": "location", "operator": "is", "value": "dc1"},
		},
	}

	conditions2 := map[string]interface{}{
		"host_tags": []interface{}{
			map[string]interface{}{"key": "location", "operator": "is", "value": "dc1"},
			map[string]interface{}{"key": "criticality", "operator": "is", "value": "prod"},
		},
	}

	hash1 := client.GenerateRuleHash("inventory_df_rules", "Test", conditions1)
	hash2 := client.GenerateRuleHash("inventory_df_rules", "Test", conditions2)

	fmt.Printf("Hashes match: %t\n", hash1 == hash2)

	// Output:
	// Hashes match: true
}

// Example_complexRule demonstrates hashing a complex rule
func Example_complexRule() {
	conditions := map[string]interface{}{
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
				"key":      "environment",
				"operator": "is",
				"value":    "production",
			},
		},
		"service_description": map[string]interface{}{
			"operator": "one_of",
			"match_on": []interface{}{"/var", "/tmp"},
		},
	}

	hash := client.GenerateRuleHash("checkgroup_parameters:filesystem", "Complex rule", conditions)
	fmt.Printf("Complex rule hash: %s\n", hash)

	// Output:
	// Complex rule hash: db3d56965e7e011074c63ec06773e221
}

// Example_normalizeConditions demonstrates condition normalization
func Example_normalizeConditions() {
	conditions := map[string]interface{}{
		"host_tags": []interface{}{
			map[string]interface{}{"key": "zebra", "operator": "is", "value": "last"},
			map[string]interface{}{"key": "alpha", "operator": "is", "value": "first"},
		},
	}

	normalized := client.NormalizeConditions(conditions)

	// Tags are sorted alphabetically by key
	fmt.Printf("First tag key: %s\n", normalized.HostTags[0].Key)
	fmt.Printf("Second tag key: %s\n", normalized.HostTags[1].Key)

	// Output:
	// First tag key: alpha
	// Second tag key: zebra
}
