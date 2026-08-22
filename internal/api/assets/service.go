// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

// Package assetsapi is the typed API layer for the assets domain
// (manufacturers, categories, status labels, asset models, hardware).
// Response models live in types.gen.go (generated from api/assets.yaml); this
// file provides the hand-written service methods on top of the shared
// transport.
//
// Request bodies are map[string]any by design: clear-on-unset semantics
// (empty string vs explicit null vs omitted key) are per-field decisions made
// by the provider layer via the tfutil body builders.
package assetsapi

import (
	"context"
	"fmt"
	"net/url"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

// Service exposes typed access to the assets domain endpoints.
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

// --- manufacturers ----------------------------------------------------------

// GetManufacturer fetches one manufacturer by id.
func (s *Service) GetManufacturer(ctx context.Context, id int64) (*Manufacturer, error) {
	var out Manufacturer
	if err := s.c.Get(ctx, fmt.Sprintf("/manufacturers/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateManufacturer creates a manufacturer and returns its id.
func (s *Service) CreateManufacturer(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/manufacturers", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateManufacturer partially updates a manufacturer.
func (s *Service) UpdateManufacturer(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/manufacturers/%d", id), body, nil)
}

// DeleteManufacturer deletes a manufacturer.
func (s *Service) DeleteManufacturer(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/manufacturers/%d", id))
}

// SearchManufacturers lists manufacturers matching the search term.
func (s *Service) SearchManufacturers(ctx context.Context, search string) (*ManufacturerList, error) {
	var out ManufacturerList
	if err := s.c.Get(ctx, "/manufacturers?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- categories -------------------------------------------------------------

// GetCategory fetches one category by id.
func (s *Service) GetCategory(ctx context.Context, id int64) (*Category, error) {
	var out Category
	if err := s.c.Get(ctx, fmt.Sprintf("/categories/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateCategory creates a category and returns its id.
func (s *Service) CreateCategory(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/categories", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateCategory partially updates a category.
func (s *Service) UpdateCategory(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/categories/%d", id), body, nil)
}

// DeleteCategory deletes a category.
func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/categories/%d", id))
}

// SearchCategories lists categories matching the search term.
func (s *Service) SearchCategories(ctx context.Context, search string) (*CategoryList, error) {
	var out CategoryList
	if err := s.c.Get(ctx, "/categories?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- status labels ----------------------------------------------------------

// GetStatusLabel fetches one status label by id.
func (s *Service) GetStatusLabel(ctx context.Context, id int64) (*StatusLabel, error) {
	var out StatusLabel
	if err := s.c.Get(ctx, fmt.Sprintf("/statuslabels/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateStatusLabel creates a status label and returns its id.
func (s *Service) CreateStatusLabel(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/statuslabels", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateStatusLabel partially updates a status label.
func (s *Service) UpdateStatusLabel(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/statuslabels/%d", id), body, nil)
}

// DeleteStatusLabel deletes a status label.
func (s *Service) DeleteStatusLabel(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/statuslabels/%d", id))
}

// SearchStatusLabels lists status labels matching the search term.
func (s *Service) SearchStatusLabels(ctx context.Context, search string) (*StatusLabelList, error) {
	var out StatusLabelList
	if err := s.c.Get(ctx, "/statuslabels?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- models -----------------------------------------------------------------

// GetModel fetches one asset model by id.
func (s *Service) GetModel(ctx context.Context, id int64) (*Model, error) {
	var out Model
	if err := s.c.Get(ctx, fmt.Sprintf("/models/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateModel creates an asset model and returns its id.
func (s *Service) CreateModel(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/models", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateModel partially updates an asset model.
func (s *Service) UpdateModel(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/models/%d", id), body, nil)
}

// DeleteModel deletes an asset model.
func (s *Service) DeleteModel(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/models/%d", id))
}

// SearchModels lists asset models matching the search term.
func (s *Service) SearchModels(ctx context.Context, search string) (*ModelList, error) {
	var out ModelList
	if err := s.c.Get(ctx, "/models?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- hardware ---------------------------------------------------------------

// GetHardware fetches one asset by id.
func (s *Service) GetHardware(ctx context.Context, id int64) (*Hardware, error) {
	var out Hardware
	if err := s.c.Get(ctx, fmt.Sprintf("/hardware/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHardwareByTag fetches one asset by its asset tag.
func (s *Service) GetHardwareByTag(ctx context.Context, tag string) (*Hardware, error) {
	var out Hardware
	if err := s.c.Get(ctx, "/hardware/bytag/"+url.PathEscape(tag), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FindHardwareBySerial finds assets by serial number. Unlike /bytag/, the
// endpoint returns a list shape.
func (s *Service) FindHardwareBySerial(ctx context.Context, serial string) (*HardwareList, error) {
	var out HardwareList
	if err := s.c.Get(ctx, "/hardware/byserial/"+url.PathEscape(serial), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateHardware creates an asset and returns its id.
func (s *Service) CreateHardware(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/hardware", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateHardware partially updates an asset.
func (s *Service) UpdateHardware(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/hardware/%d", id), body, nil)
}

// DeleteHardware deletes an asset.
func (s *Service) DeleteHardware(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/hardware/%d", id))
}
