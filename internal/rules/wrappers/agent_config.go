package wrappers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AgentConfigMRPEResource is a typed wrapper for agent_config:mrpe.
type AgentConfigMRPEResource struct {
	BaseRuleWrapper
}

// NewAgentConfigMRPEResource creates a new agent config MRPE resource.
func NewAgentConfigMRPEResource() resource.Resource {
	return &AgentConfigMRPEResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "agent_config:mrpe",
				TypeName:      "agent_config_mrpe",
				ValueAttrName: "value",
				Description: "Configures MRPE (MK's Remote Plugin Executor) checks for the CheckMK agent. " +
					"This is a typed wrapper around the `agent_config:mrpe` ruleset. " +
					"Requires activation and agent baking.",
				ValueSchema: schema.ListNestedAttribute{
					MarkdownDescription: "List of MRPE check configurations. Each entry defines a custom check " +
						"that will be executed by the CheckMK agent.",
					Required: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"description": schema.StringAttribute{
								MarkdownDescription: "Service description/name for the check.",
								Required:            true,
							},
							"command_line": schema.StringAttribute{
								MarkdownDescription: "Full command line to execute for the check.",
								Required:            true,
							},
						},
					},
				},
				ToValueRaw:   mrpeListToPythonList,
				FromValueRaw: pythonListToMRPEList,
			},
		},
	}
}

// MRPEEntry represents a single MRPE check configuration.
type MRPEEntry struct {
	Description string `json:"description"`
	CommandLine string `json:"command_line"`
}

// mrpeListToPythonList converts a list of MRPE entries to Python list format.
func mrpeListToPythonList(ctx context.Context, value attr.Value) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	listVal, ok := value.(types.List)
	if !ok {
		diags.AddError("Type Error", fmt.Sprintf("Expected List, got %T", value))
		return "", diags
	}

	if listVal.IsNull() || listVal.IsUnknown() {
		diags.AddError("Value Error", "Value cannot be null or unknown")
		return "", diags
	}

	elements := listVal.Elements()
	if len(elements) == 0 {
		return "[]", diags
	}

	// Extract MRPE entries
	var entries []map[string]string
	for _, elem := range elements {
		objVal, ok := elem.(types.Object)
		if !ok {
			continue
		}

		attrs := objVal.Attributes()
		entry := make(map[string]string)

		if desc, ok := attrs["description"].(types.String); ok && !desc.IsNull() {
			entry["description"] = desc.ValueString()
		}
		if cmd, ok := attrs["command_line"].(types.String); ok && !cmd.IsNull() {
			entry["cmdline"] = cmd.ValueString() // API uses 'cmdline' not 'command_line'
		}

		if len(entry) > 0 {
			entries = append(entries, entry)
		}
	}

	// Convert to JSON (valid Python list/dict syntax)
	jsonBytes, err := json.Marshal(entries)
	if err != nil {
		diags.AddError("Conversion Error", fmt.Sprintf("Failed to convert MRPE list to JSON: %s", err))
		return "", diags
	}

	return string(jsonBytes), diags
}

// pythonListToMRPEList converts a Python list from the API to a list of MRPE entries.
func pythonListToMRPEList(ctx context.Context, raw string) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if raw == "" || raw == "[]" {
		attrTypes := map[string]attr.Type{
			"description":  types.StringType,
			"command_line": types.StringType,
		}
		return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
	}

	// Try to parse as JSON
	var entries []map[string]interface{}
	converted := strings.ReplaceAll(raw, "'", "\"")
	if err := json.Unmarshal([]byte(converted), &entries); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse MRPE list '%s': %s", raw, err))
		attrTypes := map[string]attr.Type{
			"description":  types.StringType,
			"command_line": types.StringType,
		}
		return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
	}

	attrTypes := map[string]attr.Type{
		"description":  types.StringType,
		"command_line": types.StringType,
	}

	var elements []attr.Value
	for _, entry := range entries {
		attrs := map[string]attr.Value{
			"description":  types.StringNull(),
			"command_line": types.StringNull(),
		}

		if desc, ok := entry["description"].(string); ok {
			attrs["description"] = types.StringValue(desc)
		}
		// API returns 'cmdline', we map it to 'command_line'
		if cmd, ok := entry["cmdline"].(string); ok {
			attrs["command_line"] = types.StringValue(cmd)
		} else if cmd, ok := entry["command_line"].(string); ok {
			attrs["command_line"] = types.StringValue(cmd)
		}

		objVal, objDiags := types.ObjectValue(attrTypes, attrs)
		diags.Append(objDiags...)
		if !objDiags.HasError() {
			elements = append(elements, objVal)
		}
	}

	listVal, listDiags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elements)
	diags.Append(listDiags...)
	return listVal, diags
}
