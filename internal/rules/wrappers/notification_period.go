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
)

// HostNotificationPeriodResource is a typed wrapper for extra_host_conf:notification_period.
type HostNotificationPeriodResource struct {
	BaseRuleWrapper
}

// NewHostNotificationPeriodResource creates a new host notification period resource.
func NewHostNotificationPeriodResource() resource.Resource {
	return &HostNotificationPeriodResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_host_conf:notification_period",
				TypeName:      "host_notification_period",
				ValueAttrName: "value",
				Description: "Sets the notification period for hosts matching specified conditions. " +
					"This is a typed wrapper around the `extra_host_conf:notification_period` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.StringAttribute{
					MarkdownDescription: "Name of the time period during which notifications are sent. " +
						"Must reference an existing time period (e.g., '24X7', 'workhours').",
					Required: true,
				},
				ToValueRaw:   stringToPythonString,
				FromValueRaw: pythonStringToString,
			},
		},
	}
}

// ServiceNotificationPeriodResource is a typed wrapper for extra_service_conf:notification_period.
type ServiceNotificationPeriodResource struct {
	BaseRuleWrapper
}

// NewServiceNotificationPeriodResource creates a new service notification period resource.
func NewServiceNotificationPeriodResource() resource.Resource {
	return &ServiceNotificationPeriodResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:notification_period",
				TypeName:      "service_notification_period",
				ValueAttrName: "value",
				Description: "Sets the notification period for services matching specified conditions. " +
					"This is a typed wrapper around the `extra_service_conf:notification_period` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.StringAttribute{
					MarkdownDescription: "Name of the time period during which notifications are sent. " +
						"Must reference an existing time period (e.g., '24X7', 'workhours').",
					Required: true,
				},
				ToValueRaw:   stringToPythonString,
				FromValueRaw: pythonStringToString,
			},
		},
	}
}

// stringToPythonString converts a string value to Python string format.
// CheckMK expects string values quoted in Python style (e.g., "'24X7'").
func stringToPythonString(ctx context.Context, value attr.Value) (string, diag.Diagnostics) {
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

	// Wrap in Python single quotes
	return fmt.Sprintf("'%s'", strVal.ValueString()), diags
}

// pythonStringToString converts a Python string from the API to a plain string.
func pythonStringToString(ctx context.Context, raw string) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if raw == "" {
		return types.StringNull(), diags
	}

	// Remove Python quotes if present
	result := raw
	if strings.HasPrefix(result, "'") && strings.HasSuffix(result, "'") {
		result = result[1 : len(result)-1]
	} else if strings.HasPrefix(result, "\"") && strings.HasSuffix(result, "\"") {
		result = result[1 : len(result)-1]
	}

	return types.StringValue(result), diags
}
