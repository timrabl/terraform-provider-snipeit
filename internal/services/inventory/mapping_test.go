// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package inventory

// Pure unit tests for the API-object -> TF-state mapping and the request body
// builders. These run without TF_ACC and without any network access; the JSON
// fixtures mirror real v8.0.4 responses.

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	inventoryapi "github.com/timrabl/terraform-provider-snipeit/internal/api/inventory"
)

const accessoryFixture = `{
	"id": 7,
	"name": "USB-C Dock",
	"qty": "5",
	"category": {"id": 4, "name": "Docks"},
	"manufacturer": {"id": 2, "name": "Dell"},
	"supplier": null,
	"company": {"id": 0, "name": ""},
	"location": {"id": 9, "name": "HQ"},
	"model_number": "WD19",
	"order_number": "",
	"purchase_date": {"date": "2026-01-10", "formatted": "Sat Jan 10, 2026"},
	"min_qty": 2,
	"notes": null,
	"remaining_qty": 5
}`

func TestAccessoryFromAPI(t *testing.T) {
	var api inventoryapi.Accessory
	if err := json.Unmarshal([]byte(accessoryFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m AccessoryResourceModel
	m.fromAPI(&api)

	if m.ID.ValueInt64() != 7 {
		t.Errorf("id = %v", m.ID)
	}
	// qty arrives as the string "5" — FlexInt must decode it.
	if m.Qty.ValueInt64() != 5 {
		t.Errorf("qty = %v", m.Qty)
	}
	if m.CategoryID.ValueInt64() != 4 {
		t.Errorf("category_id = %v", m.CategoryID)
	}
	if m.ManufacturerID.ValueInt64() != 2 {
		t.Errorf("manufacturer_id = %v", m.ManufacturerID)
	}
	// Absent ref and zero ref both map to null.
	if !m.SupplierID.IsNull() {
		t.Errorf("supplier_id should be null, got %v", m.SupplierID)
	}
	if !m.CompanyID.IsNull() {
		t.Errorf("company_id should be null (zero ref), got %v", m.CompanyID)
	}
	if m.LocationID.ValueInt64() != 9 {
		t.Errorf("location_id = %v", m.LocationID)
	}
	// "" and null strings map to TF null.
	if !m.OrderNumber.IsNull() {
		t.Errorf("order_number should be null, got %v", m.OrderNumber)
	}
	if !m.Notes.IsNull() {
		t.Errorf("notes should be null, got %v", m.Notes)
	}
	// The read serializer's min_qty feeds the min_amt attribute.
	if m.MinAmt.ValueInt64() != 2 {
		t.Errorf("min_amt = %v", m.MinAmt)
	}
	if m.PurchaseDate.ValueString() != "2026-01-10" {
		t.Errorf("purchase_date = %v", m.PurchaseDate)
	}
}

const consumableFixture = `{
	"id": 11,
	"name": "Toner",
	"qty": 10,
	"category": {"id": 6, "name": "Printer Supplies"},
	"company": null,
	"manufacturer": null,
	"location": null,
	"item_no": "ITEM-9",
	"model_number": null,
	"order_number": "ORD-2",
	"purchase_date": null,
	"min_amt": 3,
	"notes": ""
}`

func TestConsumableFromAPI(t *testing.T) {
	var api inventoryapi.Consumable
	if err := json.Unmarshal([]byte(consumableFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m ConsumableResourceModel
	m.fromAPI(&api)

	if m.Qty.ValueInt64() != 10 {
		t.Errorf("qty = %v", m.Qty)
	}
	if m.ItemNo.ValueString() != "ITEM-9" {
		t.Errorf("item_no = %v", m.ItemNo)
	}
	if !m.PurchaseDate.IsNull() {
		t.Errorf("purchase_date should be null, got %v", m.PurchaseDate)
	}
	if m.MinAmt.ValueInt64() != 3 {
		t.Errorf("min_amt = %v", m.MinAmt)
	}
	if !m.Notes.IsNull() {
		t.Errorf("empty notes should be null, got %v", m.Notes)
	}
}

const componentFixture = `{
	"id": 13,
	"name": "16GB DIMM",
	"qty": 8,
	"category": {"id": 5, "name": "RAM"},
	"supplier": null,
	"company": null,
	"location": null,
	"serial": "SER-77",
	"order_number": null,
	"purchase_date": {"date": "", "formatted": ""},
	"min_amt": 0,
	"notes": null
}`

func TestComponentFromAPI(t *testing.T) {
	var api inventoryapi.Component
	if err := json.Unmarshal([]byte(componentFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m ComponentResourceModel
	m.fromAPI(&api)

	if m.Serial.ValueString() != "SER-77" {
		t.Errorf("serial = %v", m.Serial)
	}
	// A date object with empty date maps to null.
	if !m.PurchaseDate.IsNull() {
		t.Errorf("purchase_date should be null, got %v", m.PurchaseDate)
	}
	// min_amt 0 counts as unset.
	if !m.MinAmt.IsNull() {
		t.Errorf("min_amt should be null (zero), got %v", m.MinAmt)
	}
}

const accessoryCheckoutRowsFixture = `{
	"total": 2,
	"rows": [
		{"id": 21, "assigned_to": {"id": 3, "type": "user", "name": "Max"}},
		{"id": 22, "assigned_to": {"id": 4, "type": "user", "name": "Moritz"}}
	]
}`

func TestAccessoryCheckoutRowsDecode(t *testing.T) {
	var list inventoryapi.AccessoryCheckoutList
	if err := json.Unmarshal([]byte(accessoryCheckoutRowsFixture), &list); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	if len(list.Rows) != 2 {
		t.Fatalf("rows = %d", len(list.Rows))
	}
	if list.Rows[0].Id != 21 || list.Rows[0].AssignedTo == nil || list.Rows[0].AssignedTo.Id != 3 {
		t.Errorf("row 0 = %+v", list.Rows[0])
	}
	if list.Rows[1].AssignedTo.Type != "user" {
		t.Errorf("row 1 type = %q", list.Rows[1].AssignedTo.Type)
	}
}

const componentAssetRowsFixture = `{
	"total": 1,
	"rows": [
		{"assigned_pivot_id": 31, "id": 44, "qty": "2", "name": "some-asset"}
	]
}`

func TestComponentAssetRowsDecode(t *testing.T) {
	var list inventoryapi.ComponentAssetList
	if err := json.Unmarshal([]byte(componentAssetRowsFixture), &list); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	if len(list.Rows) != 1 {
		t.Fatalf("rows = %d", len(list.Rows))
	}
	row := list.Rows[0]
	if row.AssignedPivotId != 31 || row.Id != 44 {
		t.Errorf("row = %+v", row)
	}
	// qty arrives as string — FlexInt must decode it.
	if int64(row.Qty) != 2 {
		t.Errorf("qty = %v", row.Qty)
	}
}

func TestAccessoryToBodyClearSemantics(t *testing.T) {
	m := AccessoryResourceModel{
		Name:           types.StringValue("Dock"),
		Qty:            types.Int64Value(5),
		CategoryID:     types.Int64Value(4),
		ManufacturerID: types.Int64Null(),    // cleared ref -> explicit JSON null
		SupplierID:     types.Int64Unknown(), // unknown -> omitted
		PurchaseDate:   types.StringNull(),   // cleared date -> explicit JSON null
		MinAmt:         types.Int64Null(),    // cleared -> explicit JSON null
		ModelNumber:    types.StringNull(),   // cleared string -> ""
	}
	body := m.toBody()

	if body["name"] != "Dock" || body["qty"] != int64(5) || body["category_id"] != int64(4) {
		t.Errorf("required fields wrong: %v", body)
	}
	if v, ok := body["manufacturer_id"]; !ok || v != nil {
		t.Errorf("manufacturer_id must be present and nil, got %v (present=%v)", v, ok)
	}
	if _, ok := body["supplier_id"]; ok {
		t.Errorf("unknown supplier_id must be omitted")
	}
	if v, ok := body["purchase_date"]; !ok || v != nil {
		t.Errorf("purchase_date must be present and nil, got %v (present=%v)", v, ok)
	}
	if v, ok := body["min_amt"]; !ok || v != nil {
		t.Errorf("min_amt must be present and nil, got %v (present=%v)", v, ok)
	}
	if body["model_number"] != "" {
		t.Errorf("null model_number must be sent as empty string, got %v", body["model_number"])
	}
}

func TestComponentToBodyClearSemantics(t *testing.T) {
	m := ComponentResourceModel{
		Name:       types.StringValue("DIMM"),
		Qty:        types.Int64Value(8),
		CategoryID: types.Int64Value(5),
		Serial:     types.StringNull(), // cleared string -> ""
		SupplierID: types.Int64Null(),  // cleared ref -> explicit JSON null
	}
	body := m.toBody()

	if body["serial"] != "" {
		t.Errorf("null serial must be sent as empty string, got %v", body["serial"])
	}
	if v, ok := body["supplier_id"]; !ok || v != nil {
		t.Errorf("supplier_id must be present and nil, got %v (present=%v)", v, ok)
	}
}

// Explicitly configured zero values must survive the read-back mapping
// (issue #17): the API serializes unset and explicit ""/0 identically, so
// the mappers keep the prior explicit zero and only fall back to null when
// the prior was null.
func TestAccessoryFromAPIKeepsExplicitZeroValues(t *testing.T) {
	var api inventoryapi.Accessory
	if err := json.Unmarshal([]byte(`{"id": 2, "name": "z", "qty": 1, "min_qty": 0, "model_number": ""}`), &api); err != nil {
		t.Fatal(err)
	}
	var m AccessoryResourceModel
	m.MinAmt = types.Int64Value(0)
	m.ModelNumber = types.StringValue("")
	m.Notes = types.StringNull()
	m.fromAPI(&api)
	if m.MinAmt.IsNull() || m.MinAmt.ValueInt64() != 0 {
		t.Errorf("explicit min_amt=0 lost: %v", m.MinAmt)
	}
	if m.ModelNumber.IsNull() || m.ModelNumber.ValueString() != "" {
		t.Errorf("explicit empty model_number lost: %v", m.ModelNumber)
	}
	if !m.Notes.IsNull() {
		t.Errorf("unset notes should stay null: %v", m.Notes)
	}
}
