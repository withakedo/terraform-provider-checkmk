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

// ServiceActiveChecksEnabledResource is a typed wrapper for extra_service_conf:active_checks_enabled.
type ServiceActiveChecksEnabledResource struct {
	BaseRuleWrapper
}

// NewServiceActiveChecksEnabledResource creates a new service active checks enabled resource.
func NewServiceActiveChecksEnabledResource() resource.Resource {
	return &ServiceActiveChecksEnabledResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:active_checks_enabled",
				TypeName:      "service_active_checks_enabled",
				ValueAttrName: "value",
				Description: "Enables or disables active checks for services. " +
					"This is a typed wrapper around the `extra_service_conf:active_checks_enabled` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.BoolAttribute{
					MarkdownDescription: "Whether active checks are enabled for matching services. " +
						"When disabled, the service will not be actively checked by CheckMK.",
					Required: true,
				},
				ToValueRaw:   boolToPythonBool,
				FromValueRaw: pythonBoolToBool,
			},
		},
	}
}

// ServicePassiveChecksEnabledResource is a typed wrapper for extra_service_conf:passive_checks_enabled.
type ServicePassiveChecksEnabledResource struct {
	BaseRuleWrapper
}

// NewServicePassiveChecksEnabledResource creates a new service passive checks enabled resource.
func NewServicePassiveChecksEnabledResource() resource.Resource {
	return &ServicePassiveChecksEnabledResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:passive_checks_enabled",
				TypeName:      "service_passive_checks_enabled",
				ValueAttrName: "value",
				Description: "Enables or disables passive checks for services. " +
					"This is a typed wrapper around the `extra_service_conf:passive_checks_enabled` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.BoolAttribute{
					MarkdownDescription: "Whether passive checks are enabled for matching services. " +
						"When disabled, the service will not accept passive check results.",
					Required: true,
				},
				ToValueRaw:   boolToPythonBool,
				FromValueRaw: pythonBoolToBool,
			},
		},
	}
}

// ServiceProcessPerfDataResource is a typed wrapper for extra_service_conf:process_perf_data.
type ServiceProcessPerfDataResource struct {
	BaseRuleWrapper
}

// NewServiceProcessPerfDataResource creates a new service process perf data resource.
func NewServiceProcessPerfDataResource() resource.Resource {
	return &ServiceProcessPerfDataResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:process_perf_data",
				TypeName:      "service_process_perf_data",
				ValueAttrName: "value",
				Description: "Enables or disables performance data processing for services. " +
					"This is a typed wrapper around the `extra_service_conf:process_perf_data` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.BoolAttribute{
					MarkdownDescription: "Whether performance data processing is enabled for matching services. " +
						"When disabled, performance data from checks will not be stored.",
					Required: true,
				},
				ToValueRaw:   boolToPythonBool,
				FromValueRaw: pythonBoolToBool,
			},
		},
	}
}

// CustomServiceAttributesResource is a typed wrapper for custom_service_attributes.
type CustomServiceAttributesResource struct {
	BaseRuleWrapper
}

// NewCustomServiceAttributesResource creates a new custom service attributes resource.
func NewCustomServiceAttributesResource() resource.Resource {
	return &CustomServiceAttributesResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "custom_service_attributes",
				TypeName:      "custom_service_attributes",
				ValueAttrName: "value",
				Description: "Sets custom attributes for services. " +
					"This is a typed wrapper around the `custom_service_attributes` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.MapAttribute{
					MarkdownDescription: "Map of custom attribute names to values. " +
						"These attributes can be used in notification scripts and for filtering.",
					Required:    true,
					ElementType: types.StringType,
				},
				ToValueRaw:   mapToPythonDict,
				FromValueRaw: pythonDictToMap,
			},
		},
	}
}

// Conversion helpers for map types

// mapToPythonDict converts a map of strings to Python dict format.
func mapToPythonDict(ctx context.Context, value attr.Value) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	mapVal, ok := value.(types.Map)
	if !ok {
		diags.AddError("Type Error", fmt.Sprintf("Expected Map, got %T", value))
		return "", diags
	}

	if mapVal.IsNull() || mapVal.IsUnknown() {
		diags.AddError("Value Error", "Value cannot be null or unknown")
		return "", diags
	}

	elements := mapVal.Elements()
	if len(elements) == 0 {
		return "{}", diags
	}

	// Build Python dict format
	result := make(map[string]string)
	for k, v := range elements {
		strVal, ok := v.(types.String)
		if ok && !strVal.IsNull() {
			result[k] = strVal.ValueString()
		}
	}

	// Convert to JSON format which is valid Python dict syntax
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		diags.AddError("Conversion Error", fmt.Sprintf("Failed to convert map to JSON: %s", err))
		return "", diags
	}

	// Replace JSON quotes with Python-style single quotes for keys
	return string(jsonBytes), diags
}

// pythonDictToMap converts a Python dict from the API to a map.
func pythonDictToMap(ctx context.Context, raw string) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if raw == "" || raw == "{}" {
		return types.MapNull(types.StringType), diags
	}

	// Try to parse as JSON (Python dicts with double quotes are valid JSON)
	var result map[string]string
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		// Try converting Python single quotes to double quotes
		converted := strings.ReplaceAll(raw, "'", "\"")
		if err := json.Unmarshal([]byte(converted), &result); err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to parse dict value '%s': %s", raw, err))
			return types.MapNull(types.StringType), diags
		}
	}

	if len(result) == 0 {
		return types.MapNull(types.StringType), diags
	}

	elements := make(map[string]attr.Value)
	for k, v := range result {
		elements[k] = types.StringValue(v)
	}

	mapVal, mapDiags := types.MapValue(types.StringType, elements)
	diags.Append(mapDiags...)
	return mapVal, diags
}
