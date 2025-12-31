package configuration

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns all resources for the configuration service
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewFolderResource,
		NewTagGroupResource,
		NewAuxTagResource,
		NewTimePeriodResource,
	}
}

// DataSources returns all data sources for the configuration service
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAuxTagDataSource,
		NewFolderDataSource,
		NewTagGroupDataSource,
		NewTimePeriodDataSource,
	}
}
