// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

// Package tfutil holds shared plumbing for the provider's resource and data
// source implementations: provider-data extraction, import helpers, request
// body builders with Snipe-IT's clear-on-unset semantics, and state mappers.
package tfutil

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

// ClientFromProviderData extracts the configured API client in Configure of
// resources and data sources.
func ClientFromProviderData(providerData any, diags *diag.Diagnostics) *client.Client {
	if providerData == nil {
		return nil // ConfigureProvider not called yet, framework will retry.
	}
	c, ok := providerData.(*client.Client)
	if !ok {
		diags.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *client.Client, got: %T. This is a bug in the provider.", providerData),
		)
		return nil
	}
	return c
}

// ImportNumericID implements ImportState for resources addressed by their
// numeric Snipe-IT id.
func ImportNumericID(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected the numeric Snipe-IT id of the object, got %q.", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// --- request body builders -------------------------------------------------
//
// Snipe-IT accepts partial bodies; omitted keys are left untouched. Strings
// are always sent (empty string clears the field on the server), while
// numeric references and booleans are only sent when they hold a known,
// non-null value.

// BodyString always includes the key; a null/unknown value is sent as "" so
// removing an attribute from configuration clears it server-side.
func BodyString(m map[string]any, key string, v types.String) {
	if v.IsNull() || v.IsUnknown() {
		m[key] = ""
		return
	}
	m[key] = v.ValueString()
}

// BodyNullableInt sends the value, or an explicit JSON null when the config
// value is null, so removing the attribute clears the field server-side.
// Unknown values are omitted.
func BodyNullableInt(m map[string]any, key string, v types.Int64) {
	if v.IsUnknown() {
		return
	}
	if v.IsNull() {
		m[key] = nil
		return
	}
	m[key] = v.ValueInt64()
}

// BodyNullableString sends the value, or an explicit JSON null when the config
// value is null. Use for fields (like dates) where "" fails server validation.
func BodyNullableString(m map[string]any, key string, v types.String) {
	if v.IsUnknown() {
		return
	}
	if v.IsNull() {
		m[key] = nil
		return
	}
	m[key] = v.ValueString()
}

// BodyOptBool includes the key only when the value is known and non-null.
func BodyOptBool(m map[string]any, key string, v types.Bool) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	m[key] = v.ValueBool()
}

// --- state mappers ---------------------------------------------------------

// StateString maps an API string to state, treating "" as null so that
// unconfigured optional attributes stay null.
func StateString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// StateStringPtr maps a nullable API string to state.
func StateStringPtr(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// StateRefID maps a nested object reference to its id, or null when absent.
func StateRefID(r *client.Ref) types.Int64 {
	if r == nil || r.ID == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(r.ID)
}

// StateOptInt maps an integer to state, treating 0 as null (Snipe-IT uses
// 0/null interchangeably for unset numeric fields).
func StateOptInt(n int64) types.Int64 {
	if n == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(n)
}
