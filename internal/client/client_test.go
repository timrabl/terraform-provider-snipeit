// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// newTestClient returns a Client pointed at the given test server with retry
// delays eliminated.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(Config{URL: srv.URL, Token: "test-token"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c.retryDelay = func(int) time.Duration { return 0 }
	return c
}

func TestGetRawObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		if r.URL.Path != "/api/v1/manufacturers/7" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id": 7, "name": "Apple"}`))
	}))
	defer srv.Close()

	var out struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	if err := newTestClient(t, srv).Get(context.Background(), "/manufacturers/7", &out); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if out.ID != 7 || out.Name != "Apple" {
		t.Errorf("decoded %+v", out)
	}
}

func TestGetNotFoundEnvelopeOn200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Snipe-IT quirk: missing objects come back as HTTP 200 error envelopes.
		_, _ = w.Write([]byte(`{"status":"error","messages":"Manufacturer not found","payload":null}`))
	}))
	defer srv.Close()

	err := newTestClient(t, srv).Get(context.Background(), "/manufacturers/999", &struct{}{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestPostSuccessEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "probe" {
			t.Errorf("body = %v", body)
		}
		_, _ = w.Write([]byte(`{"status":"success","messages":"Created.","payload":{"id":42,"name":"probe"}}`))
	}))
	defer srv.Close()

	var out struct {
		ID int64 `json:"id"`
	}
	err := newTestClient(t, srv).Post(context.Background(), "/manufacturers", map[string]any{"name": "probe"}, &out)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if out.ID != 42 {
		t.Errorf("payload id = %d", out.ID)
	}
}

func TestPostValidationErrorEnvelopeOn200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"error","messages":{"name":["The name field is required."],"qty":["Must be positive."]},"payload":null}`))
	}))
	defer srv.Close()

	err := newTestClient(t, srv).Post(context.Background(), "/manufacturers", map[string]any{}, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *APIError, got %v", err)
	}
	want := "name: The name field is required., qty: Must be positive."
	if apiErr.Messages != want {
		t.Errorf("flattened messages = %q, want %q", apiErr.Messages, want)
	}
}

func TestDeleteNotFoundEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"error","messages":"Manufacturer does not exist","payload":null}`))
	}))
	defer srv.Close()

	err := newTestClient(t, srv).Delete(context.Background(), "/manufacturers/1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestRetryOn429(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`Too many requests`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"success","messages":"ok","payload":{"id":1}}`))
	}))
	defer srv.Close()

	err := newTestClient(t, srv).Post(context.Background(), "/things", map[string]any{"a": 1}, nil)
	if err != nil {
		t.Fatalf("Post after retries: %v", err)
	}
	if got := calls.Load(); got != 3 {
		t.Errorf("server saw %d calls, want 3", got)
	}
}

func TestRetryExhaustionSurfaces429(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`Too many requests`))
	}))
	defer srv.Close()

	err := newTestClient(t, srv).Post(context.Background(), "/things", map[string]any{"a": 1}, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", apiErr.StatusCode)
	}
}

func TestFlattenMessages(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"string", `"Simple message"`, "Simple message"},
		{"map sorted", `{"b":["two"],"a":["one","uno"]}`, "a: one; uno, b: two"},
		{"fallback", `[1,2]`, "[1,2]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := flattenMessages(json.RawMessage(tc.raw)); got != tc.want {
				t.Errorf("flattenMessages(%s) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestGetRedirectMeansNotFound(t *testing.T) {
	// Snipe-IT redirects API GETs for some soft-deleted objects to the web
	// login page; the client must not follow it and must report not-found.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	}))
	defer srv.Close()

	err := newTestClient(t, srv).Get(context.Background(), "/accessories/99", nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for redirected GET, got %v", err)
	}
}
