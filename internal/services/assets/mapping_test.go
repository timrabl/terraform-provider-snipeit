// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package assets

// Pure unit tests for the API-object -> TF-state mapping and the request body
// builders. These run without TF_ACC and without any network access; the JSON
// fixtures mirror real v8.0.4 responses including their decorated values.

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	assetsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/assets"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

const manufacturerFixture = `{
	"id": 1,
	"name": "Apple",
	"url": "https://apple.com",
	"support_url": "",
	"warranty_lookup_url": "",
	"support_phone": "",
	"support_email": "support@apple.com",
	"notes": null,
	"assets_count": 3,
	"created_at": {"datetime": "2026-08-22 13:31:07", "formatted": "2026-08-22 01:31 PM"}
}`

func TestManufacturerFromAPI(t *testing.T) {
	var api assetsapi.Manufacturer
	if err := json.Unmarshal([]byte(manufacturerFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m ManufacturerResourceModel
	m.fromAPI(&api)

	if m.ID.ValueInt64() != 1 || m.Name.ValueString() != "Apple" {
		t.Errorf("id/name = %v/%v", m.ID, m.Name)
	}
	if m.URL.ValueString() != "https://apple.com" {
		t.Errorf("url = %v", m.URL)
	}
	// "" and null must both map to TF null.
	if !m.SupportURL.IsNull() || !m.SupportPhone.IsNull() || !m.Notes.IsNull() {
		t.Errorf("empty fields should be null: support_url=%v support_phone=%v notes=%v",
			m.SupportURL, m.SupportPhone, m.Notes)
	}
	if m.SupportEmail.ValueString() != "support@apple.com" {
		t.Errorf("support_email = %v", m.SupportEmail)
	}
}

const categoryFixture = `{
	"id": 4,
	"name": "Laptops",
	"category_type": "Asset",
	"eula": false,
	"use_default_eula": "0",
	"require_acceptance": 1,
	"checkin_email": true,
	"notes": ""
}`

func TestCategoryFromAPI(t *testing.T) {
	var api assetsapi.Category
	if err := json.Unmarshal([]byte(categoryFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m CategoryResourceModel
	m.fromAPI(&api)

	// The API capitalizes the type; state must be lowercase.
	if m.CategoryType.ValueString() != "asset" {
		t.Errorf("category_type = %v", m.CategoryType)
	}
	// FlexBool handles "0", 1 and true.
	if m.UseDefaultEULA.ValueBool() != false {
		t.Errorf("use_default_eula = %v", m.UseDefaultEULA)
	}
	if m.RequireAcceptance.ValueBool() != true {
		t.Errorf("require_acceptance = %v", m.RequireAcceptance)
	}
	if m.CheckinEmail.ValueBool() != true {
		t.Errorf("checkin_email = %v", m.CheckinEmail)
	}
	if !m.Notes.IsNull() {
		t.Errorf("empty notes should be null, got %v", m.Notes)
	}
	// eula_text is write-only: fromAPI must not touch it.
	pre := CategoryResourceModel{EULAText: types.StringValue("keep me")}
	pre.fromAPI(&api)
	if pre.EULAText.ValueString() != "keep me" {
		t.Errorf("eula_text must survive fromAPI, got %v", pre.EULAText)
	}
}

const modelFixture = `{
	"id": 7,
	"name": "MacBook Pro 14",
	"model_number": "A2779",
	"category": {"id": 4, "name": "Laptops"},
	"manufacturer": {"id": 1, "name": "Apple"},
	"fieldset": null,
	"eol": "36 mo.",
	"min_amt": null,
	"requestable": "1",
	"notes": null
}`

func TestModelFromAPI(t *testing.T) {
	var api assetsapi.Model
	if err := json.Unmarshal([]byte(modelFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m ModelResourceModel
	m.fromAPI(&api)

	if m.CategoryID.ValueInt64() != 4 || m.ManufacturerID.ValueInt64() != 1 {
		t.Errorf("category/manufacturer = %v/%v", m.CategoryID, m.ManufacturerID)
	}
	if !m.FieldsetID.IsNull() {
		t.Errorf("fieldset_id should be null, got %v", m.FieldsetID)
	}
	// Decorated "36 mo." must decode to 36.
	if m.EOL.ValueInt64() != 36 {
		t.Errorf("eol = %v", m.EOL)
	}
	// null min_amt -> 0 -> TF null.
	if !m.MinAmt.IsNull() {
		t.Errorf("min_amt should be null, got %v", m.MinAmt)
	}
	if m.Requestable.ValueBool() != true {
		t.Errorf("requestable = %v", m.Requestable)
	}
}

const hardwareFixture = `{
	"id": 42,
	"asset_tag": "IT-0042",
	"purchase_cost": "1,234.50",
	"name": "",
	"serial": "C02XXXXXXX",
	"order_number": null,
	"notes": null,
	"model": {"id": 7, "name": "MacBook Pro 14"},
	"status_label": {"id": 2, "name": "Ready to Deploy", "status_type": "deployable"},
	"company": null,
	"supplier": {"id": 0, "name": ""},
	"rtd_location": {"id": 8, "name": "Rosenheim HQ"},
	"purchase_date": {"date": "2026-01-15", "formatted": "Wed Jan 15, 2026"},
	"warranty_months": "24 months",
	"requestable": 0
}`

func TestHardwareFromAPI(t *testing.T) {
	var api assetsapi.Hardware
	if err := json.Unmarshal([]byte(hardwareFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m HardwareResourceModel
	m.fromAPI(&api)

	if m.AssetTag.ValueString() != "IT-0042" {
		t.Errorf("asset_tag = %v", m.AssetTag)
	}
	if m.ModelID.ValueInt64() != 7 || m.StatusID.ValueInt64() != 2 {
		t.Errorf("model/status = %v/%v", m.ModelID, m.StatusID)
	}
	if !m.Name.IsNull() {
		t.Errorf("empty name should be null, got %v", m.Name)
	}
	if m.Serial.ValueString() != "C02XXXXXXX" {
		t.Errorf("serial = %v", m.Serial)
	}
	if m.PurchaseCost.ValueString() != "1234.50" {
		t.Errorf("purchase_cost should be normalized, got %v", m.PurchaseCost)
	}
	// absent ref and zero ref both -> null.
	if !m.CompanyID.IsNull() || !m.SupplierID.IsNull() {
		t.Errorf("company/supplier should be null: %v/%v", m.CompanyID, m.SupplierID)
	}
	// rtd_location nested ref -> location_id.
	if m.LocationID.ValueInt64() != 8 {
		t.Errorf("location_id = %v", m.LocationID)
	}
	if m.PurchaseDate.ValueString() != "2026-01-15" {
		t.Errorf("purchase_date = %v", m.PurchaseDate)
	}
	// Decorated "24 months" -> 24.
	if m.WarrantyMonths.ValueInt64() != 24 {
		t.Errorf("warranty_months = %v", m.WarrantyMonths)
	}
	if m.Requestable.ValueBool() != false {
		t.Errorf("requestable = %v", m.Requestable)
	}
}

func TestHardwareToBodyClearSemantics(t *testing.T) {
	m := HardwareResourceModel{
		AssetTag:       types.StringValue("IT-0042"),
		ModelID:        types.Int64Value(7),
		StatusID:       types.Int64Value(2),
		Name:           types.StringNull(),   // cleared string -> ""
		PurchaseDate:   types.StringNull(),   // cleared date -> explicit null
		WarrantyMonths: types.Int64Null(),    // cleared int -> explicit null
		LocationID:     types.Int64Value(8),  // rtd_location_id write-key
		SupplierID:     types.Int64Unknown(), // unknown -> omitted
		Requestable:    types.BoolUnknown(),  // unknown -> omitted
	}
	body := m.toBody()

	if body["asset_tag"] != "IT-0042" || body["model_id"] != int64(7) || body["status_id"] != int64(2) {
		t.Errorf("required keys wrong: %v", body)
	}
	if body["name"] != "" {
		t.Errorf("cleared name must be empty string, got %v", body["name"])
	}
	if v, ok := body["purchase_date"]; !ok || v != nil {
		t.Errorf("cleared purchase_date must be explicit null, got %v (present=%v)", v, ok)
	}
	if v, ok := body["warranty_months"]; !ok || v != nil {
		t.Errorf("cleared warranty_months must be explicit null, got %v (present=%v)", v, ok)
	}
	// The default location is written under rtd_location_id, never location_id.
	if body["rtd_location_id"] != int64(8) {
		t.Errorf("rtd_location_id = %v", body["rtd_location_id"])
	}
	if _, ok := body["location_id"]; ok {
		t.Errorf("location_id must not appear in request bodies")
	}
	if _, ok := body["supplier_id"]; ok {
		t.Errorf("unknown supplier_id must be omitted")
	}
	if _, ok := body["requestable"]; ok {
		t.Errorf("unknown requestable must be omitted")
	}
}

func TestModelToBodyClearSemantics(t *testing.T) {
	m := ModelResourceModel{
		Name:           types.StringValue("X"),
		CategoryID:     types.Int64Value(4),
		ManufacturerID: types.Int64Null(), // cleared ref -> explicit null
		EOL:            types.Int64Null(), // cleared -> explicit null
	}
	body := m.toBody()

	if v, ok := body["manufacturer_id"]; !ok || v != nil {
		t.Errorf("manufacturer_id must be present and nil, got %v (present=%v)", v, ok)
	}
	if v, ok := body["eol"]; !ok || v != nil {
		t.Errorf("eol must be present and nil, got %v (present=%v)", v, ok)
	}
}

func TestHardwareToBodyMoneySemantics(t *testing.T) {
	var m HardwareResourceModel
	m.AssetTag = types.StringValue("X-1")
	m.ModelID = types.Int64Value(1)
	m.StatusID = types.Int64Value(2)
	m.PurchaseCost = tfutil.NewMoneyValue("1,234.50")
	body := m.toBody()
	if body["purchase_cost"] != "1234.50" {
		t.Errorf("purchase_cost must be sent normalized, got %v", body["purchase_cost"])
	}
	m.PurchaseCost = tfutil.NewMoneyNull()
	body = m.toBody()
	if v, ok := body["purchase_cost"]; !ok || v != nil {
		t.Errorf("null purchase_cost must be sent as explicit null, got %v (present=%v)", v, ok)
	}
}
