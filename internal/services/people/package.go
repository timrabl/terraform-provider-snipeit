// Package people implements the Terraform resources and data sources of the
// people domain: users and permission groups.
//
// Layering: schemas and TF state mapping live here; API types come from
// internal/api/people (generated from api/people.yaml); HTTP transport is
// internal/client.
package people

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the resource constructors of this domain.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewGroupResource,
	}
}

// DataSources returns the data source constructors of this domain.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
		NewGroupDataSource,
		NewUserMeDataSource,
	}
}
