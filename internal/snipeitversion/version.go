// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

// Package snipeitversion is the single source of truth for the Snipe-IT
// server versions the provider knows about. Every version decision in the
// codebase goes through this package: named constants for each cutoff we gate
// on plus the support bounds, and a ServerVersion type with comparison
// helpers. Resources never compare version strings inline; they ask
// `srv.AtLeast(snipeitversion.V8_4_0)`.
package snipeitversion

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
)

// Named version cutoffs. Add a constant here (not an inline string literal
// elsewhere) whenever a new behavior boundary is gated.
var (
	// V6_4_0 is the oldest version the provider aims to support cleanly.
	V6_4_0 = version.Must(version.NewVersion("6.4.0"))
	// V8_4_0 is where several response/request shapes changed (group
	// permission values became numeric, maintenances gained required fields).
	V8_4_0 = version.Must(version.NewVersion("8.4.0"))
	// V8_7_0 is where a further round of shape changes landed.
	V8_7_0 = version.Must(version.NewVersion("8.7.0"))
)

// Support bounds. MinSupported is the floor below which the provider warns
// that behavior is untested; MaxTested is the newest version the provider has
// been verified against.
var (
	MinSupported = V6_4_0
	MaxTested    = version.Must(version.NewVersion("8.7.2"))
)

// ServerVersion is a detected Snipe-IT server version. The zero value is
// "unknown": detection failed or the server predates the version endpoint.
type ServerVersion struct {
	v *version.Version
}

// Parse turns a raw version string (as returned by GET /api/v1/version, e.g.
// "v8.0.4" or "v8.7.2 - build ...") into a ServerVersion. A leading "v" and
// any trailing build/hash suffix are tolerated. An unparseable value yields
// the unknown version and a non-nil error.
func Parse(raw string) (ServerVersion, error) {
	// Keep only the leading token: "v8.7.2 - build 24589-..." -> "v8.7.2".
	token := strings.Fields(strings.TrimSpace(raw))
	if len(token) == 0 {
		return ServerVersion{}, fmt.Errorf("empty version string")
	}
	v, err := version.NewVersion(token[0])
	if err != nil {
		return ServerVersion{}, err
	}
	return ServerVersion{v: v}, nil
}

// Known reports whether the version was detected.
func (s ServerVersion) Known() bool { return s.v != nil }

// String returns the detected version, or "unknown".
func (s ServerVersion) String() string {
	if s.v == nil {
		return "unknown"
	}
	return s.v.String()
}

// AtLeast reports whether the server is at least min. An unknown version is
// treated as newest (assume the latest behavior), so version-gated new
// behavior is applied rather than an old code path guessed — reads stay
// tolerant regardless, and writes fail loudly with a real API error if the
// assumption is wrong, which is better than silently sending a legacy shape.
func (s ServerVersion) AtLeast(min *version.Version) bool {
	if s.v == nil {
		return true
	}
	return s.v.Compare(min) >= 0
}

// Below is the negation of AtLeast.
func (s ServerVersion) Below(v *version.Version) bool { return !s.AtLeast(v) }

// IsSupported reports whether the detected version is at least MinSupported.
// An unknown version is considered supported (best effort).
func (s ServerVersion) IsSupported() bool {
	if s.v == nil {
		return true
	}
	return s.v.Compare(MinSupported) >= 0
}
