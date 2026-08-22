// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package operations

// Pure unit tests for the API-object -> TF-state mapping and the request body
// builders. These run without TF_ACC and without any network access; the JSON
// fixtures mirror real v8.0.4 responses.

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	operationsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/operations"
)

const maintenanceFixture = `{
	"id": 4,
	"asset": {"id": 11, "name": "LPT-0001"},
	"supplier": {"id": 6, "name": "ACME Repairs"},
	"asset_maintenance_type": "Repair",
	"title": "Board swap",
	"start_date": {"date": "2026-08-05", "formatted": "Tue Aug 05, 2026"},
	"completion_date": null,
	"is_warranty": "1",
	"cost": "1,500.50",
	"notes": null
}`

func TestMaintenanceFromAPI(t *testing.T) {
	var api operationsapi.Maintenance
	if err := json.Unmarshal([]byte(maintenanceFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m MaintenanceResourceModel
	m.fromAPI(&api)

	if m.Cost.ValueString() != "1500.50" {
		t.Errorf("cost should be normalized, got %v", m.Cost)
	}

	if m.ID.ValueInt64() != 4 {
		t.Errorf("id = %v", m.ID)
	}
	if m.AssetID.ValueInt64() != 11 || m.SupplierID.ValueInt64() != 6 {
		t.Errorf("asset/supplier = %v/%v", m.AssetID, m.SupplierID)
	}
	if m.Type.ValueString() != "Repair" || m.Title.ValueString() != "Board swap" {
		t.Errorf("type/title = %v/%v", m.Type, m.Title)
	}
	if m.StartDate.ValueString() != "2026-08-05" {
		t.Errorf("start_date must be the plain date, got %v", m.StartDate)
	}
	if !m.CompletionDate.IsNull() {
		t.Errorf("null completion_date must map to TF null, got %v", m.CompletionDate)
	}
	if !m.IsWarranty.ValueBool() {
		t.Errorf("is_warranty \"1\" must decode to true")
	}
	if !m.Notes.IsNull() {
		t.Errorf("null notes must map to TF null, got %v", m.Notes)
	}
}

func TestMaintenanceToBodyClearSemantics(t *testing.T) {
	m := MaintenanceResourceModel{
		AssetID:        types.Int64Value(11),
		SupplierID:     types.Int64Value(6),
		Type:           types.StringValue("Repair"),
		Title:          types.StringValue("Board swap"),
		StartDate:      types.StringValue("2026-08-05"),
		CompletionDate: types.StringNull(),  // cleared date -> explicit JSON null
		Notes:          types.StringNull(),  // cleared string -> ""
		IsWarranty:     types.BoolUnknown(), // unknown computed -> omitted
	}
	body := m.toBody()

	if body["asset_maintenance_type"] != "Repair" {
		t.Errorf("asset_maintenance_type = %v", body["asset_maintenance_type"])
	}
	if v, ok := body["completion_date"]; !ok || v != nil {
		t.Errorf("cleared completion_date must be present and nil, got %v (present=%v)", v, ok)
	}
	if body["notes"] != "" {
		t.Errorf("null notes must be sent as empty string, got %v", body["notes"])
	}
	if _, ok := body["is_warranty"]; ok {
		t.Errorf("unknown is_warranty must be omitted")
	}
}

// activityFixture mirrors a real GET /reports/activity row set, including a
// checkout row with a target and a bare create row without one.
const activityFixture = `{
	"total": 2,
	"rows": [
		{"id": 91, "action_type": "checkout",
		 "item": {"id": 11, "name": "LPT-0001", "type": "asset"},
		 "target": {"id": 3, "name": "Max Example", "type": "user"},
		 "admin": {"id": 1, "name": "Max"},
		 "note": "handover",
		 "action_date": {"datetime": "2026-08-22 12:00:00", "formatted": "..."}},
		{"id": 90, "action_type": "create",
		 "item": {"id": 5, "name": "probe-mfg", "type": "manufacturer"},
		 "target": null, "admin": null, "note": null, "action_date": null}
	]
}`

func TestActivityListDecode(t *testing.T) {
	var list operationsapi.ActivityList
	if err := json.Unmarshal([]byte(activityFixture), &list); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	if list.Total != 2 || len(list.Rows) != 2 {
		t.Fatalf("total/rows = %d/%d", list.Total, len(list.Rows))
	}
	r0 := list.Rows[0]
	if r0.Target == nil || r0.Target.Type != "user" || r0.Target.Id != 3 {
		t.Errorf("row 0 target = %+v", r0.Target)
	}
	if r0.ActionDate == nil || r0.ActionDate.DateTime != "2026-08-22 12:00:00" {
		t.Errorf("row 0 action_date = %+v", r0.ActionDate)
	}
	r1 := list.Rows[1]
	if r1.Target != nil || r1.Admin != nil || r1.ActionDate != nil {
		t.Errorf("row 1 nils not preserved: %+v", r1)
	}
}

const auditFixture = `{
	"total": 1,
	"rows": [{"id": 11, "asset_tag": "LPT-0001", "name": null}]
}`

func TestAuditListDecode(t *testing.T) {
	var list operationsapi.AuditList
	if err := json.Unmarshal([]byte(auditFixture), &list); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	if list.Total != 1 || list.Rows[0].AssetTag != "LPT-0001" || list.Rows[0].Name != nil {
		t.Errorf("audit list = %+v", list)
	}
}

// Explicitly configured zero values must survive the read-back mapping
// (issue #17): the API serializes unset and explicit ""/0 identically, so
// the mappers keep the prior explicit zero and only fall back to null when
// the prior was null.
func TestMaintenanceFromAPIKeepsExplicitZeroValues(t *testing.T) {
	var api operationsapi.Maintenance
	if err := json.Unmarshal([]byte(`{"id": 8, "title": "z", "notes": ""}`), &api); err != nil {
		t.Fatal(err)
	}
	var m MaintenanceResourceModel
	m.Notes = types.StringValue("")
	m.fromAPI(&api)
	if m.Notes.IsNull() || m.Notes.ValueString() != "" {
		t.Errorf("explicit empty notes lost: %v", m.Notes)
	}
}
