// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package tfutil

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

func TestCanonicalMoney(t *testing.T) {
	cases := map[string]string{
		"1,234.50":      "1234.5",
		"1234.50":       "1234.5",
		"1234.5":        "1234.5",
		"7.00":          "7",
		"12,345,678.99": "12345678.99",
		"0.10":          "0.1",
		"1000":          "1000",
	}
	for in, want := range cases {
		if got := CanonicalMoney(in); got != want {
			t.Errorf("CanonicalMoney(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMoneySemanticEquals(t *testing.T) {
	equal := [][2]string{
		{"1234.5", "1234.50"},
		{"1234.5", "1,234.50"},
		{"7", "7.00"},
	}
	for _, pair := range equal {
		ok, diags := NewMoneyValue(pair[0]).StringSemanticEquals(context.Background(), NewMoneyValue(pair[1]))
		if diags.HasError() || !ok {
			t.Errorf("expected %q semantically equal to %q", pair[0], pair[1])
		}
	}
	ok, _ := NewMoneyValue("1234.5").StringSemanticEquals(context.Background(), NewMoneyValue("1234.51"))
	if ok {
		t.Error("expected 1234.5 != 1234.51")
	}
}

func TestMoneyValidateAttribute(t *testing.T) {
	valid := []string{"1234.50", "1,234.50", "7", "0.1", "12,345,678.99"}
	invalid := []string{"1234,56", "12.34.56", "abc", "-5", "1.234,56", "1,23.45"}

	for _, s := range valid {
		resp := &xattr.ValidateAttributeResponse{}
		NewMoneyValue(s).ValidateAttribute(context.Background(), xattr.ValidateAttributeRequest{Path: path.Root("purchase_cost")}, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("expected %q valid", s)
		}
	}
	for _, s := range invalid {
		resp := &xattr.ValidateAttributeResponse{}
		NewMoneyValue(s).ValidateAttribute(context.Background(), xattr.ValidateAttributeRequest{Path: path.Root("purchase_cost")}, resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("expected %q invalid", s)
		}
	}
}

func TestBodyMoneyAndStateMoneyPtr(t *testing.T) {
	body := map[string]any{}
	BodyMoney(body, "purchase_cost", NewMoneyValue("1,234.50"))
	if body["purchase_cost"] != "1234.50" {
		t.Errorf("expected normalized send, got %v", body["purchase_cost"])
	}
	BodyMoney(body, "cleared", NewMoneyNull())
	if v, ok := body["cleared"]; !ok || v != nil {
		t.Errorf("expected explicit null for cleared money, got %v (present=%v)", v, ok)
	}

	api := "1,234.50"
	if got := StateMoneyPtr(&api); got.ValueString() != "1234.50" {
		t.Errorf("StateMoneyPtr normalized wrong: %q", got.ValueString())
	}
	if !StateMoneyPtr(nil).IsNull() {
		t.Error("nil should map to null")
	}
	empty := ""
	if !StateMoneyPtr(&empty).IsNull() {
		t.Error("empty should map to null")
	}
}

func TestStateMoneyPtrClearAware(t *testing.T) {
	api := "1,234.50"

	// Prior null (user cleared or never set): map to null even though the API
	// still echoes a value (Snipe-IT 8.7 ignores the clear).
	if got := StateMoneyPtrClearAware(&api, NewMoneyNull()); !got.IsNull() {
		t.Errorf("prior null should map to null, got %q", got.ValueString())
	}
	// Prior set: map the API value normally (drift is still detected).
	if got := StateMoneyPtrClearAware(&api, NewMoneyValue("0.00")); got.ValueString() != "1234.50" {
		t.Errorf("prior set should map API value, got %q", got.ValueString())
	}
	// Prior set but API cleared: maps to null (genuine clear detected).
	if got := StateMoneyPtrClearAware(nil, NewMoneyValue("1234.50")); !got.IsNull() {
		t.Errorf("API-cleared with prior set should be null, got %q", got.ValueString())
	}
}

func TestStateDateClearAware(t *testing.T) {
	date := &client.Date{Date: "2026-01-10"}

	// Prior null: null regardless of the echoed API date.
	if got := StateDateClearAware(date, types.StringNull()); !got.IsNull() {
		t.Errorf("prior null should map to null, got %q", got.ValueString())
	}
	// Prior set: map the API date.
	if got := StateDateClearAware(date, types.StringValue("2000-01-01")); got.ValueString() != "2026-01-10" {
		t.Errorf("prior set should map API date, got %q", got.ValueString())
	}
	// Prior set but API empty/nil: null.
	if got := StateDateClearAware(nil, types.StringValue("2026-01-10")); !got.IsNull() {
		t.Errorf("nil API date with prior set should be null, got %q", got.ValueString())
	}
}
