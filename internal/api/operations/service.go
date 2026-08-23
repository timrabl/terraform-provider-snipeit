// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

// Package operationsapi is the typed API layer for the operations domain
// (asset maintenances, activity report, hardware audit lists). Response
// models live in types.gen.go (generated from api/operations.yaml); this
// file provides the hand-written service methods on top of the shared
// transport.
//
// Request bodies are map[string]any by design: clear-on-unset semantics
// (empty string vs explicit null vs omitted key) are per-field decisions made
// by the provider layer via the tfutil body builders.
package operationsapi

import (
	"context"
	"fmt"
	"net/url"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/snipeitversion"
)

// Service exposes typed access to the operations domain endpoints.
type Service struct {
	c *client.Client
}

// New returns a Service backed by c.
func New(c *client.Client) *Service {
	return &Service{c: c}
}

// ServerVersion exposes the detected Snipe-IT version for version gating.
func (s *Service) ServerVersion() snipeitversion.ServerVersion { return s.c.ServerVersion }

// created is the partial payload of a create envelope.
type created struct {
	ID int64 `json:"id"`
}

// --- maintenances ----------------------------------------------------------

// GetMaintenance fetches one maintenance by id.
func (s *Service) GetMaintenance(ctx context.Context, id int64) (*Maintenance, error) {
	var out Maintenance
	if err := s.c.Get(ctx, fmt.Sprintf("/maintenances/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateMaintenance creates a maintenance and returns its id.
func (s *Service) CreateMaintenance(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/maintenances", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateMaintenance partially updates a maintenance.
func (s *Service) UpdateMaintenance(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/maintenances/%d", id), body, nil)
}

// DeleteMaintenance deletes a maintenance.
func (s *Service) DeleteMaintenance(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/maintenances/%d", id))
}

// --- reports ---------------------------------------------------------------

// ActivityReport reads activity log rows, newest first. query supports
// limit, offset, action_type, item_type and item_id.
func (s *Service) ActivityReport(ctx context.Context, query url.Values) (*ActivityList, error) {
	path := "/reports/activity"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	var out ActivityList
	if err := s.c.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- audit -----------------------------------------------------------------

// AuditDue lists assets due (or nearly due) for audit.
func (s *Service) AuditDue(ctx context.Context) (*AuditList, error) {
	var out AuditList
	if err := s.c.Get(ctx, "/hardware/audit/due", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AuditOverdue lists assets overdue for audit.
func (s *Service) AuditOverdue(ctx context.Context) (*AuditList, error) {
	var out AuditList
	if err := s.c.Get(ctx, "/hardware/audit/overdue", &out); err != nil {
		return nil, err
	}
	return &out, nil
}
