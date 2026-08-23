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
	_ resource.Resource                = &AccessoryResource{}
	_ resource.ResourceWithImportState = &AccessoryResource{}
)

// NewAccessoryResource returns a new snipeit_accessory resource.
func NewAccessoryResource() resource.Resource {
	return &AccessoryResource{}
}

// AccessoryResource manages a Snipe-IT accessory.
type AccessoryResource struct {
	svc *inventoryapi.Service
}

// AccessoryResourceModel describes the resource data model.
type AccessoryResourceModel struct {
	ID             types.Int64       `tfsdk:"id"`
	Name           types.String      `tfsdk:"name"`
	Qty            types.Int64       `tfsdk:"qty"`
	CategoryID     types.Int64       `tfsdk:"category_id"`
	ManufacturerID types.Int64       `tfsdk:"manufacturer_id"`
	SupplierID     types.Int64       `tfsdk:"supplier_id"`
	CompanyID      types.Int64       `tfsdk:"company_id"`
	LocationID     types.Int64       `tfsdk:"location_id"`
	ModelNumber    types.String      `tfsdk:"model_number"`
	OrderNumber    types.String      `tfsdk:"order_number"`
	PurchaseDate   types.String      `tfsdk:"purchase_date"`
	PurchaseCost   tfutil.MoneyValue `tfsdk:"purchase_cost"`
	MinAmt         types.Int64       `tfsdk:"min_amt"`
	Notes          types.String      `tfsdk:"notes"`
}

func (r *AccessoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_accessory"
}

func (r *AccessoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an accessory in Snipe-IT. `purchase_cost` is intentionally not " +
			"supported (the API returns it locale-formatted, which does not round-trip stably).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the accessory.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the accessory.",
				Required:            true,
			},
			"qty": schema.Int64Attribute{
				MarkdownDescription: "Total quantity of this accessory.",
				Required:            true,
			},
			"category_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the category (must be an `accessory` category).",
				Required:            true,
			},
			"manufacturer_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the manufacturer.",
				Optional:            true,
			},
			"supplier_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the supplier.",
				Optional:            true,
			},
			"company_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the owning company.",
				Optional:            true,
			},
			"location_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the location.",
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

func (r *AccessoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = inventoryapi.New(c)
	}
}

func (m *AccessoryResourceModel) toBody() map[string]any {
	body := map[string]any{
		"name":        m.Name.ValueString(),
		"qty":         m.Qty.ValueInt64(),
		"category_id": m.CategoryID.ValueInt64(),
	}
	tfutil.BodyNullableInt(body, "manufacturer_id", m.ManufacturerID)
	tfutil.BodyNullableInt(body, "supplier_id", m.SupplierID)
	tfutil.BodyNullableInt(body, "company_id", m.CompanyID)
	tfutil.BodyNullableInt(body, "location_id", m.LocationID)
	tfutil.BodyString(body, "model_number", m.ModelNumber)
	tfutil.BodyString(body, "order_number", m.OrderNumber)
	tfutil.BodyNullableString(body, "purchase_date", m.PurchaseDate)
	tfutil.BodyMoney(body, "purchase_cost", m.PurchaseCost)
	tfutil.BodyNullableInt(body, "min_amt", m.MinAmt)
	tfutil.BodyString(body, "notes", m.Notes)
	return body
}

func (m *AccessoryResourceModel) fromAPI(api *inventoryapi.Accessory) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Qty = types.Int64Value(int64(api.Qty))
	m.CategoryID = types.Int64Value(api.Category.IDOrZero())
	m.ManufacturerID = tfutil.StateRefID(api.Manufacturer)
	m.SupplierID = tfutil.StateRefID(api.Supplier)
	m.CompanyID = tfutil.StateRefID(api.Company)
	m.LocationID = tfutil.StateRefID(api.Location)
	m.ModelNumber = tfutil.StateStringPtrKeep(api.ModelNumber, m.ModelNumber)
	m.OrderNumber = tfutil.StateStringPtrPreserve(api.OrderNumber, m.OrderNumber)
	// The read serializer returns the write field min_amt as min_qty.
	m.MinAmt = tfutil.StateOptIntKeep(int64(api.MinQty), m.MinAmt)
	m.Notes = tfutil.StateStringPtrKeep(api.Notes, m.Notes)
	// Snipe-IT 8.7 ignores clearing purchase_date/purchase_cost on inventory
	// items and keeps echoing the old value; the clear-aware mappers keep the
	// field null once the user has cleared it (prior null) so the apply stays
	// consistent. On 8.0.4 the clear works server-side, so this is a no-op.
	m.PurchaseDate = tfutil.StateDateClearAware(api.PurchaseDate, m.PurchaseDate)
	m.PurchaseCost = tfutil.StateMoneyPtrClearAware(api.PurchaseCost, m.PurchaseCost)
}

func (r *AccessoryResource) read(ctx context.Context, id int64, data *AccessoryResourceModel) error {
	api, err := r.svc.GetAccessory(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *AccessoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateAccessory(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create accessory", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read accessory after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read accessory", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateAccessory(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update accessory", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read accessory after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteAccessory(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete accessory", err.Error())
	}
}

func (r *AccessoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
