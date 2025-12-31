package groups

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// NewServiceGroupResource creates a new service group resource using the base implementation.
func NewServiceGroupResource() resource.Resource {
	return NewBaseGroupResource(ServiceGroupConfig)
}
