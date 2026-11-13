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

func TestBodyStringSparse(t *testing.T) {
	cases := []struct {
		name    string
		plan    types.String
		prior   types.String
		present bool
		want    any
	}{
		{"set on create", types.StringValue("SN-1"), types.StringNull(), true, "SN-1"},
		{"set on update", types.StringValue("SN-2"), types.StringValue("SN-1"), true, "SN-2"},
		{"explicit empty is sent", types.StringValue(""), types.StringNull(), true, ""},
		{"unset on create is omitted", types.StringNull(), types.StringNull(), false, nil},
		{"unset and never set is omitted on update", types.StringNull(), types.StringNull(), false, nil},
		{"unset with empty prior is omitted", types.StringNull(), types.StringValue(""), false, nil},
		{"unset with unknown prior is omitted", types.StringNull(), types.StringUnknown(), false, nil},
		{"removed after being set is cleared", types.StringNull(), types.StringValue("SN-1"), true, ""},
		{"unknown plan is omitted", types.StringUnknown(), types.StringValue("SN-1"), false, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]any{}
			BodyStringSparse(body, "serial", tc.plan, tc.prior)
			got, ok := body["serial"]
			if ok != tc.present {
				t.Fatalf("present = %v, want %v (body %v)", ok, tc.present, body)
			}
			if ok && got != tc.want {
				t.Errorf("serial = %#v, want %#v", got, tc.want)
			}
		})
	}
}
