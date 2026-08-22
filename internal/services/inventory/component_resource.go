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
	_ resource.Resource                = &ComponentResource{}
	_ resource.ResourceWithImportState = &ComponentResource{}
)

// NewComponentResource returns a new snipeit_component resource.
func NewComponentResource() resource.Resource {
	return &ComponentResource{}
}

// ComponentResource manages a Snipe-IT component.
type ComponentResource struct {
	svc *inventoryapi.Service
}

// ComponentResourceModel describes the resource data model.
type ComponentResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Qty          types.Int64  `tfsdk:"qty"`
	CategoryID   types.Int64  `tfsdk:"category_id"`
	SupplierID   types.Int64  `tfsdk:"supplier_id"`
	CompanyID    types.Int64  `tfsdk:"company_id"`
	LocationID   types.Int64  `tfsdk:"location_id"`
	Serial       types.String `tfsdk:"serial"`
	OrderNumber  types.String `tfsdk:"order_number"`
	PurchaseDate types.String `tfsdk:"purchase_date"`
	MinAmt       types.Int64  `tfsdk:"min_amt"`
	Notes        types.String `tfsdk:"notes"`
}

func (r *ComponentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_component"
}

func (r *ComponentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a component in Snipe-IT. `purchase_cost` is intentionally not " +
			"supported (the API returns it locale-formatted, which does not round-trip stably).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the component.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the component.",
				Required:            true,
			},
			"qty": schema.Int64Attribute{
				MarkdownDescription: "Total quantity of this component.",
				Required:            true,
			},
			"category_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the category (must be a `component` category).",
				Required:            true,
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
			"serial": schema.StringAttribute{
				MarkdownDescription: "Serial number.",
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

func (r *ComponentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = inventoryapi.New(c)
	}
}

func (m *ComponentResourceModel) toBody() map[string]any {
	body := map[string]any{
		"name":        m.Name.ValueString(),
		"qty":         m.Qty.ValueInt64(),
		"category_id": m.CategoryID.ValueInt64(),
	}
	tfutil.BodyNullableInt(body, "supplier_id", m.SupplierID)
	tfutil.BodyNullableInt(body, "company_id", m.CompanyID)
	tfutil.BodyNullableInt(body, "location_id", m.LocationID)
	tfutil.BodyString(body, "serial", m.Serial)
	tfutil.BodyString(body, "order_number", m.OrderNumber)
	tfutil.BodyNullableString(body, "purchase_date", m.PurchaseDate)
	tfutil.BodyNullableInt(body, "min_amt", m.MinAmt)
	tfutil.BodyString(body, "notes", m.Notes)
	return body
}

func (m *ComponentResourceModel) fromAPI(api *inventoryapi.Component) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Qty = types.Int64Value(int64(api.Qty))
	m.CategoryID = types.Int64Value(api.Category.IDOrZero())
	m.SupplierID = tfutil.StateRefID(api.Supplier)
	m.CompanyID = tfutil.StateRefID(api.Company)
	m.LocationID = tfutil.StateRefID(api.Location)
	m.Serial = tfutil.StateStringPtr(api.Serial)
	m.OrderNumber = tfutil.StateStringPtr(api.OrderNumber)
	m.MinAmt = tfutil.StateOptInt(int64(api.MinAmt))
	m.Notes = tfutil.StateStringPtr(api.Notes)
	if api.PurchaseDate != nil && api.PurchaseDate.Date != "" {
		m.PurchaseDate = types.StringValue(api.PurchaseDate.Date)
	} else {
		m.PurchaseDate = types.StringNull()
	}
}

func (r *ComponentResource) read(ctx context.Context, id int64, data *ComponentResourceModel) error {
	api, err := r.svc.GetComponent(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *ComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ComponentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateComponent(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create component", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read component after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ComponentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read component", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ComponentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateComponent(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update component", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read component after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ComponentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteComponent(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete component", err.Error())
	}
}

func (r *ComponentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
