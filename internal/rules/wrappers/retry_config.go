package wrappers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ServiceMaxCheckAttemptsResource is a typed wrapper for extra_service_conf:max_check_attempts.
type ServiceMaxCheckAttemptsResource struct {
	BaseRuleWrapper
}

// NewServiceMaxCheckAttemptsResource creates a new service max check attempts resource.
func NewServiceMaxCheckAttemptsResource() resource.Resource {
	return &ServiceMaxCheckAttemptsResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:max_check_attempts",
				TypeName:      "service_max_check_attempts",
				ValueAttrName: "value",
				Description: "Sets the maximum check attempts for services before changing state to hard. " +
					"This is a typed wrapper around the `extra_service_conf:max_check_attempts` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.Int64Attribute{
					MarkdownDescription: "Maximum number of check attempts before the service state becomes hard. " +
						"If a service fails this many times in a row, it transitions from soft to hard state.",
					Required: true,
				},
				ToValueRaw:   intToIntString,
				FromValueRaw: intStringToInt,
			},
		},
	}
}

// ServiceRetryIntervalResource is a typed wrapper for extra_service_conf:retry_interval.
type ServiceRetryIntervalResource struct {
	BaseRuleWrapper
}

// NewServiceRetryIntervalResource creates a new service retry interval resource.
func NewServiceRetryIntervalResource() resource.Resource {
	return &ServiceRetryIntervalResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:retry_interval",
				TypeName:      "service_retry_interval",
				ValueAttrName: "value",
				Description: "Sets the retry interval for services in soft state. " +
					"This is a typed wrapper around the `extra_service_conf:retry_interval` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.Int64Attribute{
					MarkdownDescription: "Retry interval in seconds. How often to recheck a service in soft state.",
					Required:            true,
				},
				ToValueRaw:   intToFloatString,
				FromValueRaw: floatStringToInt,
			},
		},
	}
}

// intToIntString converts an int64 value to an integer string for the API.
// Some CheckMK rulesets expect integer format (e.g., "3" not "3.0").
func intToIntString(ctx context.Context, value attr.Value) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	intVal, ok := value.(types.Int64)
	if !ok {
		diags.AddError("Type Error", fmt.Sprintf("Expected Int64, got %T", value))
		return "", diags
	}

	if intVal.IsNull() || intVal.IsUnknown() {
		diags.AddError("Value Error", "Value cannot be null or unknown")
		return "", diags
	}

	return strconv.FormatInt(intVal.ValueInt64(), 10), diags
}

// intStringToInt converts an integer string from the API to an int64 value.
func intStringToInt(ctx context.Context, raw string) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if raw == "" {
		return types.Int64Null(), diags
	}

	// Try parsing as int first
	intVal, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		// Try parsing as float (API might return "3.0" format)
		floatVal, floatErr := strconv.ParseFloat(raw, 64)
		if floatErr != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to parse value '%s' as integer: %s", raw, err))
			return types.Int64Null(), diags
		}
		intVal = int64(floatVal)
	}

	return types.Int64Value(intVal), diags
}
