// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package people

// Pure unit tests for the API-object -> TF-state mapping and the request body
// builders. These run without TF_ACC and without any network access; the JSON
// fixtures mirror real v8.0.4 responses.

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	peopleapi "github.com/timrabl/terraform-provider-snipeit/internal/api/people"
)

const userFixture = `{
	"id": 7,
	"username": "mmustermann",
	"first_name": "Max",
	"last_name": "Mustermann",
	"email": "",
	"employee_num": null,
	"jobtitle": "Tester",
	"phone": null,
	"notes": "managed by terraform",
	"company": {"id": 3, "name": "ACME GmbH"},
	"department": null,
	"location": {"id": 0, "name": ""},
	"manager": null,
	"activated": true,
	"groups": {"total": 2, "rows": [{"id": 4, "name": "admins"}, {"id": 9, "name": "it"}]},
	"created_at": {"datetime": "2026-08-22 13:31:07", "formatted": "2026-08-22 01:31 PM"}
}`

func TestUserFromAPI(t *testing.T) {
	var api peopleapi.User
	if err := json.Unmarshal([]byte(userFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	m := UserResourceModel{Password: types.StringValue("keep-me")}
	if err := m.fromAPI(context.Background(), &api); err != nil {
		t.Fatalf("fromAPI: %v", err)
	}

	if m.ID.ValueInt64() != 7 || m.Username.ValueString() != "mmustermann" {
		t.Errorf("id/username = %v/%v", m.ID, m.Username)
	}
	if m.FirstName.ValueString() != "Max" || m.LastName.ValueString() != "Mustermann" {
		t.Errorf("first/last = %v/%v", m.FirstName, m.LastName)
	}
	// "" and null both map to TF null.
	if !m.Email.IsNull() {
		t.Errorf("email should be null, got %v", m.Email)
	}
	if !m.EmployeeNum.IsNull() {
		t.Errorf("employee_num should be null, got %v", m.EmployeeNum)
	}
	if m.CompanyID.ValueInt64() != 3 {
		t.Errorf("company_id = %v", m.CompanyID)
	}
	// Absent ref and zero ref both map to null.
	if !m.DepartmentID.IsNull() || !m.LocationID.IsNull() || !m.ManagerID.IsNull() {
		t.Errorf("department/location/manager should be null, got %v/%v/%v",
			m.DepartmentID, m.LocationID, m.ManagerID)
	}
	if !m.Activated.ValueBool() {
		t.Errorf("activated = %v", m.Activated)
	}
	// The write-only password must never be touched by fromAPI.
	if m.Password.ValueString() != "keep-me" {
		t.Errorf("password must stay untouched, got %v", m.Password)
	}
	var ids []int64
	if diags := m.Groups.ElementsAs(context.Background(), &ids, false); diags.HasError() {
		t.Fatalf("groups: %v", diags)
	}
	if len(ids) != 2 || ids[0] != 4 || ids[1] != 9 {
		t.Errorf("groups = %v", ids)
	}
}

func TestUserFromAPINoGroups(t *testing.T) {
	var api peopleapi.User
	if err := json.Unmarshal([]byte(`{"id":1,"username":"x","first_name":"X","activated":false,"groups":null}`), &api); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	var m UserResourceModel
	if err := m.fromAPI(context.Background(), &api); err != nil {
		t.Fatalf("fromAPI: %v", err)
	}
	if !m.Groups.IsNull() {
		t.Errorf("groups should be null when absent, got %v", m.Groups)
	}
}

func TestUserToBodyPasswordSemantics(t *testing.T) {
	groups, diags := types.SetValueFrom(context.Background(), types.Int64Type, []int64{4})
	if diags.HasError() {
		t.Fatalf("set: %v", diags)
	}
	m := UserResourceModel{
		Username:  types.StringValue("mmustermann"),
		FirstName: types.StringValue("Max"),
		Password:  types.StringValue("s3cret"),
		Email:     types.StringNull(),   // cleared string -> ""
		CompanyID: types.Int64Null(),    // cleared ref -> explicit JSON null
		ManagerID: types.Int64Unknown(), // unknown -> omitted
		Groups:    groups,
	}
	body, err := m.toBody(context.Background())
	if err != nil {
		t.Fatalf("toBody: %v", err)
	}

	if body["password"] != "s3cret" || body["password_confirmation"] != "s3cret" {
		t.Errorf("password pair = %v/%v", body["password"], body["password_confirmation"])
	}
	if body["email"] != "" {
		t.Errorf("null email must be sent as empty string, got %v", body["email"])
	}
	if v, ok := body["company_id"]; !ok || v != nil {
		t.Errorf("company_id must be present and nil, got %v (present=%v)", v, ok)
	}
	if _, ok := body["manager_id"]; ok {
		t.Errorf("unknown manager_id must be omitted")
	}
	if got, ok := body["groups"].([]int64); !ok || len(got) != 1 || got[0] != 4 {
		t.Errorf("groups = %v", body["groups"])
	}
}

func TestUserToBodyNoPasswordWhenNull(t *testing.T) {
	m := UserResourceModel{
		Username:  types.StringValue("x"),
		FirstName: types.StringValue("X"),
		Password:  types.StringNull(),
		Groups:    types.SetNull(types.Int64Type),
	}
	body, err := m.toBody(context.Background())
	if err != nil {
		t.Fatalf("toBody: %v", err)
	}
	if _, ok := body["password"]; ok {
		t.Errorf("null password must not be sent")
	}
	if _, ok := body["password_confirmation"]; ok {
		t.Errorf("null password_confirmation must not be sent")
	}
	// Null groups clear membership with an explicit null.
	if v, ok := body["groups"]; !ok || v != nil {
		t.Errorf("null groups must be present and nil, got %v (present=%v)", v, ok)
	}
}

const groupFixture = `{
	"id": 4,
	"name": "admins",
	"notes": null,
	"permissions": {"assets.view": "1", "assets.create": "0"}
}`

func TestGroupFromAPIConfiguredPermissions(t *testing.T) {
	var api peopleapi.Group
	if err := json.Unmarshal([]byte(groupFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	configured, diags := types.MapValueFrom(context.Background(), types.StringType,
		map[string]string{"assets.view": "1", "assets.create": "0"})
	if diags.HasError() {
		t.Fatalf("map: %v", diags)
	}
	m := GroupResourceModel{Permissions: configured}
	if err := m.fromAPI(context.Background(), &api); err != nil {
		t.Fatalf("fromAPI: %v", err)
	}

	if m.Name.ValueString() != "admins" {
		t.Errorf("name = %v", m.Name)
	}
	perms := map[string]string{}
	if diags := m.Permissions.ElementsAs(context.Background(), &perms, false); diags.HasError() {
		t.Fatalf("perms: %v", diags)
	}
	if perms["assets.view"] != "1" || perms["assets.create"] != "0" {
		t.Errorf("permissions = %v", perms)
	}
}

func TestGroupFromAPIUnconfiguredPermissionsStayNull(t *testing.T) {
	// A group created without permissions gets a server-generated full
	// default map; reflecting it against a null config would drift.
	var api peopleapi.Group
	if err := json.Unmarshal([]byte(`{"id":4,"name":"g","permissions":{"assets.view":"0","assets.create":"0"}}`), &api); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	m := GroupResourceModel{Permissions: types.MapNull(types.StringType)}
	if err := m.fromAPI(context.Background(), &api); err != nil {
		t.Fatalf("fromAPI: %v", err)
	}
	if !m.Permissions.IsNull() {
		t.Errorf("unconfigured permissions must stay null, got %v", m.Permissions)
	}
}

func TestGroupToBody(t *testing.T) {
	configured, diags := types.MapValueFrom(context.Background(), types.StringType,
		map[string]string{"assets.view": "1"})
	if diags.HasError() {
		t.Fatalf("map: %v", diags)
	}
	m := GroupResourceModel{
		Name:        types.StringValue("admins"),
		Notes:       types.StringNull(),
		Permissions: configured,
	}
	body, err := m.toBody(context.Background())
	if err != nil {
		t.Fatalf("toBody: %v", err)
	}
	// name is always present (required by group PATCH even for
	// permission-only changes).
	if body["name"] != "admins" {
		t.Errorf("name = %v", body["name"])
	}
	if body["notes"] != "" {
		t.Errorf("null notes must be sent as empty string, got %v", body["notes"])
	}
	perms, ok := body["permissions"].(map[string]string)
	if !ok || perms["assets.view"] != "1" {
		t.Errorf("permissions = %v", body["permissions"])
	}
}

func TestGroupToBodyOmitsNullPermissions(t *testing.T) {
	m := GroupResourceModel{
		Name:        types.StringValue("g"),
		Permissions: types.MapNull(types.StringType),
	}
	body, err := m.toBody(context.Background())
	if err != nil {
		t.Fatalf("toBody: %v", err)
	}
	if _, ok := body["permissions"]; ok {
		t.Errorf("null permissions must be omitted (server keeps stored map)")
	}
}
