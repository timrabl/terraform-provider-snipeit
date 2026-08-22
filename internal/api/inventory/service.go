// Package inventoryapi is the typed API layer for the inventory domain
// (accessories, consumables, components) and the checkout/checkin flows,
// including the asset checkout flow. Response models live in types.gen.go
// (generated from api/inventory.yaml); this file provides the hand-written
// service methods on top of the shared transport.
//
// Request bodies are map[string]any by design: clear-on-unset semantics
// (empty string vs explicit null vs omitted key) are per-field decisions made
// by the provider layer via the tfutil body builders.
package inventoryapi

import (
	"context"
	"fmt"
	"net/url"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

// Service exposes typed access to the inventory domain endpoints.
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

// --- accessories ------------------------------------------------------------

// GetAccessory fetches one accessory by id.
func (s *Service) GetAccessory(ctx context.Context, id int64) (*Accessory, error) {
	var out Accessory
	if err := s.c.Get(ctx, fmt.Sprintf("/accessories/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateAccessory creates an accessory and returns its id.
func (s *Service) CreateAccessory(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/accessories", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateAccessory partially updates an accessory.
func (s *Service) UpdateAccessory(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/accessories/%d", id), body, nil)
}

// DeleteAccessory deletes an accessory.
func (s *Service) DeleteAccessory(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/accessories/%d", id))
}

// SearchAccessories lists accessories matching the search term.
func (s *Service) SearchAccessories(ctx context.Context, search string) (*AccessoryList, error) {
	var out AccessoryList
	if err := s.c.Get(ctx, "/accessories?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAccessoryCheckouts lists the checkout (pivot) rows of an accessory.
func (s *Service) ListAccessoryCheckouts(ctx context.Context, accessoryID int64) (*AccessoryCheckoutList, error) {
	var out AccessoryCheckoutList
	if err := s.c.Get(ctx, fmt.Sprintf("/accessories/%d/checkedout?limit=500", accessoryID), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CheckoutAccessory checks one accessory unit out (users only on v8.0.4).
// The response payload is null; identify the new pivot row by diffing
// ListAccessoryCheckouts before and after.
func (s *Service) CheckoutAccessory(ctx context.Context, accessoryID int64, body map[string]any) error {
	return s.c.Post(ctx, fmt.Sprintf("/accessories/%d/checkout", accessoryID), body, nil)
}

// CheckinAccessory checks an accessory unit back in. pivotID is the checkout
// row id from ListAccessoryCheckouts, NOT the accessory id.
func (s *Service) CheckinAccessory(ctx context.Context, pivotID int64) error {
	return s.c.Post(ctx, fmt.Sprintf("/accessories/%d/checkin", pivotID), map[string]any{}, nil)
}

// --- consumables ------------------------------------------------------------

// GetConsumable fetches one consumable by id.
func (s *Service) GetConsumable(ctx context.Context, id int64) (*Consumable, error) {
	var out Consumable
	if err := s.c.Get(ctx, fmt.Sprintf("/consumables/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateConsumable creates a consumable and returns its id.
func (s *Service) CreateConsumable(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/consumables", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateConsumable partially updates a consumable.
func (s *Service) UpdateConsumable(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/consumables/%d", id), body, nil)
}

// DeleteConsumable deletes a consumable.
func (s *Service) DeleteConsumable(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/consumables/%d", id))
}

// SearchConsumables lists consumables matching the search term.
func (s *Service) SearchConsumables(ctx context.Context, search string) (*ConsumableList, error) {
	var out ConsumableList
	if err := s.c.Get(ctx, "/consumables?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CheckoutConsumable consumes one unit of a consumable. IRREVERSIBLE: the API
// has no consumable checkin.
func (s *Service) CheckoutConsumable(ctx context.Context, consumableID int64, body map[string]any) error {
	return s.c.Post(ctx, fmt.Sprintf("/consumables/%d/checkout", consumableID), body, nil)
}

// --- components -------------------------------------------------------------

// GetComponent fetches one component by id.
func (s *Service) GetComponent(ctx context.Context, id int64) (*Component, error) {
	var out Component
	if err := s.c.Get(ctx, fmt.Sprintf("/components/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateComponent creates a component and returns its id.
func (s *Service) CreateComponent(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/components", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateComponent partially updates a component.
func (s *Service) UpdateComponent(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/components/%d", id), body, nil)
}

// DeleteComponent deletes a component.
func (s *Service) DeleteComponent(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/components/%d", id))
}

// SearchComponents lists components matching the search term.
func (s *Service) SearchComponents(ctx context.Context, search string) (*ComponentList, error) {
	var out ComponentList
	if err := s.c.Get(ctx, "/components?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListComponentAssets lists the assets holding units of a component; rows
// carry the assigned_pivot_id needed for checkin.
func (s *Service) ListComponentAssets(ctx context.Context, componentID int64) (*ComponentAssetList, error) {
	var out ComponentAssetList
	if err := s.c.Get(ctx, fmt.Sprintf("/components/%d/assets?limit=500", componentID), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CheckoutComponent checks component units out to an asset.
func (s *Service) CheckoutComponent(ctx context.Context, componentID int64, body map[string]any) error {
	return s.c.Post(ctx, fmt.Sprintf("/components/%d/checkout", componentID), body, nil)
}

// CheckinComponent checks component units back in. pivotID is the
// assigned_pivot_id from ListComponentAssets, NOT the component id.
func (s *Service) CheckinComponent(ctx context.Context, pivotID int64, body map[string]any) error {
	return s.c.Post(ctx, fmt.Sprintf("/components/%d/checkin", pivotID), body, nil)
}

// --- hardware checkout ------------------------------------------------------

// GetHardwareAssignment reads the assignment subset of an asset detail.
func (s *Service) GetHardwareAssignment(ctx context.Context, assetID int64) (*HardwareAssignment, error) {
	var out HardwareAssignment
	if err := s.c.Get(ctx, fmt.Sprintf("/hardware/%d", assetID), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CheckoutHardware checks an asset out to a user, asset or location.
func (s *Service) CheckoutHardware(ctx context.Context, assetID int64, body map[string]any) error {
	return s.c.Post(ctx, fmt.Sprintf("/hardware/%d/checkout", assetID), body, nil)
}

// CheckinHardware checks an asset back in (addresses the asset id).
func (s *Service) CheckinHardware(ctx context.Context, assetID int64) error {
	return s.c.Post(ctx, fmt.Sprintf("/hardware/%d/checkin", assetID), map[string]any{}, nil)
}
