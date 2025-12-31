package groups

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns all resources for the groups package
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewHostGroupResource,
		NewServiceGroupResource,
	}
}

// DataSources returns all data sources for the groups package
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewHostGroupDataSource,
		NewServiceGroupDataSource,
	}
}
