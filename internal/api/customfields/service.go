// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

// Package customfieldsapi is the typed API layer for the customfields domain
// (custom fields, fieldsets, and field/fieldset associations). Response
// models live in types.gen.go (generated from api/customfields.yaml); this
// file provides the hand-written service methods on top of the shared
// transport.
//
// Request bodies are map[string]any by design: clear-on-unset semantics
// (empty string vs explicit null vs omitted key) are per-field decisions made
// by the provider layer via the tfutil body builders.
package customfieldsapi

import (
	"context"
	"fmt"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

// Service exposes typed access to the customfields domain endpoints.
type Service struct {
	c *client.Client
}

// New returns a Service backed by c.
func New(c *client.Client) *Service {
	return &Service{c: c}
}

// created is the partial payload of a create envelope.
type created struct {
	ID int64 `json:"id"`
}

// --- fields ----------------------------------------------------------------

// GetField fetches one custom field by id.
func (s *Service) GetField(ctx context.Context, id int64) (*Field, error) {
	var out Field
	if err := s.c.Get(ctx, fmt.Sprintf("/fields/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateField creates a custom field and returns its id.
func (s *Service) CreateField(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/fields", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateField partially updates a custom field.
func (s *Service) UpdateField(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/fields/%d", id), body, nil)
}

// DeleteField deletes a custom field.
func (s *Service) DeleteField(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/fields/%d", id))
}

// ListFields lists all custom fields. The endpoint has no ?search= support;
// callers filter client-side.
func (s *Service) ListFields(ctx context.Context) (*FieldList, error) {
	var out FieldList
	if err := s.c.Get(ctx, "/fields", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AssociateField adds the field to a fieldset.
func (s *Service) AssociateField(ctx context.Context, fieldID, fieldsetID int64) error {
	body := map[string]any{"fieldset_id": fieldsetID}
	return s.c.Post(ctx, fmt.Sprintf("/fields/%d/associate", fieldID), body, nil)
}

// DisassociateField removes the field from a fieldset.
func (s *Service) DisassociateField(ctx context.Context, fieldID, fieldsetID int64) error {
	body := map[string]any{"fieldset_id": fieldsetID}
	return s.c.Post(ctx, fmt.Sprintf("/fields/%d/disassociate", fieldID), body, nil)
}

// --- fieldsets -------------------------------------------------------------

// GetFieldset fetches one fieldset by id, member fields embedded.
func (s *Service) GetFieldset(ctx context.Context, id int64) (*Fieldset, error) {
	var out Fieldset
	if err := s.c.Get(ctx, fmt.Sprintf("/fieldsets/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateFieldset creates a fieldset and returns its id.
func (s *Service) CreateFieldset(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/fieldsets", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateFieldset partially updates a fieldset.
func (s *Service) UpdateFieldset(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/fieldsets/%d", id), body, nil)
}

// DeleteFieldset deletes a fieldset.
func (s *Service) DeleteFieldset(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/fieldsets/%d", id))
}

// ListFieldsets lists all fieldsets. The endpoint has no ?search= support;
// callers filter client-side.
func (s *Service) ListFieldsets(ctx context.Context) (*FieldsetList, error) {
	var out FieldsetList
	if err := s.c.Get(ctx, "/fieldsets", &out); err != nil {
		return nil, err
	}
	return &out, nil
}
