// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package organization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	organizationapi "github.com/timrabl/terraform-provider-snipeit/internal/api/organization"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

// NewCompanyDataSource returns a new snipeit_company data source.
func NewCompanyDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[organizationapi.Company, CompanyResourceModel]{
		TypeSuffix: "_company",
		Path:       "/companies",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single company by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":    tfutil.DSID("company"),
				"name":  tfutil.DSLookupString("Name of the company. Set it to look up by exact name."),
				"phone": tfutil.DSString("Phone number of the company."),
				"fax":   tfutil.DSString("Fax number of the company."),
				"email": tfutil.DSString("Email address of the company."),
				"notes": tfutil.DSString("Free-form notes."),
			},
		},
		IDOf: func(m *CompanyResourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[CompanyResourceModel]{{
			Attr:  "name",
			Get:   func(m *CompanyResourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(_ context.Context, api *organizationapi.Company, m *CompanyResourceModel) error {
			m.fromAPI(api)
			return nil
		},
	})
}

// NewDepartmentDataSource returns a new snipeit_department data source.
func NewDepartmentDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[organizationapi.Department, DepartmentResourceModel]{
		TypeSuffix: "_department",
		Path:       "/departments",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single department by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":          tfutil.DSID("department"),
				"name":        tfutil.DSLookupString("Name of the department. Set it to look up by exact name."),
				"company_id":  tfutil.DSInt64("Id of the company this department belongs to."),
				"manager_id":  tfutil.DSInt64("Id of the user managing this department."),
				"location_id": tfutil.DSInt64("Id of the location of this department."),
				"notes":       tfutil.DSString("Free-form notes."),
			},
		},
		IDOf: func(m *DepartmentResourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[DepartmentResourceModel]{{
			Attr:  "name",
			Get:   func(m *DepartmentResourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(_ context.Context, api *organizationapi.Department, m *DepartmentResourceModel) error {
			m.fromAPI(api)
			return nil
		},
	})
}

// NewLocationDataSource returns a new snipeit_location data source.
func NewLocationDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[organizationapi.Location, LocationResourceModel]{
		TypeSuffix: "_location",
		Path:       "/locations",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single location by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":         tfutil.DSID("location"),
				"name":       tfutil.DSLookupString("Name of the location. Set it to look up by exact name."),
				"address":    tfutil.DSString("Street address of the location."),
				"address2":   tfutil.DSString("Street address of the location, line 2."),
				"city":       tfutil.DSString("City of the location."),
				"state":      tfutil.DSString("State/province of the location."),
				"country":    tfutil.DSString("Country of the location."),
				"zip":        tfutil.DSString("Postal code of the location."),
				"phone":      tfutil.DSString("Phone number of the location."),
				"fax":        tfutil.DSString("Fax number of the location."),
				"currency":   tfutil.DSString("Currency used at the location."),
				"parent_id":  tfutil.DSInt64("Id of the parent location."),
				"manager_id": tfutil.DSInt64("Id of the user managing this location."),
				"ldap_ou":    tfutil.DSString("LDAP organizational unit associated with the location."),
				"notes":      tfutil.DSString("Free-form notes."),
			},
		},
		IDOf: func(m *LocationResourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[LocationResourceModel]{{
			Attr:  "name",
			Get:   func(m *LocationResourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(_ context.Context, api *organizationapi.Location, m *LocationResourceModel) error {
			m.fromAPI(api)
			return nil
		},
	})
}

// NewSupplierDataSource returns a new snipeit_supplier data source.
func NewSupplierDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[organizationapi.Supplier, SupplierResourceModel]{
		TypeSuffix: "_supplier",
		Path:       "/suppliers",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single supplier by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":       tfutil.DSID("supplier"),
				"name":     tfutil.DSLookupString("Name of the supplier. Set it to look up by exact name."),
				"address":  tfutil.DSString("Street address of the supplier."),
				"address2": tfutil.DSString("Street address of the supplier, line 2."),
				"city":     tfutil.DSString("City of the supplier."),
				"state":    tfutil.DSString("State/province of the supplier."),
				"country":  tfutil.DSString("Country of the supplier."),
				"zip":      tfutil.DSString("Postal code of the supplier."),
				"phone":    tfutil.DSString("Phone number of the supplier."),
				"fax":      tfutil.DSString("Fax number of the supplier."),
				"email":    tfutil.DSString("Email address of the supplier."),
				"contact":  tfutil.DSString("Contact person at the supplier."),
				"url":      tfutil.DSString("Website of the supplier."),
				"notes":    tfutil.DSString("Free-form notes."),
			},
		},
		IDOf: func(m *SupplierResourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[SupplierResourceModel]{{
			Attr:  "name",
			Get:   func(m *SupplierResourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(_ context.Context, api *organizationapi.Supplier, m *SupplierResourceModel) error {
			m.fromAPI(api)
			return nil
		},
	})
}
