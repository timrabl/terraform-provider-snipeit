// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package tfutil

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

// Snipe-IT serializes money fields locale-formatted ("1,234.50", two padded
// decimals, comma thousands separators) while accepting only plain decimals
// on write (comma-formatted input makes the API answer with a redirect).
// MoneyType/MoneyValue make the attribute stable anyway: input is validated
// to a dot-decimal, requests always send the normalized plain form, reads
// strip the separators, and semantic equality treats "1234.5", "1234.50"
// and "1,234.50" as the same value.

var moneyPattern = regexp.MustCompile(`^\d{1,3}(,\d{3})*(\.\d+)?$|^\d+(\.\d+)?$`)

// CanonicalMoney reduces a money string to its canonical decimal form:
// thousands separators stripped, trailing fractional zeros and a trailing
// decimal point removed. "1,234.50" -> "1234.5", "7.00" -> "7".
func CanonicalMoney(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	if s == "" {
		return "0"
	}
	return s
}

// --- type ------------------------------------------------------------------

// MoneyType is the attribute type for money values.
type MoneyType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = MoneyType{}

func (t MoneyType) Equal(o attr.Type) bool {
	other, ok := o.(MoneyType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t MoneyType) String() string {
	return "tfutil.MoneyType"
}

func (t MoneyType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return MoneyValue{StringValue: in}, nil
}

func (t MoneyType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T for MoneyType", attrValue)
	}
	value, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to MoneyValue: %v", diags)
	}
	return value, nil
}

func (t MoneyType) ValueType(ctx context.Context) attr.Value {
	return MoneyValue{}
}

// --- value -----------------------------------------------------------------

// MoneyValue is a string value with money semantics.
type MoneyValue struct {
	basetypes.StringValue
}

var (
	_ basetypes.StringValuableWithSemanticEquals = MoneyValue{}
	_ xattr.ValidateableAttribute                = MoneyValue{}
)

func (v MoneyValue) Equal(o attr.Value) bool {
	other, ok := o.(MoneyValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v MoneyValue) Type(ctx context.Context) attr.Type {
	return MoneyType{}
}

// StringSemanticEquals reports two money values as equal when they denote the
// same amount, regardless of separator or decimal-padding differences.
func (v MoneyValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	newValue, ok := newValuable.(MoneyValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("Expected tfutil.MoneyValue, got: %T. This is a bug in the provider.", newValuable),
		)
		return false, diags
	}
	return CanonicalMoney(v.ValueString()) == CanonicalMoney(newValue.ValueString()), diags
}

// ValidateAttribute rejects values the API cannot digest, most importantly
// the European decimal-comma form, which is indistinguishable from a
// thousands separator.
func (v MoneyValue) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	if !moneyPattern.MatchString(strings.TrimSpace(v.ValueString())) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Money Value",
			fmt.Sprintf("%q is not a valid amount. Use a dot as the decimal separator, "+
				"e.g. \"1234.50\" (thousands commas like \"1,234.50\" are accepted).", v.ValueString()),
		)
	}
}

// NewMoneyValue returns a known MoneyValue.
func NewMoneyValue(s string) MoneyValue {
	return MoneyValue{StringValue: basetypes.NewStringValue(s)}
}

// NewMoneyNull returns a null MoneyValue.
func NewMoneyNull() MoneyValue {
	return MoneyValue{StringValue: basetypes.NewStringNull()}
}

// StateMoneyPtr maps a nullable API money string into state, normalized
// (separators stripped); null and "" become null.
func StateMoneyPtr(s *string) MoneyValue {
	if s == nil || strings.TrimSpace(*s) == "" {
		return NewMoneyNull()
	}
	return NewMoneyValue(client.NormalizeMoney(*s))
}

// BodyMoney sends the normalized plain-decimal amount, or an explicit JSON
// null when the config value is null so removing the attribute clears the
// field server-side (the API accepts null and rejects comma-formatted input).
func BodyMoney(m map[string]any, key string, v MoneyValue) {
	if v.IsUnknown() {
		return
	}
	if v.IsNull() {
		m[key] = nil
		return
	}
	m[key] = client.NormalizeMoney(v.ValueString())
}
