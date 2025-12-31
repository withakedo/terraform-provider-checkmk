package wrappers

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// AllResources returns all typed rule wrapper resources for registration with the provider.
func AllResources() []func() resource.Resource {
	return []func() resource.Resource{
		// Check intervals (check_interval.go)
		NewHostCheckIntervalResource,
		NewServiceCheckIntervalResource,

		// Notification periods (notification_period.go)
		NewHostNotificationPeriodResource,
		NewServiceNotificationPeriodResource,

		// Retry config (retry_config.go)
		NewServiceMaxCheckAttemptsResource,
		NewServiceRetryIntervalResource,

		// Custom rules (custom_rule.go)
		NewServiceCustomRuleResource,

		// Host configuration (host_config.go)
		NewHostCheckCommandsResource,
		NewHostMaxCheckAttemptsResource,
		NewHostRetryIntervalResource,
		NewHostCheckPeriodResource,
		NewServiceCheckPeriodResource,

		// Notification configuration (notification_config.go)
		// Host notifications
		NewHostFirstNotificationDelayResource,
		NewHostNotificationsEnabledResource,
		NewHostNotificationOptionsResource,
		NewHostNotificationIntervalResource,
		// Service notifications
		NewServiceFirstNotificationDelayResource,
		NewServiceNotificationsEnabledResource,
		NewServiceFlapDetectionEnabledResource,
		NewServiceNotificationOptionsResource,
		NewServiceNotificationIntervalResource,

		// Service configuration (service_config.go)
		NewServiceActiveChecksEnabledResource,
		NewServicePassiveChecksEnabledResource,
		NewServiceProcessPerfDataResource,
		NewCustomServiceAttributesResource,
		NewServiceLabelRulesResource,

		// Agent configuration (agent_config.go)
		NewAgentConfigMRPEResource,
	}
}
