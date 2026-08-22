// Package licensingapi is the typed API layer for the licensing domain
// (licenses and license seats). Response models live in types.gen.go
// (generated from api/licensing.yaml); this file provides the hand-written
// service methods on top of the shared transport.
//
// Request bodies are map[string]any by design: clear-on-unset semantics
// (empty string vs explicit null vs omitted key) are per-field decisions made
// by the provider layer via the tfutil body builders.
package licensingapi

import (
	"context"
	"fmt"
	"net/url"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

// Service exposes typed access to the licensing domain endpoints.
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

// Free reports whether the seat carries no assignment.
func (s *LicenseSeat) Free() bool {
	return s.AssignedUser == nil && s.AssignedAsset == nil
}

// --- licenses ---------------------------------------------------------------

// GetLicense fetches one license by id.
func (s *Service) GetLicense(ctx context.Context, id int64) (*License, error) {
	var out License
	if err := s.c.Get(ctx, fmt.Sprintf("/licenses/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateLicense creates a license and returns its id.
func (s *Service) CreateLicense(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/licenses", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateLicense partially updates a license.
func (s *Service) UpdateLicense(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/licenses/%d", id), body, nil)
}

// DeleteLicense deletes a license.
func (s *Service) DeleteLicense(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/licenses/%d", id))
}

// SearchLicenses lists licenses matching the search term.
func (s *Service) SearchLicenses(ctx context.Context, search string) (*LicenseList, error) {
	var out LicenseList
	if err := s.c.Get(ctx, "/licenses?limit=100&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- seats ------------------------------------------------------------------

// GetSeat fetches one seat of a license.
func (s *Service) GetSeat(ctx context.Context, licenseID, seatID int64) (*LicenseSeat, error) {
	var out LicenseSeat
	if err := s.c.Get(ctx, fmt.Sprintf("/licenses/%d/seats/%d", licenseID, seatID), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSeats fetches one page of a license's seats.
func (s *Service) ListSeats(ctx context.Context, licenseID int64, limit, offset int) (*LicenseSeatList, error) {
	var out LicenseSeatList
	p := fmt.Sprintf("/licenses/%d/seats?limit=%d&offset=%d", licenseID, limit, offset)
	if err := s.c.Get(ctx, p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FreeSeatIDs lists all currently unassigned seats of a license across all
// pages, in the API's listing order.
func (s *Service) FreeSeatIDs(ctx context.Context, licenseID int64) ([]int64, error) {
	var free []int64
	offset := 0
	const pageSize = 100
	for {
		page, err := s.ListSeats(ctx, licenseID, pageSize, offset)
		if err != nil {
			return nil, err
		}
		for i := range page.Rows {
			if page.Rows[i].Free() {
				free = append(free, page.Rows[i].Id)
			}
		}
		offset += len(page.Rows)
		if len(page.Rows) == 0 || int64(offset) >= page.Total {
			break
		}
	}
	return free, nil
}

// AssignSeat PATCHes a seat's assignment. Callers must respect the
// occupied-seat quirk: to switch the target type of an occupied seat,
// ReleaseSeat first, then AssignSeat (a single combined PATCH fails with
// "Target not found").
func (s *Service) AssignSeat(ctx context.Context, licenseID, seatID int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/licenses/%d/seats/%d", licenseID, seatID), body, nil)
}

// ReleaseSeat clears a seat's assignment (both targets explicit JSON null).
func (s *Service) ReleaseSeat(ctx context.Context, licenseID, seatID int64) error {
	body := map[string]any{"assigned_to": nil, "asset_id": nil}
	return s.c.Patch(ctx, fmt.Sprintf("/licenses/%d/seats/%d", licenseID, seatID), body, nil)
}
