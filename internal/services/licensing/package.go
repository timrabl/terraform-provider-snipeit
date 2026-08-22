// Package licensing implements the Terraform resources and data sources of
// the licensing domain: licenses and license seat assignments.
//
// Layering: schemas and TF state mapping live here; API types come from
// internal/api/licensing (generated from api/licensing.yaml); HTTP transport
// is internal/client.
package licensing

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the resource constructors of this domain.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewLicenseResource,
		NewLicenseSeatResource,
	}
}

// DataSources returns the data source constructors of this domain.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewLicenseDataSource,
	}
}
