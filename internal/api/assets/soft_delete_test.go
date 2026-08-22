// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package assetsapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

// GET /hardware/{id} keeps returning soft-deleted assets with deleted_at
// set; the service must report them as gone.
func TestGetHardwareSoftDeletedIsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": 80, "asset_tag": "bt-HW-020",
			"warranty_months": null, "requestable": false,
			"deleted_at": {"datetime": "2026-08-22 17:51:53", "formatted": "..."}
		}`))
	}))
	defer srv.Close()

	c, err := client.New(client.Config{URL: srv.URL, Token: "t"})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := New(c).GetHardware(context.Background(), 80); !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for soft-deleted asset, got %v", err)
	}
	if _, err := New(c).GetHardwareByTag(context.Background(), "bt-HW-020"); !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for soft-deleted asset by tag, got %v", err)
	}
}
