// Package rules provides rule-related resources for the CheckMK Terraform provider.
// This includes the generic rule resource and notification rules.
// Note: Rule wrappers are registered separately by the provider to avoid import cycles.
package rules

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns rule resources for registration with the provider.
// Note: Rule wrappers from internal/rules/wrappers are registered
// separately by the provider to allow wrappers to import condition utilities
// from this package without creating an import cycle.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewRuleResource,
		NewNotificationRuleResource,
	}
}

// DataSources returns all rule-related data sources for registration with the provider.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRuleDataSource,
		NewNotificationRuleDataSource,
	}
}
