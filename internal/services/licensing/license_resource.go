// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package licensing

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	licensingapi "github.com/timrabl/terraform-provider-snipeit/internal/api/licensing"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &LicenseResource{}
	_ resource.ResourceWithImportState = &LicenseResource{}
)

// NewLicenseResource returns a new snipeit_license resource.
func NewLicenseResource() resource.Resource {
	return &LicenseResource{}
}

// LicenseResource manages a Snipe-IT license.
//
// purchase_cost is deliberately not exposed: the API returns it formatted in
// the instance locale, which does not round-trip stably in state (same
// decision as snipeit_hardware).
type LicenseResource struct {
	svc *licensingapi.Service
}

// LicenseResourceModel describes the resource data model.
type LicenseResourceModel struct {
	ID              types.Int64       `tfsdk:"id"`
	Name            types.String      `tfsdk:"name"`
	Seats           types.Int64       `tfsdk:"seats"`
	CategoryID      types.Int64       `tfsdk:"category_id"`
	CompanyID       types.Int64       `tfsdk:"company_id"`
	ManufacturerID  types.Int64       `tfsdk:"manufacturer_id"`
	SupplierID      types.Int64       `tfsdk:"supplier_id"`
	OrderNumber     types.String      `tfsdk:"order_number"`
	PurchaseOrder   types.String      `tfsdk:"purchase_order"`
	PurchaseDate    types.String      `tfsdk:"purchase_date"`
	PurchaseCost    tfutil.MoneyValue `tfsdk:"purchase_cost"`
	ExpirationDate  types.String      `tfsdk:"expiration_date"`
	TerminationDate types.String      `tfsdk:"termination_date"`
	LicenseName     types.String      `tfsdk:"license_name"`
	LicenseEmail    types.String      `tfsdk:"license_email"`
	Serial          types.String      `tfsdk:"serial"`
	Reassignable    types.Bool        `tfsdk:"reassignable"`
	Maintained      types.Bool        `tfsdk:"maintained"`
	Notes           types.String      `tfsdk:"notes"`
}

func (r *LicenseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license"
}

func (r *LicenseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a license in Snipe-IT. Snipe-IT automatically maintains one seat " +
			"per unit of `seats`; individual seats are assigned with the `snipeit_license_seat` resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the license.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the license.",
				Required:            true,
			},
			"seats": schema.Int64Attribute{
				MarkdownDescription: "Number of seats. Reducing this below the number of assigned seats fails.",
				Required:            true,
			},
			"category_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the category (must be a `license` category).",
				Required:            true,
			},
			"company_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the company owning this license.",
				Optional:            true,
			},
			"manufacturer_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the manufacturer of this license.",
				Optional:            true,
			},
			"supplier_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the supplier of this license.",
				Optional:            true,
			},
			"order_number": schema.StringAttribute{
				MarkdownDescription: "Order number of the purchase.",
				Optional:            true,
			},
			"purchase_order": schema.StringAttribute{
				MarkdownDescription: "Purchase order number.",
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
			"expiration_date": schema.StringAttribute{
				MarkdownDescription: "Expiration date in `YYYY-MM-DD` format.",
				Optional:            true,
			},
			"termination_date": schema.StringAttribute{
				MarkdownDescription: "Termination date in `YYYY-MM-DD` format.",
				Optional:            true,
			},
			"license_name": schema.StringAttribute{
				MarkdownDescription: "Name of the entity the license is registered to.",
				Optional:            true,
			},
			"license_email": schema.StringAttribute{
				MarkdownDescription: "Email address the license is registered to.",
				Optional:            true,
			},
			"serial": schema.StringAttribute{
				MarkdownDescription: "Product key / serial of the license.",
				Optional:            true,
			},
			"reassignable": schema.BoolAttribute{
				MarkdownDescription: "Whether seats can be reassigned after checkin.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"maintained": schema.BoolAttribute{
				MarkdownDescription: "Whether the license is under maintenance.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Free-form notes.",
				Optional:            true,
			},
		},
	}
}

func (r *LicenseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = licensingapi.New(c)
	}
}

func (m *LicenseResourceModel) toBody() map[string]any {
	body := map[string]any{
		"name":        m.Name.ValueString(),
		"seats":       m.Seats.ValueInt64(),
		"category_id": m.CategoryID.ValueInt64(),
	}
	tfutil.BodyNullableInt(body, "company_id", m.CompanyID)
	tfutil.BodyNullableInt(body, "manufacturer_id", m.ManufacturerID)
	tfutil.BodyNullableInt(body, "supplier_id", m.SupplierID)
	tfutil.BodyString(body, "order_number", m.OrderNumber)
	tfutil.BodyString(body, "purchase_order", m.PurchaseOrder)
	tfutil.BodyNullableString(body, "purchase_date", m.PurchaseDate)
	tfutil.BodyMoney(body, "purchase_cost", m.PurchaseCost)
	tfutil.BodyNullableString(body, "expiration_date", m.ExpirationDate)
	tfutil.BodyNullableString(body, "termination_date", m.TerminationDate)
	tfutil.BodyString(body, "license_name", m.LicenseName)
	tfutil.BodyString(body, "license_email", m.LicenseEmail)
	tfutil.BodyString(body, "serial", m.Serial)
	tfutil.BodyString(body, "notes", m.Notes)
	tfutil.BodyOptBool(body, "reassignable", m.Reassignable)
	tfutil.BodyOptBool(body, "maintained", m.Maintained)
	return body
}

// stateDate maps a nested API date to state, treating absent/empty as null.
func stateDate(d *client.Date) types.String {
	if d == nil || d.Date == "" {
		return types.StringNull()
	}
	return types.StringValue(d.Date)
}

func (m *LicenseResourceModel) fromAPI(api *licensingapi.License) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Seats = types.Int64Value(int64(api.Seats))
	m.CategoryID = types.Int64Value(api.Category.IDOrZero())
	m.CompanyID = tfutil.StateRefID(api.Company)
	m.ManufacturerID = tfutil.StateRefID(api.Manufacturer)
	m.SupplierID = tfutil.StateRefID(api.Supplier)
	m.OrderNumber = tfutil.StateStringPtr(api.OrderNumber)
	m.PurchaseOrder = tfutil.StateStringPtr(api.PurchaseOrder)
	m.PurchaseDate = stateDate(api.PurchaseDate)
	m.PurchaseCost = tfutil.StateMoneyPtr(api.PurchaseCost)
	m.ExpirationDate = stateDate(api.ExpirationDate)
	m.TerminationDate = stateDate(api.TerminationDate)
	m.LicenseName = tfutil.StateStringPtr(api.LicenseName)
	m.LicenseEmail = tfutil.StateStringPtr(api.LicenseEmail)
	m.Serial = tfutil.StateStringPtr(api.ProductKey)
	m.Reassignable = types.BoolValue(bool(api.Reassignable))
	m.Maintained = types.BoolValue(bool(api.Maintained))
	m.Notes = tfutil.StateStringPtr(api.Notes)
}

func (r *LicenseResource) read(ctx context.Context, id int64, data *LicenseResourceModel) error {
	api, err := r.svc.GetLicense(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *LicenseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LicenseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateLicense(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create license", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read license after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LicenseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LicenseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read license", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LicenseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LicenseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateLicense(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update license", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read license after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LicenseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LicenseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteLicense(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete license", err.Error())
	}
}

func (r *LicenseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
