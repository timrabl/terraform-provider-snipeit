// Package inventory implements the Terraform resources and data sources of
// the inventory domain: accessories, consumables, components, and the
// checkout/checkin action resources (including asset checkouts).
//
// Layering: schemas and TF state mapping live here; API types come from
// internal/api/inventory (generated from api/inventory.yaml); HTTP transport
// is internal/client.
package inventory

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the resource constructors of this domain.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewAccessoryResource,
		NewConsumableResource,
		NewComponentResource,
		NewHardwareCheckoutResource,
		NewAccessoryCheckoutResource,
		NewComponentCheckoutResource,
		NewConsumableCheckoutResource,
	}
}

// DataSources returns the data source constructors of this domain.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccessoryDataSource,
		NewConsumableDataSource,
		NewComponentDataSource,
	}
}
