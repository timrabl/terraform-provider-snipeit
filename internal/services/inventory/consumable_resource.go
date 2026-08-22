// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	inventoryapi "github.com/timrabl/terraform-provider-snipeit/internal/api/inventory"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &ConsumableResource{}
	_ resource.ResourceWithImportState = &ConsumableResource{}
)

// NewConsumableResource returns a new snipeit_consumable resource.
func NewConsumableResource() resource.Resource {
	return &ConsumableResource{}
}

// ConsumableResource manages a Snipe-IT consumable.
type ConsumableResource struct {
	svc *inventoryapi.Service
}

// ConsumableResourceModel describes the resource data model.
type ConsumableResourceModel struct {
	ID             types.Int64       `tfsdk:"id"`
	Name           types.String      `tfsdk:"name"`
	Qty            types.Int64       `tfsdk:"qty"`
	CategoryID     types.Int64       `tfsdk:"category_id"`
	CompanyID      types.Int64       `tfsdk:"company_id"`
	ManufacturerID types.Int64       `tfsdk:"manufacturer_id"`
	LocationID     types.Int64       `tfsdk:"location_id"`
	ItemNo         types.String      `tfsdk:"item_no"`
	ModelNumber    types.String      `tfsdk:"model_number"`
	OrderNumber    types.String      `tfsdk:"order_number"`
	PurchaseDate   types.String      `tfsdk:"purchase_date"`
	PurchaseCost   tfutil.MoneyValue `tfsdk:"purchase_cost"`
	MinAmt         types.Int64       `tfsdk:"min_amt"`
	Notes          types.String      `tfsdk:"notes"`
}

func (r *ConsumableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_consumable"
}

func (r *ConsumableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a consumable in Snipe-IT. `purchase_cost` is intentionally not " +
			"supported (the API returns it locale-formatted, which does not round-trip stably).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the consumable.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the consumable.",
				Required:            true,
			},
			"qty": schema.Int64Attribute{
				MarkdownDescription: "Total quantity of this consumable.",
				Required:            true,
			},
			"category_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the category (must be a `consumable` category).",
				Required:            true,
			},
			"company_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the owning company.",
				Optional:            true,
			},
			"manufacturer_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the manufacturer.",
				Optional:            true,
			},
			"location_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the location.",
				Optional:            true,
			},
			"item_no": schema.StringAttribute{
				MarkdownDescription: "Item number.",
				Optional:            true,
			},
			"model_number": schema.StringAttribute{
				MarkdownDescription: "Manufacturer model number.",
				Optional:            true,
			},
			"order_number": schema.StringAttribute{
				MarkdownDescription: "Order number of the purchase.",
				Optional:            true,
			},
			"purchase_date": schema.StringAttribute{
				MarkdownDescription: "Purchase date in `YYYY-MM-DD` format.",
				Optional:            true,
			},
			"purchase_cost": schema.StringAttribute{
				MarkdownDescription: "Purchase cost as a decimal string, e.g. `1234.50`. " +
					"Stored normalized; `1234.5`, `1234.50` and `1,234.50` are the same value.",
				CustomType: tfutil.MoneyType{},
				Optional:   true,
			},
			"min_amt": schema.Int64Attribute{
				MarkdownDescription: "Minimum quantity before a low-stock alert triggers.",
				Optional:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Free-form notes.",
				Optional:            true,
			},
		},
	}
}

func (r *ConsumableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = inventoryapi.New(c)
	}
}

func (m *ConsumableResourceModel) toBody() map[string]any {
	body := map[string]any{
		"name":        m.Name.ValueString(),
		"qty":         m.Qty.ValueInt64(),
		"category_id": m.CategoryID.ValueInt64(),
	}
	tfutil.BodyNullableInt(body, "company_id", m.CompanyID)
	tfutil.BodyNullableInt(body, "manufacturer_id", m.ManufacturerID)
	tfutil.BodyNullableInt(body, "location_id", m.LocationID)
	tfutil.BodyString(body, "item_no", m.ItemNo)
	tfutil.BodyString(body, "model_number", m.ModelNumber)
	tfutil.BodyString(body, "order_number", m.OrderNumber)
	tfutil.BodyNullableString(body, "purchase_date", m.PurchaseDate)
	tfutil.BodyMoney(body, "purchase_cost", m.PurchaseCost)
	tfutil.BodyNullableInt(body, "min_amt", m.MinAmt)
	tfutil.BodyString(body, "notes", m.Notes)
	return body
}

func (m *ConsumableResourceModel) fromAPI(api *inventoryapi.Consumable) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Qty = types.Int64Value(int64(api.Qty))
	m.CategoryID = types.Int64Value(api.Category.IDOrZero())
	m.CompanyID = tfutil.StateRefID(api.Company)
	m.ManufacturerID = tfutil.StateRefID(api.Manufacturer)
	m.LocationID = tfutil.StateRefID(api.Location)
	m.ItemNo = tfutil.StateStringPtrKeep(api.ItemNo, m.ItemNo)
	m.ModelNumber = tfutil.StateStringPtrKeep(api.ModelNumber, m.ModelNumber)
	m.OrderNumber = tfutil.StateStringPtrKeep(api.OrderNumber, m.OrderNumber)
	m.MinAmt = tfutil.StateOptIntKeep(int64(api.MinAmt), m.MinAmt)
	m.Notes = tfutil.StateStringPtrKeep(api.Notes, m.Notes)
	if api.PurchaseDate != nil && api.PurchaseDate.Date != "" {
		m.PurchaseDate = types.StringValue(api.PurchaseDate.Date)
	} else {
		m.PurchaseDate = types.StringNull()
	}
	m.PurchaseCost = tfutil.StateMoneyPtr(api.PurchaseCost)
}

func (r *ConsumableResource) read(ctx context.Context, id int64, data *ConsumableResourceModel) error {
	api, err := r.svc.GetConsumable(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *ConsumableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConsumableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateConsumable(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create consumable", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read consumable after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConsumableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConsumableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read consumable", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConsumableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConsumableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateConsumable(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update consumable", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read consumable after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConsumableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConsumableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteConsumable(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete consumable", err.Error())
	}
}

func (r *ConsumableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
