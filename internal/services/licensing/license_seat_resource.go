// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package licensing

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	licensingapi "github.com/timrabl/terraform-provider-snipeit/internal/api/licensing"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &LicenseSeatResource{}
	_ resource.ResourceWithImportState = &LicenseSeatResource{}
)

// NewLicenseSeatResource returns a new snipeit_license_seat resource.
func NewLicenseSeatResource() resource.Resource {
	return &LicenseSeatResource{}
}

// LicenseSeatResource manages the assignment of one seat of an existing
// license. Seats themselves are created and destroyed by Snipe-IT together
// with the license's `seats` count; this resource claims a free seat on
// create, reassigns it on update, and releases it on delete.
type LicenseSeatResource struct {
	svc *licensingapi.Service
}

// LicenseSeatResourceModel describes the resource data model.
type LicenseSeatResourceModel struct {
	ID                types.Int64 `tfsdk:"id"`
	LicenseID         types.Int64 `tfsdk:"license_id"`
	SeatID            types.Int64 `tfsdk:"seat_id"`
	AssignedToUserID  types.Int64 `tfsdk:"assigned_to_user_id"`
	AssignedToAssetID types.Int64 `tfsdk:"assigned_to_asset_id"`
}

func (r *LicenseSeatResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license_seat"
}

func (r *LicenseSeatResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Assigns one seat of a `snipeit_license` to a user or an asset. " +
			"On create a free seat is claimed; on destroy the seat is checked in again. " +
			"Exactly one of `assigned_to_user_id` and `assigned_to_asset_id` must be set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the claimed seat (same as `seat_id`).",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"license_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the license whose seat is assigned. Changing this forces a new assignment.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"seat_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the claimed seat, chosen automatically from the license's free seats.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"assigned_to_user_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the user the seat is assigned to.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(
						path.MatchRoot("assigned_to_user_id"),
						path.MatchRoot("assigned_to_asset_id"),
					),
				},
			},
			"assigned_to_asset_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the asset the seat is assigned to.",
				Optional:            true,
			},
		},
	}
}

func (r *LicenseSeatResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = licensingapi.New(c)
	}
}

// assignmentBody builds the PATCH body. Both keys are always present so that
// reassigning between user and asset clears the other side (the API clears on
// explicit JSON null).
func (m *LicenseSeatResourceModel) assignmentBody() map[string]any {
	body := map[string]any{
		"assigned_to": nil,
		"asset_id":    nil,
	}
	if !m.AssignedToUserID.IsNull() && !m.AssignedToUserID.IsUnknown() {
		body["assigned_to"] = m.AssignedToUserID.ValueInt64()
	}
	if !m.AssignedToAssetID.IsNull() && !m.AssignedToAssetID.IsUnknown() {
		body["asset_id"] = m.AssignedToAssetID.ValueInt64()
	}
	return body
}

func (m *LicenseSeatResourceModel) matchesAssignment(api *licensingapi.LicenseSeat) bool {
	wantUser := int64(0)
	if !m.AssignedToUserID.IsNull() && !m.AssignedToUserID.IsUnknown() {
		wantUser = m.AssignedToUserID.ValueInt64()
	}
	wantAsset := int64(0)
	if !m.AssignedToAssetID.IsNull() && !m.AssignedToAssetID.IsUnknown() {
		wantAsset = m.AssignedToAssetID.ValueInt64()
	}
	return api.AssignedUser.IDOrZero() == wantUser && api.AssignedAsset.IDOrZero() == wantAsset
}

func (r *LicenseSeatResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LicenseSeatResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	licenseID := data.LicenseID.ValueInt64()

	// Claim a free seat. Terraform may create several seat assignments for
	// the same license concurrently and the API happily overwrites existing
	// assignments (last write wins), so after claiming we verify the seat
	// still carries our assignment and move on to the next free seat if a
	// concurrent claim stole it.
	const maxAttempts = 20
	claimed := false
	tried := map[int64]bool{}
	for attempt := 0; attempt < maxAttempts && !claimed; attempt++ {
		free, err := r.svc.FreeSeatIDs(ctx, licenseID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list license seats", err.Error())
			return
		}
		var seatID int64
		for _, id := range free {
			if !tried[id] {
				seatID = id
				break
			}
		}
		if seatID == 0 {
			resp.Diagnostics.AddError("No free license seat",
				fmt.Sprintf("License %d has no unassigned seat left; raise its seats count.", licenseID))
			return
		}
		tried[seatID] = true

		if err := r.svc.AssignSeat(ctx, licenseID, seatID, data.assignmentBody()); err != nil {
			resp.Diagnostics.AddError("Unable to assign license seat", err.Error())
			return
		}
		api, err := r.svc.GetSeat(ctx, licenseID, seatID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to verify license seat assignment", err.Error())
			return
		}
		if data.matchesAssignment(api) {
			data.SeatID = types.Int64Value(seatID)
			data.ID = types.Int64Value(seatID)
			claimed = true
		}
	}
	if !claimed {
		resp.Diagnostics.AddError("Unable to claim license seat",
			fmt.Sprintf("Could not claim a seat of license %d after %d attempts.", licenseID, maxAttempts))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LicenseSeatResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LicenseSeatResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api, err := r.svc.GetSeat(ctx, data.LicenseID.ValueInt64(), data.SeatID.ValueInt64())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read license seat", err.Error())
		return
	}
	// A seat that was released outside Terraform no longer represents this
	// assignment; drop it from state so the next apply claims a seat again.
	if api.Free() {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.Int64Value(api.Id)
	data.SeatID = types.Int64Value(api.Id)
	data.AssignedToUserID = tfutil.StateRefID(api.AssignedUser)
	data.AssignedToAssetID = tfutil.StateRefID(api.AssignedAsset)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LicenseSeatResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LicenseSeatResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	licenseID := data.LicenseID.ValueInt64()
	seatID := data.SeatID.ValueInt64()
	// Changing the assignment target type in a single PATCH fails with
	// "Target not found" on an occupied seat (v8.0.4); releasing the seat
	// first (both targets null) and then assigning works reliably.
	if err := r.svc.ReleaseSeat(ctx, licenseID, seatID); err != nil {
		resp.Diagnostics.AddError("Unable to release license seat before reassignment", err.Error())
		return
	}
	if err := r.svc.AssignSeat(ctx, licenseID, seatID, data.assignmentBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update license seat", err.Error())
		return
	}
	api, err := r.svc.GetSeat(ctx, licenseID, seatID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read license seat after update", err.Error())
		return
	}
	data.ID = types.Int64Value(api.Id)
	data.AssignedToUserID = tfutil.StateRefID(api.AssignedUser)
	data.AssignedToAssetID = tfutil.StateRefID(api.AssignedAsset)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LicenseSeatResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LicenseSeatResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.ReleaseSeat(ctx, data.LicenseID.ValueInt64(), data.SeatID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to release license seat", err.Error())
	}
}

func (r *LicenseSeatResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID",
			fmt.Sprintf("Expected \"<license_id>/<seat_id>\", got %q.", req.ID))
		return
	}
	licenseID, err1 := strconv.ParseInt(parts[0], 10, 64)
	seatID, err2 := strconv.ParseInt(parts[1], 10, 64)
	if err1 != nil || err2 != nil {
		resp.Diagnostics.AddError("Invalid Import ID",
			fmt.Sprintf("Expected numeric ids in \"<license_id>/<seat_id>\", got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("license_id"), licenseID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("seat_id"), seatID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), seatID)...)
}
