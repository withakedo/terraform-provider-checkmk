package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConvertToPython(t *testing.T) {
	tests := []struct {
		name     string
		input    attr.Value
		expected string
		wantErr  bool
	}{
		// Null and basic types
		{
			name:     "null value",
			input:    types.StringNull(),
			expected: "None",
		},
		{
			name:     "simple string",
			input:    types.StringValue("hello"),
			expected: "'hello'",
		},
		{
			name:     "string with single quote",
			input:    types.StringValue("it's working"),
			expected: `'it\'s working'`,
		},
		{
			name:     "string with backslash",
			input:    types.StringValue(`path\to\file`),
			expected: `'path\\to\\file'`,
		},
		{
			name:     "empty string",
			input:    types.StringValue(""),
			expected: "''",
		},

		// Booleans
		{
			name:     "bool true",
			input:    types.BoolValue(true),
			expected: "True",
		},
		{
			name:     "bool false",
			input:    types.BoolValue(false),
			expected: "False",
		},

		// Numbers
		{
			name:     "integer",
			input:    types.Int64Value(42),
			expected: "42",
		},
		{
			name:     "negative integer",
			input:    types.Int64Value(-123),
			expected: "-123",
		},
		{
			name:     "float",
			input:    types.Float64Value(3.14),
			expected: "3.14",
		},
		{
			name:     "float whole number",
			input:    types.Float64Value(5.0),
			expected: "5.0",
		},
		{
			name:     "zero",
			input:    types.Int64Value(0),
			expected: "0",
		},

		// Lists
		{
			name:     "empty list",
			input:    types.ListValueMust(types.StringType, []attr.Value{}),
			expected: "[]",
		},
		{
			name: "string list",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
				types.StringValue("c"),
			}),
			expected: "['a', 'b', 'c']",
		},
		{
			name: "number list",
			input: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(2),
				types.Int64Value(3),
			}),
			expected: "[1, 2, 3]",
		},

		// Maps
		{
			name:     "empty map",
			input:    types.MapValueMust(types.StringType, map[string]attr.Value{}),
			expected: "{}",
		},
		{
			name: "simple map",
			input: types.MapValueMust(types.StringType, map[string]attr.Value{
				"env":  types.StringValue("production"),
				"tier": types.StringValue("frontend"),
			}),
			expected: "{'env': 'production', 'tier': 'frontend'}",
		},
		{
			name: "map with numbers",
			input: types.MapValueMust(types.Int64Type, map[string]attr.Value{
				"count": types.Int64Value(5),
				"limit": types.Int64Value(100),
			}),
			expected: "{'count': 5, 'limit': 100}",
		},

		// Objects
		{
			name: "simple object",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"name":    types.StringType,
					"enabled": types.BoolType,
				},
				map[string]attr.Value{
					"name":    types.StringValue("test"),
					"enabled": types.BoolValue(true),
				},
			),
			expected: "{'enabled': True, 'name': 'test'}",
		},

		// Nested structures
		{
			name: "nested map",
			input: types.MapValueMust(
				types.MapType{ElemType: types.StringType},
				map[string]attr.Value{
					"labels": types.MapValueMust(types.StringType, map[string]attr.Value{
						"app": types.StringValue("myapp"),
					}),
				},
			),
			expected: "{'labels': {'app': 'myapp'}}",
		},
		{
			name: "map with list",
			input: types.MapValueMust(
				types.ListType{ElemType: types.StringType},
				map[string]attr.Value{
					"hosts": types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("host1"),
						types.StringValue("host2"),
					}),
				},
			),
			expected: "{'hosts': ['host1', 'host2']}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertToPython(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToPython() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("convertToPython() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestConvertToPython_Unknown(t *testing.T) {
	unknown := types.StringUnknown()
	_, err := convertToPython(unknown)
	if err == nil {
		t.Error("expected error for unknown value, got nil")
	}
}

func TestPythonString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "'hello'"},
		{"", "''"},
		{"it's", `'it\'s'`},
		{`back\slash`, `'back\\slash'`},
		{`both's and \`, `'both\'s and \\'`},
		{"with\nnewline", "'with\nnewline'"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := pythonString(tt.input)
			if result != tt.expected {
				t.Errorf("pythonString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatPythonFloat(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{3.14, "3.14"},
		{5.0, "5.0"},
		{0.0, "0.0"},
		{-1.5, "-1.5"},
		{1000000.0, "1e+06"},
		{0.000001, "1e-06"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatPythonFloat(tt.input)
			if result != tt.expected {
				t.Errorf("formatPythonFloat(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToPythonFunction_Metadata(t *testing.T) {
	f := NewToPythonFunction()
	var resp function.MetadataResponse
	f.Metadata(context.Background(), function.MetadataRequest{}, &resp)

	if resp.Name != "topython" {
		t.Errorf("Metadata() Name = %q, want %q", resp.Name, "topython")
	}
}

func TestToPythonFunction_Definition(t *testing.T) {
	f := NewToPythonFunction()
	var resp function.DefinitionResponse
	f.Definition(context.Background(), function.DefinitionRequest{}, &resp)

	if len(resp.Definition.Parameters) != 1 {
		t.Errorf("Definition() Parameters count = %d, want 1", len(resp.Definition.Parameters))
	}
}
