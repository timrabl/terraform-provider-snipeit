// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package tfutil

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStateStringKeep(t *testing.T) {
	cases := []struct {
		name  string
		api   string
		prior types.String
		want  types.String
	}{
		{"value stays value", "x", types.StringNull(), types.StringValue("x")},
		{"empty with null prior stays null", "", types.StringNull(), types.StringNull()},
		{"empty with explicit empty prior keeps empty", "", types.StringValue(""), types.StringValue("")},
		{"empty with non-empty prior maps null (real drift)", "", types.StringValue("old"), types.StringNull()},
		{"empty with unknown prior maps null", "", types.StringUnknown(), types.StringNull()},
	}
	for _, c := range cases {
		if got := StateStringKeep(c.api, c.prior); !got.Equal(c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestStateStringPtrKeep(t *testing.T) {
	empty, val := "", "x"
	cases := []struct {
		name  string
		api   *string
		prior types.String
		want  types.String
	}{
		{"value stays value", &val, types.StringNull(), types.StringValue("x")},
		{"nil with null prior stays null", nil, types.StringNull(), types.StringNull()},
		{"nil with explicit empty prior keeps empty", nil, types.StringValue(""), types.StringValue("")},
		{"empty with explicit empty prior keeps empty", &empty, types.StringValue(""), types.StringValue("")},
		{"empty with non-empty prior maps null", &empty, types.StringValue("old"), types.StringNull()},
	}
	for _, c := range cases {
		if got := StateStringPtrKeep(c.api, c.prior); !got.Equal(c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestStateOptIntKeep(t *testing.T) {
	cases := []struct {
		name  string
		api   int64
		prior types.Int64
		want  types.Int64
	}{
		{"value stays value", 7, types.Int64Null(), types.Int64Value(7)},
		{"zero with null prior stays null", 0, types.Int64Null(), types.Int64Null()},
		{"zero with explicit zero prior keeps zero", 0, types.Int64Value(0), types.Int64Value(0)},
		{"zero with non-zero prior maps null (real drift)", 0, types.Int64Value(5), types.Int64Null()},
		{"zero with unknown prior maps null", 0, types.Int64Unknown(), types.Int64Null()},
	}
	for _, c := range cases {
		if got := StateOptIntKeep(c.api, c.prior); !got.Equal(c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
