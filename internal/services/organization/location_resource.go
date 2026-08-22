// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package organization

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	organizationapi "github.com/timrabl/terraform-provider-snipeit/internal/api/organization"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &LocationResource{}
	_ resource.ResourceWithImportState = &LocationResource{}
)

// NewLocationResource returns a new snipeit_location resource.
func NewLocationResource() resource.Resource {
	return &LocationResource{}
}

// LocationResource manages a Snipe-IT location.
type LocationResource struct {
	svc *organizationapi.Service
}

// LocationResourceModel describes the resource data model.
type LocationResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Address   types.String `tfsdk:"address"`
	Address2  types.String `tfsdk:"address2"`
	City      types.String `tfsdk:"city"`
	State     types.String `tfsdk:"state"`
	Country   types.String `tfsdk:"country"`
	Zip       types.String `tfsdk:"zip"`
	Phone     types.String `tfsdk:"phone"`
	Fax       types.String `tfsdk:"fax"`
	Currency  types.String `tfsdk:"currency"`
	ParentID  types.Int64  `tfsdk:"parent_id"`
	ManagerID types.Int64  `tfsdk:"manager_id"`
	LdapOU    types.String `tfsdk:"ldap_ou"`
	Notes     types.String `tfsdk:"notes"`
}

func (r *LocationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

func (r *LocationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	optString := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Optional: true}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a location in Snipe-IT.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the location.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the location.",
				Required:            true,
			},
			"address":  optString("Street address of the location."),
			"address2": optString("Street address of the location, line 2."),
			"city":     optString("City of the location."),
			"state":    optString("State/province of the location."),
			"country":  optString("Country of the location (two-letter code recommended)."),
			"zip":      optString("Postal code of the location."),
			"phone":    optString("Phone number of the location."),
			"fax":      optString("Fax number of the location."),
			"currency": optString("Currency used at the location, e.g. `EUR`."),
			"parent_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the parent location.",
				Optional:            true,
			},
			"manager_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the user managing this location.",
				Optional:            true,
			},
			"ldap_ou": optString("LDAP organizational unit associated with the location."),
			"notes":   optString("Free-form notes."),
		},
	}
}

func (r *LocationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = organizationapi.New(c)
	}
}

func (m *LocationResourceModel) toBody() map[string]any {
	body := map[string]any{"name": m.Name.ValueString()}
	tfutil.BodyString(body, "address", m.Address)
	tfutil.BodyString(body, "address2", m.Address2)
	tfutil.BodyString(body, "city", m.City)
	tfutil.BodyString(body, "state", m.State)
	tfutil.BodyString(body, "country", m.Country)
	tfutil.BodyString(body, "zip", m.Zip)
	tfutil.BodyString(body, "phone", m.Phone)
	tfutil.BodyString(body, "fax", m.Fax)
	tfutil.BodyString(body, "currency", m.Currency)
	tfutil.BodyString(body, "ldap_ou", m.LdapOU)
	tfutil.BodyString(body, "notes", m.Notes)
	tfutil.BodyNullableInt(body, "parent_id", m.ParentID)
	tfutil.BodyNullableInt(body, "manager_id", m.ManagerID)
	return body
}

func (m *LocationResourceModel) fromAPI(api *organizationapi.Location) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Address = tfutil.StateStringPtrKeep(api.Address, m.Address)
	m.Address2 = tfutil.StateStringPtrKeep(api.Address2, m.Address2)
	m.City = tfutil.StateStringPtrKeep(api.City, m.City)
	m.State = tfutil.StateStringPtrKeep(api.State, m.State)
	m.Country = tfutil.StateStringPtrKeep(api.Country, m.Country)
	m.Zip = tfutil.StateStringPtrKeep(api.Zip, m.Zip)
	m.Phone = tfutil.StateStringPtrKeep(api.Phone, m.Phone)
	m.Fax = tfutil.StateStringPtrKeep(api.Fax, m.Fax)
	m.Currency = tfutil.StateStringPtrKeep(api.Currency, m.Currency)
	m.ParentID = tfutil.StateRefID(api.Parent)
	m.ManagerID = tfutil.StateRefID(api.Manager)
	m.LdapOU = tfutil.StateStringPtrKeep(api.LdapOu, m.LdapOU)
	m.Notes = tfutil.StateStringPtrKeep(api.Notes, m.Notes)
}

func (r *LocationResource) read(ctx context.Context, id int64, data *LocationResourceModel) error {
	api, err := r.svc.GetLocation(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *LocationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LocationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateLocation(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create location", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read location after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LocationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LocationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read location", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LocationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LocationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateLocation(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update location", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read location after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LocationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LocationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteLocation(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete location", err.Error())
	}
}

func (r *LocationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
