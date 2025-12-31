package users

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewContactGroupResource,
		NewPasswordResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
		NewContactGroupDataSource,
		NewPasswordDataSource,
	}
}
