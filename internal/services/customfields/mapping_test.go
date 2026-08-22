// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package customfields

// Pure unit tests for the API-object -> TF-state mapping and the request body
// builders. These run without TF_ACC and without any network access; the JSON
// fixtures mirror real v8.0.4 responses.

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	customfieldsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/customfields"
)

// fieldFixture mirrors a real GET /fields/{id} response: the element comes
// back under "type", help_text/show_in_email are absent (write-only).
const fieldFixture = `{
	"id": 7,
	"name": "MAC Address",
	"db_column_name": "_snipeit_mac_address_7",
	"format": "MAC",
	"field_values": null,
	"field_values_array": null,
	"type": "text",
	"required": false
}`

func TestFieldFromAPI(t *testing.T) {
	var api customfieldsapi.Field
	if err := json.Unmarshal([]byte(fieldFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	m := FieldResourceModel{
		// Configured write-only values must survive a read untouched.
		HelpText:    types.StringValue("keep me"),
		ShowInEmail: types.BoolValue(true),
	}
	m.fromAPI(&api)

	if m.ID.ValueInt64() != 7 {
		t.Errorf("id = %v", m.ID)
	}
	if m.Element.ValueString() != "text" {
		t.Errorf("element must come from the JSON key \"type\", got %v", m.Element)
	}
	if m.Format.ValueString() != "MAC" {
		t.Errorf("format = %v", m.Format)
	}
	if m.DBColumnName.ValueString() != "_snipeit_mac_address_7" {
		t.Errorf("db_column_name = %v", m.DBColumnName)
	}
	if !m.FieldValues.IsNull() {
		t.Errorf("null field_values must map to TF null, got %v", m.FieldValues)
	}
	if m.HelpText.ValueString() != "keep me" || !m.ShowInEmail.ValueBool() {
		t.Errorf("write-only attributes must be left untouched by fromAPI")
	}
}

// fieldsetFixture mirrors a real GET /fieldsets/{id} response with the member
// fields embedded under fields.rows ("required" arrives as a string bool).
const fieldsetFixture = `{
	"id": 2,
	"name": "Laptop Fields",
	"fields": {
		"total": 2,
		"rows": [
			{"id": 7, "name": "MAC Address", "db_column_name": "_snipeit_mac_address_7",
			 "format": "MAC", "type": "text", "required": "1"},
			{"id": 9, "name": "Warranty Vendor", "db_column_name": "_snipeit_warranty_vendor_9",
			 "format": "ANY", "type": "listbox", "required": 0}
		]
	}
}`

func TestFieldsetFromAPI(t *testing.T) {
	var api customfieldsapi.Fieldset
	if err := json.Unmarshal([]byte(fieldsetFixture), &api); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	var m FieldsetResourceModel
	m.fromAPI(&api)

	if m.ID.ValueInt64() != 2 || m.Name.ValueString() != "Laptop Fields" {
		t.Errorf("id/name = %v/%v", m.ID, m.Name)
	}
	if len(api.Fields.Rows) != 2 {
		t.Fatalf("embedded rows = %d", len(api.Fields.Rows))
	}
	if api.Fields.Rows[0].Element != "text" || !bool(api.Fields.Rows[0].Required) {
		t.Errorf("row 0 element/required = %v/%v", api.Fields.Rows[0].Element, api.Fields.Rows[0].Required)
	}
	if bool(api.Fields.Rows[1].Required) {
		t.Errorf("row 1 required (0) must decode to false")
	}
}

func TestFieldToBodySemantics(t *testing.T) {
	m := FieldResourceModel{
		Name:        types.StringValue("Serial 2"),
		Element:     types.StringValue("text"),
		Format:      types.StringUnknown(), // unknown computed -> omitted
		FieldValues: types.StringNull(),    // cleared string -> ""
		HelpText:    types.StringValue("hint"),
		ShowInEmail: types.BoolNull(), // unset Optional bool -> omitted
	}
	body := m.toBody()

	if body["name"] != "Serial 2" || body["element"] != "text" {
		t.Errorf("name/element = %v/%v", body["name"], body["element"])
	}
	if _, ok := body["format"]; ok {
		t.Errorf("unknown format must be omitted")
	}
	if body["field_values"] != "" {
		t.Errorf("null field_values must be sent as empty string, got %v", body["field_values"])
	}
	if body["help_text"] != "hint" {
		t.Errorf("help_text = %v", body["help_text"])
	}
	if _, ok := body["show_in_email"]; ok {
		t.Errorf("null show_in_email must be omitted")
	}
}

func TestSplitFieldAssociationID(t *testing.T) {
	for _, tc := range []struct {
		in          string
		field, fset int64
		expectError bool
	}{
		{"7:2", 7, 2, false},
		{"7/2", 7, 2, false},
		{"7", 0, 0, true},
		{"a:b", 0, 0, true},
	} {
		f, s, err := splitFieldAssociationID(tc.in)
		if tc.expectError {
			if err == nil {
				t.Errorf("%q: expected error", tc.in)
			}
			continue
		}
		if err != nil || f != tc.field || s != tc.fset {
			t.Errorf("%q: got %d/%d err=%v", tc.in, f, s, err)
		}
	}
}

// Explicitly configured zero values must survive the read-back mapping
// (issue #17): the API serializes unset and explicit ""/0 identically, so
// the mappers keep the prior explicit zero and only fall back to null when
// the prior was null.
func TestFieldFromAPIKeepsExplicitZeroValues(t *testing.T) {
	var api customfieldsapi.Field
	if err := json.Unmarshal([]byte(`{"id": 6, "name": "z", "type": "text", "field_values": ""}`), &api); err != nil {
		t.Fatal(err)
	}
	var m FieldResourceModel
	m.FieldValues = types.StringValue("")
	m.fromAPI(&api)
	if m.FieldValues.IsNull() || m.FieldValues.ValueString() != "" {
		t.Errorf("explicit empty field_values lost: %v", m.FieldValues)
	}
}
