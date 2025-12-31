package wrappers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HostCheckCommandsResource is a typed wrapper for host_check_commands.
type HostCheckCommandsResource struct {
	BaseRuleWrapper
}

// NewHostCheckCommandsResource creates a new host check commands resource.
func NewHostCheckCommandsResource() resource.Resource {
	return &HostCheckCommandsResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "host_check_commands",
				TypeName:      "host_check_commands",
				ValueAttrName: "value",
				Description: "Sets the host check command. " +
					"This is a typed wrapper around the `host_check_commands` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.StringAttribute{
					MarkdownDescription: "The host check command to use. " +
						"Common values include 'ping', 'smart-ping', 'tcp', 'ok', 'agent', 'service'.",
					Required: true,
				},
				ToValueRaw:   stringToPythonString,
				FromValueRaw: pythonStringToString,
			},
		},
	}
}

// HostMaxCheckAttemptsResource is a typed wrapper for extra_host_conf:max_check_attempts.
type HostMaxCheckAttemptsResource struct {
	BaseRuleWrapper
}

// NewHostMaxCheckAttemptsResource creates a new host max check attempts resource.
func NewHostMaxCheckAttemptsResource() resource.Resource {
	return &HostMaxCheckAttemptsResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_host_conf:max_check_attempts",
				TypeName:      "host_max_check_attempts",
				ValueAttrName: "value",
				Description: "Sets the maximum check attempts for hosts before changing state to hard. " +
					"This is a typed wrapper around the `extra_host_conf:max_check_attempts` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.Int64Attribute{
					MarkdownDescription: "Maximum number of check attempts before the host state becomes hard. " +
						"If a host fails this many times in a row, it transitions from soft to hard state.",
					Required: true,
				},
				ToValueRaw:   intToIntString,
				FromValueRaw: intStringToInt,
			},
		},
	}
}

// HostRetryIntervalResource is a typed wrapper for extra_host_conf:retry_interval.
type HostRetryIntervalResource struct {
	BaseRuleWrapper
}

// NewHostRetryIntervalResource creates a new host retry interval resource.
func NewHostRetryIntervalResource() resource.Resource {
	return &HostRetryIntervalResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_host_conf:retry_interval",
				TypeName:      "host_retry_interval",
				ValueAttrName: "value",
				Description: "Sets the retry interval for hosts in soft state. " +
					"This is a typed wrapper around the `extra_host_conf:retry_interval` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.Int64Attribute{
					MarkdownDescription: "Retry interval in seconds. How often to recheck a host in soft state.",
					Required:            true,
				},
				ToValueRaw:   intToFloatString,
				FromValueRaw: floatStringToInt,
			},
		},
	}
}

// HostCheckPeriodResource is a typed wrapper for extra_host_conf:check_period.
type HostCheckPeriodResource struct {
	BaseRuleWrapper
}

// NewHostCheckPeriodResource creates a new host check period resource.
func NewHostCheckPeriodResource() resource.Resource {
	return &HostCheckPeriodResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_host_conf:check_period",
				TypeName:      "host_check_period",
				ValueAttrName: "value",
				Description: "Sets the check period for hosts. " +
					"This is a typed wrapper around the `extra_host_conf:check_period` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.StringAttribute{
					MarkdownDescription: "Name of the time period during which checks are performed. " +
						"Must reference an existing time period (e.g., '24X7', 'workhours').",
					Required: true,
				},
				ToValueRaw:   stringToPythonString,
				FromValueRaw: pythonStringToString,
			},
		},
	}
}

// ServiceCheckPeriodResource is a typed wrapper for extra_service_conf:check_period.
type ServiceCheckPeriodResource struct {
	BaseRuleWrapper
}

// NewServiceCheckPeriodResource creates a new service check period resource.
func NewServiceCheckPeriodResource() resource.Resource {
	return &ServiceCheckPeriodResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:check_period",
				TypeName:      "service_check_period",
				ValueAttrName: "value",
				Description: "Sets the check period for services. " +
					"This is a typed wrapper around the `extra_service_conf:check_period` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.StringAttribute{
					MarkdownDescription: "Name of the time period during which checks are performed. " +
						"Must reference an existing time period (e.g., '24X7', 'workhours').",
					Required: true,
				},
				ToValueRaw:   stringToPythonString,
				FromValueRaw: pythonStringToString,
			},
		},
	}
}

// boolToPythonBool converts a bool value to API format ("'1'" or "'0'" - quoted strings).
func boolToPythonBool(ctx context.Context, value attr.Value) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	boolVal, ok := value.(types.Bool)
	if !ok {
		diags.AddError("Type Error", fmt.Sprintf("Expected Bool, got %T", value))
		return "", diags
	}

	if boolVal.IsNull() || boolVal.IsUnknown() {
		diags.AddError("Value Error", "Value cannot be null or unknown")
		return "", diags
	}

	// CheckMK API expects "'1'" or "'0'" for enabled/disabled rules (quoted strings)
	if boolVal.ValueBool() {
		return "'1'", diags
	}
	return "'0'", diags
}

// pythonBoolToBool converts a Python boolean from the API to a bool value.
func pythonBoolToBool(ctx context.Context, raw string) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if raw == "" {
		return types.BoolNull(), diags
	}

	// Handle Python boolean values (may be quoted or unquoted)
	switch raw {
	case "True", "true", "1", "'1'", "\"1\"":
		return types.BoolValue(true), diags
	case "False", "false", "0", "'0'", "\"0\"":
		return types.BoolValue(false), diags
	default:
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse value '%s' as boolean", raw))
		return types.BoolNull(), diags
	}
}
