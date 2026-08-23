// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package assets

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	assetsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/assets"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &ManufacturerResource{}
	_ resource.ResourceWithImportState = &ManufacturerResource{}
)

// NewManufacturerResource returns a new snipeit_manufacturer resource.
func NewManufacturerResource() resource.Resource {
	return &ManufacturerResource{}
}

// ManufacturerResource manages a Snipe-IT manufacturer.
type ManufacturerResource struct {
	svc *assetsapi.Service
}

// ManufacturerResourceModel describes the resource data model.
type ManufacturerResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	URL               types.String `tfsdk:"url"`
	SupportURL        types.String `tfsdk:"support_url"`
	WarrantyLookupURL types.String `tfsdk:"warranty_lookup_url"`
	SupportPhone      types.String `tfsdk:"support_phone"`
	SupportEmail      types.String `tfsdk:"support_email"`
	Notes             types.String `tfsdk:"notes"`
}

func (r *ManufacturerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_manufacturer"
}

func (r *ManufacturerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a manufacturer in Snipe-IT.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the manufacturer.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the manufacturer. Must be unique.",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Website of the manufacturer.",
				Optional:            true,
			},
			"support_url": schema.StringAttribute{
				MarkdownDescription: "Support website of the manufacturer.",
				Optional:            true,
			},
			"warranty_lookup_url": schema.StringAttribute{
				MarkdownDescription: "Warranty lookup URL of the manufacturer.",
				Optional:            true,
			},
			"support_phone": schema.StringAttribute{
				MarkdownDescription: "Support phone number of the manufacturer.",
				Optional:            true,
			},
			"support_email": schema.StringAttribute{
				MarkdownDescription: "Support email address of the manufacturer.",
				Optional:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Free-form notes.",
				Optional:            true,
			},
		},
	}
}

func (r *ManufacturerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = assetsapi.New(c)
	}
}

func (m *ManufacturerResourceModel) toBody() map[string]any {
	body := map[string]any{"name": m.Name.ValueString()}
	tfutil.BodyString(body, "url", m.URL)
	tfutil.BodyString(body, "support_url", m.SupportURL)
	tfutil.BodyString(body, "warranty_lookup_url", m.WarrantyLookupURL)
	tfutil.BodyString(body, "support_phone", m.SupportPhone)
	tfutil.BodyString(body, "support_email", m.SupportEmail)
	tfutil.BodyString(body, "notes", m.Notes)
	return body
}

func (m *ManufacturerResourceModel) fromAPI(api *assetsapi.Manufacturer) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.URL = tfutil.StateStringKeep(api.Url, m.URL)
	m.SupportURL = tfutil.StateStringKeep(api.SupportUrl, m.SupportURL)
	m.WarrantyLookupURL = tfutil.StateStringKeep(api.WarrantyLookupUrl, m.WarrantyLookupURL)
	m.SupportPhone = tfutil.StateStringKeep(api.SupportPhone, m.SupportPhone)
	m.SupportEmail = tfutil.StateStringKeep(api.SupportEmail, m.SupportEmail)
	m.Notes = tfutil.StateStringPtrPreserve(api.Notes, m.Notes)
}

func (r *ManufacturerResource) read(ctx context.Context, id int64, data *ManufacturerResourceModel) error {
	api, err := r.svc.GetManufacturer(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *ManufacturerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ManufacturerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateManufacturer(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create manufacturer", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read manufacturer after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ManufacturerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ManufacturerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read manufacturer", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ManufacturerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ManufacturerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateManufacturer(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update manufacturer", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read manufacturer after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ManufacturerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ManufacturerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteManufacturer(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete manufacturer", err.Error())
	}
}

func (r *ManufacturerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
