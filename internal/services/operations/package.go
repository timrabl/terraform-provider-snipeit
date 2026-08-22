// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

// Package operations implements the Terraform resources and data sources of
// the operations domain: asset maintenances, the activity report, and the
// hardware audit due/overdue lists.
//
// Layering: schemas and TF state mapping live here; API types come from
// internal/api/operations (generated from api/operations.yaml); HTTP
// transport is internal/client.
package operations

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the resource constructors of this domain.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewMaintenanceResource,
	}
}

// DataSources returns the data source constructors of this domain.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewActivityReportsDataSource,
		NewHardwareAuditDueDataSource,
		NewHardwareAuditOverdueDataSource,
	}
}
