// Package customfields implements the Terraform resources and data sources of
// the customfields domain: custom fields, fieldsets, and the association
// between them.
//
// Layering: schemas and TF state mapping live here; API types come from
// internal/api/customfields (generated from api/customfields.yaml); HTTP
// transport is internal/client.
package customfields

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the resource constructors of this domain.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewFieldsetResource,
		NewFieldResource,
		NewFieldFieldsetAssociationResource,
	}
}

// DataSources returns the data source constructors of this domain.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewFieldsetDataSource,
		NewFieldDataSource,
	}
}
