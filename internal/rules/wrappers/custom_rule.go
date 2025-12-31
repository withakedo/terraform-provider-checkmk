package wrappers

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ServiceCustomRuleResource is a typed wrapper for custom_checks.
type ServiceCustomRuleResource struct {
	BaseRuleWrapper
}

// NewServiceCustomRuleResource creates a new service custom rule resource.
func NewServiceCustomRuleResource() resource.Resource {
	return &ServiceCustomRuleResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "custom_checks",
				TypeName:      "service_custom_rule",
				ValueAttrName: "value",
				Description: "Defines a custom check command for services matching specified conditions. " +
					"This is a typed wrapper around the `custom_checks` ruleset. " +
					"Requires activation.\n\n" +
					"The value should be the full command line to execute as the custom check.",
				ValueSchema: schema.StringAttribute{
					MarkdownDescription: "The command to execute as the custom check. " +
						"This should be the full path to the check script or command.",
					Required: true,
				},
				ToValueRaw:   commandToPythonTuple,
				FromValueRaw: pythonTupleToCommand,
			},
		},
	}
}

// commandToPythonTuple converts a command string to Python tuple format.
// CheckMK custom_checks expects a tuple/dict structure.
// Simple format: ('service_name', 'command')
func commandToPythonTuple(ctx context.Context, value attr.Value) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	strVal, ok := value.(types.String)
	if !ok {
		diags.AddError("Type Error", fmt.Sprintf("Expected String, got %T", value))
		return "", diags
	}

	if strVal.IsNull() || strVal.IsUnknown() {
		diags.AddError("Value Error", "Value cannot be null or unknown")
		return "", diags
	}

	command := strVal.ValueString()

	// Generate a service name from the command (basename without extension)
	serviceName := extractServiceName(command)

	// Format as Python tuple: ('Service Name', 'command')
	// The tuple format is: (service_description, command_line)
	return fmt.Sprintf("('%s', '%s')", serviceName, command), diags
}

// pythonTupleToCommand converts a Python tuple from the API to a command string.
func pythonTupleToCommand(ctx context.Context, raw string) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if raw == "" {
		return types.StringNull(), diags
	}

	// Parse Python tuple format: ('service_name', 'command')
	// or dict format: {'service_description': 'name', 'command_line': 'cmd'}
	command := extractCommandFromPython(raw)
	if command == "" {
		// If we can't parse it, return the raw value
		command = raw
	}

	return types.StringValue(command), diags
}

// extractServiceName generates a service name from a command path.
func extractServiceName(command string) string {
	// Get the basename
	parts := strings.Split(command, "/")
	name := parts[len(parts)-1]

	// Remove common prefixes and extensions
	name = strings.TrimPrefix(name, "check_")
	name = strings.TrimSuffix(name, ".sh")
	name = strings.TrimSuffix(name, ".py")
	name = strings.TrimSuffix(name, ".pl")

	// Convert underscores to spaces and capitalize
	name = strings.ReplaceAll(name, "_", " ")
	name = cases.Title(language.English).String(name)

	if name == "" {
		name = "Custom Check"
	}

	return name
}

// extractCommandFromPython extracts the command from various Python formats.
func extractCommandFromPython(raw string) string {
	raw = strings.TrimSpace(raw)

	// Handle tuple format: ('name', 'command') or ("name", "command")
	if strings.HasPrefix(raw, "(") && strings.HasSuffix(raw, ")") {
		inner := raw[1 : len(raw)-1]
		// Split by comma and get the second element (command)
		parts := splitPythonTuple(inner)
		if len(parts) >= 2 {
			return unquotePython(parts[1])
		}
	}

	// Handle dict format: {'command_line': 'cmd', ...}
	if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
		// Look for 'command_line' key
		if idx := strings.Index(raw, "'command_line'"); idx != -1 {
			rest := raw[idx+len("'command_line'"):]
			// Find the value after the colon
			if colonIdx := strings.Index(rest, ":"); colonIdx != -1 {
				valueStart := rest[colonIdx+1:]
				valueStart = strings.TrimSpace(valueStart)
				// Extract the quoted value
				if strings.HasPrefix(valueStart, "'") {
					endQuote := strings.Index(valueStart[1:], "'")
					if endQuote != -1 {
						return valueStart[1 : endQuote+1]
					}
				}
			}
		}
	}

	// If it's just a quoted string, unquote it
	return unquotePython(raw)
}

// splitPythonTuple splits a Python tuple's contents by comma, respecting quotes.
func splitPythonTuple(s string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]
		if !inQuote && (c == '\'' || c == '"') {
			inQuote = true
			quoteChar = c
			current.WriteByte(c)
		} else if inQuote && c == quoteChar {
			inQuote = false
			current.WriteByte(c)
		} else if !inQuote && c == ',' {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}

	return parts
}

// unquotePython removes Python quotes from a string.
func unquotePython(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) ||
			(strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) {
			return s[1 : len(s)-1]
		}
	}
	return s
}
