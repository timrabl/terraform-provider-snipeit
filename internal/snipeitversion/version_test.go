// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package snipeitversion

import "testing"

func TestParse(t *testing.T) {
	cases := map[string]struct {
		want string
		ok   bool
	}{
		"v8.0.4":                    {"8.0.4", true},
		"8.7.2":                     {"8.7.2", true},
		"v8.7.2 - build 24589-gf0b": {"8.7.2", true}, // trailing build suffix
		"v6.4.2":                    {"6.4.2", true},
		"garbage":                   {"", false},
	}
	for in, tc := range cases {
		got, err := Parse(in)
		if tc.ok {
			if err != nil {
				t.Errorf("Parse(%q) unexpected err %v", in, err)
				continue
			}
			if got.String() != tc.want {
				t.Errorf("Parse(%q) = %q, want %q", in, got.String(), tc.want)
			}
			if !got.Known() {
				t.Errorf("Parse(%q) should be Known", in)
			}
		} else if err == nil {
			t.Errorf("Parse(%q) expected error", in)
		}
	}
}

func TestAtLeastAndSupport(t *testing.T) {
	v80, _ := Parse("8.0.4")
	if v80.AtLeast(V8_4_0) {
		t.Error("8.0.4 should be below 8.4.0")
	}
	if !v80.AtLeast(V6_4_0) {
		t.Error("8.0.4 should be at least 6.4.0")
	}
	if !v80.IsSupported() {
		t.Error("8.0.4 should be supported")
	}

	v87, _ := Parse("8.7.2")
	if !v87.AtLeast(V8_4_0) || !v87.AtLeast(V8_7_0) {
		t.Error("8.7.2 should be at least 8.4.0 and 8.7.0")
	}

	v60, _ := Parse("6.0.14")
	if v60.IsSupported() {
		t.Error("6.0.14 is below MinSupported (6.4.0)")
	}
	if !v60.Below(V8_4_0) {
		t.Error("6.0.14 should be below 8.4.0")
	}

	// Unknown version: assume newest, best-effort supported.
	var unknown ServerVersion
	if unknown.Known() {
		t.Error("zero value should be unknown")
	}
	if !unknown.AtLeast(V8_7_0) || !unknown.IsSupported() {
		t.Error("unknown should assume newest and be best-effort supported")
	}
	if unknown.String() != "unknown" {
		t.Errorf("unknown.String() = %q", unknown.String())
	}
}
