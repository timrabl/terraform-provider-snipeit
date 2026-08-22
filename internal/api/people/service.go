// Package peopleapi is the typed API layer for the people domain (users and
// permission groups). Response models live in types.gen.go (generated from
// api/people.yaml); this file provides the hand-written service methods on
// top of the shared transport.
//
// Request bodies are map[string]any by design: clear-on-unset semantics
// (empty string vs explicit null vs omitted key) are per-field decisions made
// by the provider layer via the tfutil body builders.
package peopleapi

import (
	"context"
	"fmt"
	"net/url"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

// Service exposes typed access to the people domain endpoints.
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

// --- users -------------------------------------------------------------------

// GetUser fetches one user by id. The password is never part of the response.
func (s *Service) GetUser(ctx context.Context, id int64) (*User, error) {
	var out User
	if err := s.c.Get(ctx, fmt.Sprintf("/users/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Me fetches the user owning the configured API token.
func (s *Service) Me(ctx context.Context) (*User, error) {
	var out User
	if err := s.c.Get(ctx, "/users/me", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateUser creates a user and returns its id.
func (s *Service) CreateUser(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/users", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateUser partially updates a user.
func (s *Service) UpdateUser(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/users/%d", id), body, nil)
}

// DeleteUser deletes a user. Users with assigned assets, accessories or
// licenses cannot be deleted.
func (s *Service) DeleteUser(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/users/%d", id))
}

// SearchUsers lists users matching the search term.
func (s *Service) SearchUsers(ctx context.Context, search string) (*UserList, error) {
	var out UserList
	if err := s.c.Get(ctx, "/users?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- groups ------------------------------------------------------------------

// GetGroup fetches one group by id.
func (s *Service) GetGroup(ctx context.Context, id int64) (*Group, error) {
	var out Group
	if err := s.c.Get(ctx, fmt.Sprintf("/groups/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateGroup creates a group and returns its id.
func (s *Service) CreateGroup(ctx context.Context, body map[string]any) (int64, error) {
	var out created
	if err := s.c.Post(ctx, "/groups", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// UpdateGroup partially updates a group.
//
// v8.0.x quirk: when body includes a permissions map the server answers HTTP
// 500 even though the update IS applied, and name is required even for
// permission-only changes. Callers must verify server state on error before
// surfacing it (see the group resource).
func (s *Service) UpdateGroup(ctx context.Context, id int64, body map[string]any) error {
	return s.c.Patch(ctx, fmt.Sprintf("/groups/%d", id), body, nil)
}

// DeleteGroup deletes a group.
func (s *Service) DeleteGroup(ctx context.Context, id int64) error {
	return s.c.Delete(ctx, fmt.Sprintf("/groups/%d", id))
}

// SearchGroups lists groups matching the search term.
func (s *Service) SearchGroups(ctx context.Context, search string) (*GroupList, error) {
	var out GroupList
	if err := s.c.Get(ctx, "/groups?limit=500&search="+url.QueryEscape(search), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
