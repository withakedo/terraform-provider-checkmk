package labels

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// NewHostLabelsResource creates a new host labels resource using the base implementation.
func NewHostLabelsResource() resource.Resource {
	return NewBaseLabelResource(HostLabelsConfig)
}
