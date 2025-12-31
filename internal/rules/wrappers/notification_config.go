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

// Host Notification Resources

// HostFirstNotificationDelayResource is a typed wrapper for extra_host_conf:first_notification_delay.
type HostFirstNotificationDelayResource struct {
	BaseRuleWrapper
}

// NewHostFirstNotificationDelayResource creates a new host first notification delay resource.
func NewHostFirstNotificationDelayResource() resource.Resource {
	return &HostFirstNotificationDelayResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_host_conf:first_notification_delay",
				TypeName:      "host_first_notification_delay",
				ValueAttrName: "value",
				Description: "Sets the delay before the first notification for hosts. " +
					"This is a typed wrapper around the `extra_host_conf:first_notification_delay` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.Int64Attribute{
					MarkdownDescription: "Delay in seconds before the first notification is sent after a state change.",
					Required:            true,
				},
				ToValueRaw:   intToFloatString,
				FromValueRaw: floatStringToInt,
			},
		},
	}
}

// HostNotificationsEnabledResource is a typed wrapper for extra_host_conf:notifications_enabled.
type HostNotificationsEnabledResource struct {
	BaseRuleWrapper
}

// NewHostNotificationsEnabledResource creates a new host notifications enabled resource.
func NewHostNotificationsEnabledResource() resource.Resource {
	return &HostNotificationsEnabledResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_host_conf:notifications_enabled",
				TypeName:      "host_notifications_enabled",
				ValueAttrName: "value",
				Description: "Enables or disables notifications for hosts. " +
					"This is a typed wrapper around the `extra_host_conf:notifications_enabled` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.BoolAttribute{
					MarkdownDescription: "Whether notifications are enabled for matching hosts.",
					Required:            true,
				},
				ToValueRaw:   boolToPythonBool,
				FromValueRaw: pythonBoolToBool,
			},
		},
	}
}

// HostNotificationOptionsResource is a typed wrapper for extra_host_conf:notification_options.
type HostNotificationOptionsResource struct {
	BaseRuleWrapper
}

// NewHostNotificationOptionsResource creates a new host notification options resource.
func NewHostNotificationOptionsResource() resource.Resource {
	return &HostNotificationOptionsResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_host_conf:notification_options",
				TypeName:      "host_notification_options",
				ValueAttrName: "value",
				Description: "Sets which host state changes trigger notifications. " +
					"This is a typed wrapper around the `extra_host_conf:notification_options` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.SetAttribute{
					MarkdownDescription: "Set of notification options. Valid values: 'd' (down), 'u' (unreachable), " +
						"'r' (recovery), 'f' (flapping start/stop), 's' (scheduled downtime).",
					Required:    true,
					ElementType: types.StringType,
				},
				ToValueRaw:   setToPythonString,
				FromValueRaw: pythonStringToSet,
			},
		},
	}
}

// HostNotificationIntervalResource is a typed wrapper for extra_host_conf:notification_interval.
type HostNotificationIntervalResource struct {
	BaseRuleWrapper
}

// NewHostNotificationIntervalResource creates a new host notification interval resource.
func NewHostNotificationIntervalResource() resource.Resource {
	return &HostNotificationIntervalResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_host_conf:notification_interval",
				TypeName:      "host_notification_interval",
				ValueAttrName: "value",
				Description: "Sets the interval between repeated notifications for hosts. " +
					"This is a typed wrapper around the `extra_host_conf:notification_interval` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.Int64Attribute{
					MarkdownDescription: "Interval in seconds between repeated notifications. Set to 0 to disable repeated notifications.",
					Required:            true,
				},
				ToValueRaw:   intToFloatString,
				FromValueRaw: floatStringToInt,
			},
		},
	}
}

// Service Notification Resources

// ServiceFirstNotificationDelayResource is a typed wrapper for extra_service_conf:first_notification_delay.
type ServiceFirstNotificationDelayResource struct {
	BaseRuleWrapper
}

// NewServiceFirstNotificationDelayResource creates a new service first notification delay resource.
func NewServiceFirstNotificationDelayResource() resource.Resource {
	return &ServiceFirstNotificationDelayResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:first_notification_delay",
				TypeName:      "service_first_notification_delay",
				ValueAttrName: "value",
				Description: "Sets the delay before the first notification for services. " +
					"This is a typed wrapper around the `extra_service_conf:first_notification_delay` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.Int64Attribute{
					MarkdownDescription: "Delay in seconds before the first notification is sent after a state change.",
					Required:            true,
				},
				ToValueRaw:   intToFloatString,
				FromValueRaw: floatStringToInt,
			},
		},
	}
}

// ServiceNotificationsEnabledResource is a typed wrapper for extra_service_conf:notifications_enabled.
type ServiceNotificationsEnabledResource struct {
	BaseRuleWrapper
}

// NewServiceNotificationsEnabledResource creates a new service notifications enabled resource.
func NewServiceNotificationsEnabledResource() resource.Resource {
	return &ServiceNotificationsEnabledResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:notifications_enabled",
				TypeName:      "service_notifications_enabled",
				ValueAttrName: "value",
				Description: "Enables or disables notifications for services. " +
					"This is a typed wrapper around the `extra_service_conf:notifications_enabled` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.BoolAttribute{
					MarkdownDescription: "Whether notifications are enabled for matching services.",
					Required:            true,
				},
				ToValueRaw:   boolToPythonBool,
				FromValueRaw: pythonBoolToBool,
			},
		},
	}
}

// ServiceFlapDetectionEnabledResource is a typed wrapper for extra_service_conf:flap_detection_enabled.
type ServiceFlapDetectionEnabledResource struct {
	BaseRuleWrapper
}

// NewServiceFlapDetectionEnabledResource creates a new service flap detection enabled resource.
func NewServiceFlapDetectionEnabledResource() resource.Resource {
	return &ServiceFlapDetectionEnabledResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:flap_detection_enabled",
				TypeName:      "service_flap_detection_enabled",
				ValueAttrName: "value",
				Description: "Enables or disables flap detection for services. " +
					"This is a typed wrapper around the `extra_service_conf:flap_detection_enabled` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.BoolAttribute{
					MarkdownDescription: "Whether flap detection is enabled for matching services. " +
						"Flapping occurs when a service rapidly alternates between states.",
					Required: true,
				},
				ToValueRaw:   boolToPythonBool,
				FromValueRaw: pythonBoolToBool,
			},
		},
	}
}

// ServiceNotificationOptionsResource is a typed wrapper for extra_service_conf:notification_options.
type ServiceNotificationOptionsResource struct {
	BaseRuleWrapper
}

// NewServiceNotificationOptionsResource creates a new service notification options resource.
func NewServiceNotificationOptionsResource() resource.Resource {
	return &ServiceNotificationOptionsResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:notification_options",
				TypeName:      "service_notification_options",
				ValueAttrName: "value",
				Description: "Sets which service state changes trigger notifications. " +
					"This is a typed wrapper around the `extra_service_conf:notification_options` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.SetAttribute{
					MarkdownDescription: "Set of notification options. Valid values: 'w' (warning), 'c' (critical), " +
						"'u' (unknown), 'r' (recovery), 'f' (flapping start/stop), 's' (scheduled downtime).",
					Required:    true,
					ElementType: types.StringType,
				},
				ToValueRaw:   setToPythonString,
				FromValueRaw: pythonStringToSet,
			},
		},
	}
}

// ServiceNotificationIntervalResource is a typed wrapper for extra_service_conf:notification_interval.
type ServiceNotificationIntervalResource struct {
	BaseRuleWrapper
}

// NewServiceNotificationIntervalResource creates a new service notification interval resource.
func NewServiceNotificationIntervalResource() resource.Resource {
	return &ServiceNotificationIntervalResource{
		BaseRuleWrapper: BaseRuleWrapper{
			Config: RuleWrapperConfig{
				Ruleset:       "extra_service_conf:notification_interval",
				TypeName:      "service_notification_interval",
				ValueAttrName: "value",
				Description: "Sets the interval between repeated notifications for services. " +
					"This is a typed wrapper around the `extra_service_conf:notification_interval` ruleset. " +
					"Requires activation.",
				ValueSchema: schema.Int64Attribute{
					MarkdownDescription: "Interval in seconds between repeated notifications. Set to 0 to disable repeated notifications.",
					Required:            true,
				},
				ToValueRaw:   intToFloatString,
				FromValueRaw: floatStringToInt,
			},
		},
	}
}

// Conversion helpers for set/list types

// setToPythonString converts a set of strings to Python format.
func setToPythonString(ctx context.Context, value attr.Value) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	setVal, ok := value.(types.Set)
	if !ok {
		diags.AddError("Type Error", fmt.Sprintf("Expected Set, got %T", value))
		return "", diags
	}

	if setVal.IsNull() || setVal.IsUnknown() {
		diags.AddError("Value Error", "Value cannot be null or unknown")
		return "", diags
	}

	elements := setVal.Elements()
	if len(elements) == 0 {
		return "''", diags
	}

	// Build comma-separated string
	var values []string
	for _, elem := range elements {
		strElem, ok := elem.(types.String)
		if ok && !strElem.IsNull() {
			values = append(values, strElem.ValueString())
		}
	}

	// Join with commas (CheckMK expects 'd,u,r' format for notification_options)
	return "'" + strings.Join(values, ",") + "'", diags
}

// pythonStringToSet converts a Python comma-separated string to a set.
func pythonStringToSet(ctx context.Context, raw string) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if raw == "" || raw == "''" {
		return types.SetNull(types.StringType), diags
	}

	// Remove Python quotes
	raw = strings.TrimPrefix(raw, "'")
	raw = strings.TrimSuffix(raw, "'")
	raw = strings.TrimPrefix(raw, "\"")
	raw = strings.TrimSuffix(raw, "\"")

	if raw == "" {
		return types.SetNull(types.StringType), diags
	}

	// Split by comma
	parts := strings.Split(raw, ",")
	elements := make([]attr.Value, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			elements = append(elements, types.StringValue(part))
		}
	}

	if len(elements) == 0 {
		return types.SetNull(types.StringType), diags
	}

	setVal, setDiags := types.SetValue(types.StringType, elements)
	diags.Append(setDiags...)
	return setVal, diags
}
