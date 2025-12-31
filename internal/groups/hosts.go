package groups

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// NewHostGroupResource creates a new host group resource using the base implementation.
func NewHostGroupResource() resource.Resource {
	return NewBaseGroupResource(HostGroupConfig)
}
