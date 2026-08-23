// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Ref is a nested related-object reference as returned by GET endpoints,
// e.g. "manufacturer": {"id": 3, "name": "Apple"}.
type Ref struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// IDOrZero returns the referenced id, or 0 when the reference is absent.
func (r *Ref) IDOrZero() int64 {
	if r == nil {
		return 0
	}
	return r.ID
}

// DateTime is the nested timestamp shape {"datetime": "...", "formatted": "..."}
// used for created_at/updated_at and similar fields.
type DateTime struct {
	DateTime  string `json:"datetime"`
	Formatted string `json:"formatted"`
}

// Date is the nested date shape {"date": "2020-01-01", "formatted": "..."}
// used for purchase_date and similar fields.
type Date struct {
	Date      string `json:"date"`
	Formatted string `json:"formatted"`
}

// FlexInt unmarshals integers that the API sometimes returns as bare numbers,
// sometimes as strings, and sometimes as decorated strings such as "24 mo." or
// "36 months". A null or empty value decodes to 0.
type FlexInt int64

func (f *FlexInt) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "null" || s == `""` {
		*f = 0
		return nil
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		*f = FlexInt(n)
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	// Take the leading integer of decorated values like "24 mo.".
	str = strings.TrimSpace(str)
	end := 0
	for end < len(str) && (str[end] >= '0' && str[end] <= '9' || (end == 0 && str[end] == '-')) {
		end++
	}
	if end == 0 {
		*f = 0
		return nil
	}
	n, err := strconv.ParseInt(str[:end], 10, 64)
	if err != nil {
		return err
	}
	*f = FlexInt(n)
	return nil
}

// FlexBool unmarshals booleans that the API sometimes returns as true/false,
// sometimes as 0/1, and sometimes as "0"/"1".
type FlexBool bool

func (f *FlexBool) UnmarshalJSON(data []byte) error {
	s := strings.Trim(strings.TrimSpace(string(data)), `"`)
	switch s {
	case "true", "1":
		*f = true
	case "false", "0", "null", "":
		*f = false
	default:
		var b bool
		if err := json.Unmarshal(data, &b); err != nil {
			return err
		}
		*f = FlexBool(b)
	}
	return nil
}

// NormalizeMoney strips thousands separators from formatted money strings such
// as "1,234.56" so they round-trip stably in state.
func NormalizeMoney(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), ",", "")
}

// FlexString unmarshals a JSON value that may arrive as a string or a bare
// number and always yields its string form. Snipe-IT changed several fields
// from string-encoded to numeric between 8.0 and 8.4 (e.g. group permission
// values went from "0"/"1" to 0/1), so decoding them as a plain string fails
// with "cannot unmarshal number into ...". FlexString tolerates both.
type FlexString string

func (f *FlexString) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "null" {
		*f = ""
		return nil
	}
	// Already a JSON string.
	if len(s) > 0 && s[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		*f = FlexString(str)
		return nil
	}
	// Bare number (or bool) -> use its literal text.
	*f = FlexString(s)
	return nil
}
