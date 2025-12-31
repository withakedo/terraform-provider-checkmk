package provider

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var _ function.Function = &ToPythonFunction{}

// ToPythonFunction converts HCL values to Python literal format.
// This is needed because CheckMK's rule API expects value_raw in Python format,
// not JSON. Python uses single quotes for strings and True/False for booleans.
type ToPythonFunction struct{}

// NewToPythonFunction creates a new instance of the topython function.
func NewToPythonFunction() function.Function {
	return &ToPythonFunction{}
}

// Metadata returns the function name.
func (f *ToPythonFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "topython"
}

// Definition returns the function definition.
func (f *ToPythonFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Converts an HCL value to Python literal format",
		Description: "Converts any HCL value (string, number, bool, list, map, object) to a Python literal string. " +
			"This is required for CheckMK rule values which expect Python format, not JSON. " +
			"Key differences: Python uses single quotes for strings, True/False for booleans, and None for null.",
		MarkdownDescription: "Converts any HCL value to Python literal format for use with `checkmk_rule.value_raw`.\n\n" +
			"CheckMK's rule API expects values in Python literal format, not JSON:\n" +
			"- Strings use single quotes: `'value'`\n" +
			"- Booleans are `True`/`False`\n" +
			"- Null is `None`\n" +
			"- Maps use single quotes: `{'key': 'value'}`\n\n" +
			"**Example:**\n" +
			"```hcl\n" +
			"resource \"checkmk_rule\" \"labels\" {\n" +
			"  ruleset   = \"host_label_rules\"\n" +
			"  value_raw = provider::checkmk::topython({\n" +
			"    env  = \"production\"\n" +
			"    tier = \"frontend\"\n" +
			"  })\n" +
			"}\n" +
			"```",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:                "value",
				Description:         "The HCL value to convert to Python literal format.",
				MarkdownDescription: "The HCL value to convert. Can be any type: string, number, bool, list, map, or object.",
			},
		},
		Return: function.StringReturn{},
	}
}

// Run executes the function logic.
func (f *ToPythonFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var arg types.Dynamic

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &arg))
	if resp.Error != nil {
		return
	}

	result, err := convertToPython(arg.UnderlyingValue())
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error,
			function.NewFuncError(fmt.Sprintf("Error converting value to Python: %s", err)))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, result))
}

// convertToPython recursively converts an attr.Value to Python literal format.
func convertToPython(val attr.Value) (string, error) {
	if val == nil || val.IsNull() {
		return "None", nil
	}

	if val.IsUnknown() {
		return "", fmt.Errorf("cannot convert unknown value to Python")
	}

	switch v := val.(type) {
	case types.String:
		return pythonString(v.ValueString()), nil

	case types.Bool:
		if v.ValueBool() {
			return "True", nil
		}
		return "False", nil

	case types.Int64:
		return fmt.Sprintf("%d", v.ValueInt64()), nil

	case types.Int32:
		return fmt.Sprintf("%d", v.ValueInt32()), nil

	case types.Float64:
		return formatPythonFloat(v.ValueFloat64()), nil

	case types.Float32:
		return formatPythonFloat(float64(v.ValueFloat32())), nil

	case types.Number:
		// Number uses *big.Float
		bf := v.ValueBigFloat()
		if bf == nil {
			return "None", nil
		}
		return formatBigFloat(bf), nil

	case types.List:
		return convertListToPython(v)

	case types.Set:
		return convertSetToPython(v)

	case types.Tuple:
		return convertTupleToPython(v)

	case types.Map:
		return convertMapToPython(v)

	case types.Object:
		return convertObjectToPython(v)

	default:
		return "", fmt.Errorf("unsupported type: %T", val)
	}
}

// pythonString formats a string as a Python string literal with single quotes.
// Handles escaping of single quotes and backslashes.
func pythonString(s string) string {
	// Escape backslashes first, then single quotes
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return fmt.Sprintf("'%s'", escaped)
}

// formatPythonFloat formats a float64 as a Python float literal.
func formatPythonFloat(f float64) string {
	// Use %g for compact representation, but ensure it looks like a float
	s := fmt.Sprintf("%g", f)
	// If no decimal point or exponent, add .0 to make it clearly a float
	if !strings.Contains(s, ".") && !strings.Contains(s, "e") && !strings.Contains(s, "E") {
		s += ".0"
	}
	return s
}

// formatBigFloat formats a *big.Float as a Python number literal.
func formatBigFloat(bf *big.Float) string {
	// Check if it's an integer
	if bf.IsInt() {
		// Convert to integer string
		i, _ := bf.Int64()
		return fmt.Sprintf("%d", i)
	}
	// Use float representation
	f, _ := bf.Float64()
	return formatPythonFloat(f)
}

// convertListToPython converts a types.List to Python list literal.
func convertListToPython(list types.List) (string, error) {
	elements := list.Elements()
	parts := make([]string, 0, len(elements))

	for _, elem := range elements {
		s, err := convertToPython(elem)
		if err != nil {
			return "", err
		}
		parts = append(parts, s)
	}

	return fmt.Sprintf("[%s]", strings.Join(parts, ", ")), nil
}

// convertSetToPython converts a types.Set to Python set literal.
// Note: Empty sets in Python are set(), not {}.
func convertSetToPython(set types.Set) (string, error) {
	elements := set.Elements()
	if len(elements) == 0 {
		return "set()", nil
	}

	parts := make([]string, 0, len(elements))
	for _, elem := range elements {
		s, err := convertToPython(elem)
		if err != nil {
			return "", err
		}
		parts = append(parts, s)
	}

	// Sort for deterministic output
	sort.Strings(parts)

	return fmt.Sprintf("{%s}", strings.Join(parts, ", ")), nil
}

// convertTupleToPython converts a types.Tuple to Python tuple literal.
func convertTupleToPython(tuple types.Tuple) (string, error) {
	elements := tuple.Elements()
	parts := make([]string, 0, len(elements))

	for _, elem := range elements {
		s, err := convertToPython(elem)
		if err != nil {
			return "", err
		}
		parts = append(parts, s)
	}

	// Single-element tuples need a trailing comma in Python
	if len(parts) == 1 {
		return fmt.Sprintf("(%s,)", parts[0]), nil
	}

	return fmt.Sprintf("(%s)", strings.Join(parts, ", ")), nil
}

// convertMapToPython converts a types.Map to Python dict literal.
func convertMapToPython(m types.Map) (string, error) {
	elements := m.Elements()
	if len(elements) == 0 {
		return "{}", nil
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(elements))
	for k := range elements {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(elements))
	for _, k := range keys {
		v := elements[k]
		valStr, err := convertToPython(v)
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s: %s", pythonString(k), valStr))
	}

	return fmt.Sprintf("{%s}", strings.Join(parts, ", ")), nil
}

// convertObjectToPython converts a types.Object to Python dict literal.
func convertObjectToPython(obj types.Object) (string, error) {
	attrs := obj.Attributes()
	if len(attrs) == 0 {
		return "{}", nil
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(attrs))
	for _, k := range keys {
		v := attrs[k]
		valStr, err := convertToPython(v)
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s: %s", pythonString(k), valStr))
	}

	return fmt.Sprintf("{%s}", strings.Join(parts, ", ")), nil
}
