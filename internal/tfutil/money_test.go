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
