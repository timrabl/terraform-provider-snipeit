// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	inventoryapi "github.com/timrabl/terraform-provider-snipeit/internal/api/inventory"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var _ resource.Resource = &ConsumableCheckoutResource{}

// NewConsumableCheckoutResource returns a new snipeit_consumable_checkout resource.
func NewConsumableCheckoutResource() resource.Resource {
	return &ConsumableCheckoutResource{}
}

// ConsumableCheckoutResource models the irreversible consumption of one unit
// of a consumable by a user.
type ConsumableCheckoutResource struct {
	svc *inventoryapi.Service
}

// ConsumableCheckoutResourceModel describes the resource data model.
type ConsumableCheckoutResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	ConsumableID types.Int64  `tfsdk:"consumable_id"`
	UserID       types.Int64  `tfsdk:"user_id"`
	Note         types.String `tfsdk:"note"`
}

func (r *ConsumableCheckoutResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_consumable_checkout"
}

func (r *ConsumableCheckoutResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Checks one unit of a consumable out to a user. **Consumption is " +
			"irreversible**: the Snipe-IT API has no consumable checkin, so destroying this resource " +
			"only removes it from the Terraform state — the consumed unit is not returned to stock. " +
			"The API also exposes no per-checkout identity, so drift (changes made outside Terraform) " +
			"cannot be detected for this resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Synthetic id (equal to `consumable_id`); the API exposes no checkout row id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"consumable_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the consumable to check out.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the user consuming the unit.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "Note recorded with the checkout.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ConsumableCheckoutResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = inventoryapi.New(c)
	}
}

func (r *ConsumableCheckoutResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConsumableCheckoutResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	consumableID := data.ConsumableID.ValueInt64()
	body := map[string]any{"assigned_to": data.UserID.ValueInt64()}
	if !data.Note.IsNull() && !data.Note.IsUnknown() {
		body["note"] = data.Note.ValueString()
	}
	if err := r.svc.CheckoutConsumable(ctx, consumableID, body); err != nil {
		resp.Diagnostics.AddError("Unable to check out consumable", err.Error())
		return
	}

	data.ID = types.Int64Value(consumableID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConsumableCheckoutResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// The API provides no identity for individual consumable checkouts
	// (GET /consumables/{id}/users returns HTML-decorated rows without ids),
	// so the state is kept as-is. See the resource description.
	var data ConsumableCheckoutResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConsumableCheckoutResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes RequiresReplace; Update is never called.
	resp.Diagnostics.AddError("Unexpected update", "snipeit_consumable_checkout does not support in-place updates.")
}

func (r *ConsumableCheckoutResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No consumable checkin exists in the Snipe-IT API; deletion only forgets
	// the checkout in state. Documented in the resource description.
}
