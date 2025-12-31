package labels

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// NewServiceLabelsResource creates a new service labels resource using the base implementation.
func NewServiceLabelsResource() resource.Resource {
	return NewBaseLabelResource(ServiceLabelsConfig)
}
