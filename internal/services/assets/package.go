// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

// Package assets implements the Terraform resources and data sources of the
// assets domain: manufacturers, categories, status labels, asset models and
// hardware (assets).
//
// Layering: schemas and TF state mapping live here; API types come from
// internal/api/assets (generated from api/assets.yaml); HTTP transport is
// internal/client.
package assets

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the resource constructors of this domain.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewManufacturerResource,
		NewCategoryResource,
		NewStatusLabelResource,
		NewModelResource,
		NewHardwareResource,
	}
}

// DataSources returns the data source constructors of this domain.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewManufacturerDataSource,
		NewCategoryDataSource,
		NewStatusLabelDataSource,
		NewModelDataSource,
		NewHardwareDataSource,
	}
}
