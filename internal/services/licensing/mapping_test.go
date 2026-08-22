package licensing

// Pure unit tests for the API-object -> TF-state mapping and the request body
// builders. These run without TF_ACC and without any network access; the JSON
// fixtures mirror real v8.0.4 responses.

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	licensingapi "github.com/timrabl/terraform-provider-snipeit/internal/api/licensing"
)

// licenseFixture mirrors a real GET /licenses/{id} response including the
// decorated values: seats as plain int, booleans as 0/1, dates as objects,
// and the serial echoed back under product_key.
const licenseFixture = `{
	"id": 7,
	"name": "Office Suite",
	"seats": 5,
	"free_seats_count": "4",
	"category": {"id": 12, "name": "Software"},
	"company": null,
	"manufacturer": {"id": 0, "name": ""},
	"supplier": null,
	"order_number": "ORD-42",
	"purchase_order": "",
	"purchase_date": {"date": "2026-01-10", "formatted": "Sat Jan 10, 2026"},
	"expiration_date": null,
	"termination_date": null,
	"license_name": "Example Licensee",
	"license_email": null,
	"product_key": "AAAA-BBBB-CCCC",
	"reassignable": 1,
	"maintained": "0",
	"notes": null
}`

func TestLicenseFromAPI(t *testing.T) {
	var api licensingapi.License
	if err := json.Unmarshal([]byte(licenseFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m LicenseResourceModel
	m.fromAPI(&api)

	if m.ID.ValueInt64() != 7 || m.Name.ValueString() != "Office Suite" {
		t.Errorf("id/name = %v/%v", m.ID, m.Name)
	}
	if m.Seats.ValueInt64() != 5 {
		t.Errorf("seats = %v", m.Seats)
	}
	if m.CategoryID.ValueInt64() != 12 {
		t.Errorf("category_id = %v", m.CategoryID)
	}
	// Absent ref and zero ref both map to TF null.
	if !m.CompanyID.IsNull() || !m.ManufacturerID.IsNull() || !m.SupplierID.IsNull() {
		t.Errorf("company/manufacturer/supplier ids should be null, got %v/%v/%v",
			m.CompanyID, m.ManufacturerID, m.SupplierID)
	}
	if m.OrderNumber.ValueString() != "ORD-42" {
		t.Errorf("order_number = %v", m.OrderNumber)
	}
	// "" maps to null.
	if !m.PurchaseOrder.IsNull() {
		t.Errorf("purchase_order should be null, got %v", m.PurchaseOrder)
	}
	// Nested date object -> bare date string; null date -> null.
	if m.PurchaseDate.ValueString() != "2026-01-10" {
		t.Errorf("purchase_date = %v", m.PurchaseDate)
	}
	if !m.ExpirationDate.IsNull() {
		t.Errorf("expiration_date should be null, got %v", m.ExpirationDate)
	}
	// product_key (GET) feeds the serial attribute (write).
	if m.Serial.ValueString() != "AAAA-BBBB-CCCC" {
		t.Errorf("serial = %v", m.Serial)
	}
	// 0/1 and "0"/"1" booleans decode via FlexBool.
	if !m.Reassignable.ValueBool() {
		t.Errorf("reassignable should be true, got %v", m.Reassignable)
	}
	if m.Maintained.ValueBool() {
		t.Errorf("maintained should be false, got %v", m.Maintained)
	}
}

const seatAssignedFixture = `{
	"id": 16,
	"license_id": 7,
	"assigned_user": null,
	"assigned_asset": {"id": 24, "name": "(hw-0001) - Test Model"}
}`

const seatFreeFixture = `{
	"id": 17,
	"license_id": 7,
	"assigned_user": null,
	"assigned_asset": null
}`

func TestLicenseSeatFree(t *testing.T) {
	var assigned, free licensingapi.LicenseSeat
	if err := json.Unmarshal([]byte(seatAssignedFixture), &assigned); err != nil {
		t.Fatalf("unmarshal assigned fixture: %v", err)
	}
	if err := json.Unmarshal([]byte(seatFreeFixture), &free); err != nil {
		t.Fatalf("unmarshal free fixture: %v", err)
	}

	if assigned.Free() {
		t.Error("asset-assigned seat must not be free")
	}
	if !free.Free() {
		t.Error("unassigned seat must be free")
	}
	if assigned.AssignedAsset.IDOrZero() != 24 {
		t.Errorf("assigned asset id = %d", assigned.AssignedAsset.IDOrZero())
	}
}

func TestLicenseToBodyClearSemantics(t *testing.T) {
	m := LicenseResourceModel{
		Name:           types.StringValue("Office Suite"),
		Seats:          types.Int64Value(5),
		CategoryID:     types.Int64Value(12),
		CompanyID:      types.Int64Null(),    // cleared ref -> explicit JSON null
		SupplierID:     types.Int64Unknown(), // unknown -> omitted
		PurchaseDate:   types.StringNull(),   // cleared date -> explicit JSON null
		ExpirationDate: types.StringValue("2027-01-10"),
		Serial:         types.StringNull(),  // cleared string -> ""
		Maintained:     types.BoolUnknown(), // unknown bool -> omitted
		Reassignable:   types.BoolValue(true),
	}
	body := m.toBody()

	if body["name"] != "Office Suite" || body["seats"] != int64(5) || body["category_id"] != int64(12) {
		t.Errorf("required fields wrong: %v", body)
	}
	if v, ok := body["company_id"]; !ok || v != nil {
		t.Errorf("company_id must be present and nil, got %v (present=%v)", v, ok)
	}
	if _, ok := body["supplier_id"]; ok {
		t.Error("unknown supplier_id must be omitted")
	}
	if v, ok := body["purchase_date"]; !ok || v != nil {
		t.Errorf("purchase_date must be present and nil, got %v (present=%v)", v, ok)
	}
	if body["expiration_date"] != "2027-01-10" {
		t.Errorf("expiration_date = %v", body["expiration_date"])
	}
	if body["serial"] != "" {
		t.Errorf("null serial must be sent as empty string, got %v", body["serial"])
	}
	if _, ok := body["maintained"]; ok {
		t.Error("unknown maintained must be omitted")
	}
	if body["reassignable"] != true {
		t.Errorf("reassignable = %v", body["reassignable"])
	}
}

func TestSeatAssignmentBodyAlwaysCarriesBothKeys(t *testing.T) {
	m := LicenseSeatResourceModel{
		AssignedToUserID:  types.Int64Value(1),
		AssignedToAssetID: types.Int64Null(),
	}
	body := m.assignmentBody()

	if body["assigned_to"] != int64(1) {
		t.Errorf("assigned_to = %v", body["assigned_to"])
	}
	if v, ok := body["asset_id"]; !ok || v != nil {
		t.Errorf("asset_id must be present and nil, got %v (present=%v)", v, ok)
	}
}
