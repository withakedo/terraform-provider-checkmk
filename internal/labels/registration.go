package labels

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns all resources for the labels package
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewHostLabelsResource,
		NewServiceLabelsResource,
	}
}

// DataSources returns all data sources for the labels package
func DataSources() []func() datasource.DataSource {
	return nil
}
