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

// HostCheckIntervalResource is a typed wrapper for extra_host_conf:check_interval.
type HostCheckIntervalResource struct {
	BaseRuleWrapper
}

// NewHostCheckIntervalResource creates a new host check interval resource.
func NewHostCheckIntervalResource() resource.Resource {
	return &HostCheckIntervalResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_host_conf:check_interval",
				TypeName:      "host_check_interval",
				ValueAttrName: "value",
				Description: "Sets the check interval for hosts matching specified conditions. " +
					"This is a typed wrapper around the `extra_host_conf:check_interval` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.Int64Attribute{
					MarkdownDescription: "Check interval in seconds. How often CheckMK checks the host status.",
					Required:            true,
				},
				ToValueRaw:   intToFloatString,
				FromValueRaw: floatStringToInt,
			},
		},
	}
}

// ServiceCheckIntervalResource is a typed wrapper for extra_service_conf:check_interval.
type ServiceCheckIntervalResource struct {
	BaseRuleWrapper
}

// NewServiceCheckIntervalResource creates a new service check interval resource.
func NewServiceCheckIntervalResource() resource.Resource {
	return &ServiceCheckIntervalResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:check_interval",
				TypeName:      "service_check_interval",
				ValueAttrName: "value",
				Description: "Sets the check interval for services matching specified conditions. " +
					"This is a typed wrapper around the `extra_service_conf:check_interval` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.Int64Attribute{
					MarkdownDescription: "Check interval in seconds. How often CheckMK checks the service status.",
					Required:            true,
				},
				ToValueRaw:   intToFloatString,
				FromValueRaw: floatStringToInt,
			},
		},
	}
}

// intToFloatString converts an int64 value to a float string for the API.
// CheckMK expects check intervals as floats (e.g., "30.0").
func intToFloatString(ctx context.Context, value attr.Value) (string, diag.Diagnostics) {
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

	// Convert to float string (API expects float format)
	return fmt.Sprintf("%.1f", float64(intVal.ValueInt64())), diags
}

// floatStringToInt converts a float string from the API to an int64 value.
func floatStringToInt(ctx context.Context, raw string) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if raw == "" {
		return types.Int64Null(), diags
	}

	// Parse as float, then convert to int
	floatVal, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse value '%s' as float: %s", raw, err))
		return types.Int64Null(), diags
	}

	return types.Int64Value(int64(floatVal)), diags
}
