// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package organization

// Pure unit tests for the API-object -> TF-state mapping and the request body
// builders. These run without TF_ACC and without any network access; the JSON
// fixtures mirror real v8.0.4 responses.

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	organizationapi "github.com/timrabl/terraform-provider-snipeit/internal/api/organization"
)

const companyFixture = `{
	"id": 3,
	"name": "ACME GmbH",
	"phone": "+49 8031 0000",
	"fax": null,
	"email": "",
	"notes": "managed by terraform",
	"created_at": {"datetime": "2026-08-22 13:31:07", "formatted": "2026-08-22 01:31 PM"},
	"assets_count": 12
}`

func TestCompanyFromAPI(t *testing.T) {
	var api organizationapi.Company
	if err := json.Unmarshal([]byte(companyFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m CompanyResourceModel
	m.fromAPI(&api)

	if m.ID.ValueInt64() != 3 {
		t.Errorf("id = %v", m.ID)
	}
	if m.Name.ValueString() != "ACME GmbH" {
		t.Errorf("name = %v", m.Name)
	}
	if m.Phone.ValueString() != "+49 8031 0000" {
		t.Errorf("phone = %v", m.Phone)
	}
	// null and "" must both map to TF null so unset attributes stay unset.
	if !m.Fax.IsNull() {
		t.Errorf("fax should be null, got %v", m.Fax)
	}
	if !m.Email.IsNull() {
		t.Errorf("email should be null, got %v", m.Email)
	}
	if m.Notes.ValueString() != "managed by terraform" {
		t.Errorf("notes = %v", m.Notes)
	}
}

const departmentFixture = `{
	"id": 5,
	"name": "IT",
	"company": {"id": 3, "name": "ACME GmbH"},
	"manager": null,
	"location": {"id": 0, "name": ""},
	"notes": null
}`

func TestDepartmentFromAPI(t *testing.T) {
	var api organizationapi.Department
	if err := json.Unmarshal([]byte(departmentFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m DepartmentResourceModel
	m.fromAPI(&api)

	if m.CompanyID.ValueInt64() != 3 {
		t.Errorf("company_id = %v", m.CompanyID)
	}
	if !m.ManagerID.IsNull() {
		t.Errorf("manager_id should be null (absent ref), got %v", m.ManagerID)
	}
	// A ref with id 0 counts as unset too.
	if !m.LocationID.IsNull() {
		t.Errorf("location_id should be null (zero ref), got %v", m.LocationID)
	}
}

const locationFixture = `{
	"id": 8,
	"name": "Rosenheim HQ",
	"address": "Musterstr. 1",
	"address2": null,
	"city": "Rosenheim",
	"state": "",
	"country": "DE",
	"zip": "83022",
	"phone": null,
	"fax": null,
	"currency": "EUR",
	"parent": {"id": 2, "name": "Bavaria"},
	"manager": null,
	"ldap_ou": null,
	"notes": ""
}`

func TestLocationFromAPI(t *testing.T) {
	var api organizationapi.Location
	if err := json.Unmarshal([]byte(locationFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m LocationResourceModel
	m.fromAPI(&api)

	if m.City.ValueString() != "Rosenheim" || m.Country.ValueString() != "DE" {
		t.Errorf("city/country = %v/%v", m.City, m.Country)
	}
	if !m.State.IsNull() {
		t.Errorf("state should be null, got %v", m.State)
	}
	if m.ParentID.ValueInt64() != 2 {
		t.Errorf("parent_id = %v", m.ParentID)
	}
	if !m.ManagerID.IsNull() {
		t.Errorf("manager_id should be null, got %v", m.ManagerID)
	}
	if m.Currency.ValueString() != "EUR" {
		t.Errorf("currency = %v", m.Currency)
	}
}

func TestDepartmentToBodyClearSemantics(t *testing.T) {
	m := DepartmentResourceModel{
		Name:       types.StringValue("IT"),
		CompanyID:  types.Int64Value(3),
		ManagerID:  types.Int64Null(),    // cleared ref -> explicit JSON null
		LocationID: types.Int64Unknown(), // unknown -> omitted
		Notes:      types.StringNull(),   // cleared string -> ""
	}
	body := m.toBody()

	if body["name"] != "IT" {
		t.Errorf("name = %v", body["name"])
	}
	if body["company_id"] != int64(3) {
		t.Errorf("company_id = %v", body["company_id"])
	}
	if v, ok := body["manager_id"]; !ok || v != nil {
		t.Errorf("manager_id must be present and nil, got %v (present=%v)", v, ok)
	}
	if _, ok := body["location_id"]; ok {
		t.Errorf("unknown location_id must be omitted")
	}
	if body["notes"] != "" {
		t.Errorf("null notes must be sent as empty string, got %v", body["notes"])
	}
}

func TestSupplierToBodySendsAllStrings(t *testing.T) {
	m := SupplierResourceModel{
		Name: types.StringValue("ACME"),
		City: types.StringValue("Rosenheim"),
		// everything else unset
	}
	body := m.toBody()

	// Strings are always present ("" clears server-side).
	for _, key := range []string{"address", "address2", "city", "state", "country",
		"zip", "phone", "fax", "email", "contact", "url", "notes"} {
		if _, ok := body[key]; !ok {
			t.Errorf("string key %q must always be present", key)
		}
	}
	if body["city"] != "Rosenheim" {
		t.Errorf("city = %v", body["city"])
	}
	if body["state"] != "" {
		t.Errorf("unset state must be empty string, got %v", body["state"])
	}
}
