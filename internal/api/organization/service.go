// Package organizationapi is the typed API layer for the organization domain
// (companies, departments, locations, suppliers). Response models live in
// types.gen.go (generated from api/organization.yaml); this file provides the
// hand-written service methods on top of the shared transport.
//
// Request bodies are map[string]any by design: clear-on-unset semantics
// (empty string vs explicit null vs omitted key) are per-field decisions made
// by the provider layer via the tfutil body builders.
package organizationapi

import (
	"context"
	"fmt"
	"net/url"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

// Service exposes typed access to the organization domain endpoints.
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

// --- companies -------------------------------------------------------------

// GetCompany fetches one company by id.
func (s *Service) GetCompany(ctx context.Context, id int64) (*Company, error) {
	var out Company
	if err := s.c.Get(ctx, fmt.Sprintf("/companies/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateCompany creates a company and returns its id.
func (s *Service) CreateCompany(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/companies", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateCompany partially updates a company.
func (s *Service) UpdateCompany(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/companies/%d", id), body, nil)
}

// DeleteCompany deletes a company.
func (s *Service) DeleteCompany(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/companies/%d", id))
}

// SearchCompanies lists companies matching the search term.
func (s *Service) SearchCompanies(ctx context.Context, search string) (*CompanyList, error) {
	var out CompanyList
	if err := s.c.Get(ctx, "/companies?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- departments -----------------------------------------------------------

// GetDepartment fetches one department by id.
func (s *Service) GetDepartment(ctx context.Context, id int64) (*Department, error) {
	var out Department
	if err := s.c.Get(ctx, fmt.Sprintf("/departments/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateDepartment creates a department and returns its id.
func (s *Service) CreateDepartment(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/departments", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateDepartment partially updates a department.
func (s *Service) UpdateDepartment(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/departments/%d", id), body, nil)
}

// DeleteDepartment deletes a department.
func (s *Service) DeleteDepartment(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/departments/%d", id))
}

// SearchDepartments lists departments matching the search term.
func (s *Service) SearchDepartments(ctx context.Context, search string) (*DepartmentList, error) {
	var out DepartmentList
	if err := s.c.Get(ctx, "/departments?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- locations -------------------------------------------------------------

// GetLocation fetches one location by id.
func (s *Service) GetLocation(ctx context.Context, id int64) (*Location, error) {
	var out Location
	if err := s.c.Get(ctx, fmt.Sprintf("/locations/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateLocation creates a location and returns its id.
func (s *Service) CreateLocation(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/locations", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateLocation partially updates a location.
func (s *Service) UpdateLocation(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/locations/%d", id), body, nil)
}

// DeleteLocation deletes a location.
func (s *Service) DeleteLocation(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/locations/%d", id))
}

// SearchLocations lists locations matching the search term.
func (s *Service) SearchLocations(ctx context.Context, search string) (*LocationList, error) {
	var out LocationList
	if err := s.c.Get(ctx, "/locations?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- suppliers -------------------------------------------------------------

// GetSupplier fetches one supplier by id.
func (s *Service) GetSupplier(ctx context.Context, id int64) (*Supplier, error) {
	var out Supplier
	if err := s.c.Get(ctx, fmt.Sprintf("/suppliers/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateSupplier creates a supplier and returns its id.
func (s *Service) CreateSupplier(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/suppliers", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateSupplier partially updates a supplier.
func (s *Service) UpdateSupplier(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/suppliers/%d", id), body, nil)
}

// DeleteSupplier deletes a supplier.
func (s *Service) DeleteSupplier(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/suppliers/%d", id))
}

// SearchSuppliers lists suppliers matching the search term.
func (s *Service) SearchSuppliers(ctx context.Context, search string) (*SupplierList, error) {
	var out SupplierList
	if err := s.c.Get(ctx, "/suppliers?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
