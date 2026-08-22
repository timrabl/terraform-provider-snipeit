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
	_ resource.Resource                = &CompanyResource{}
	_ resource.ResourceWithImportState = &CompanyResource{}
)

// NewCompanyResource returns a new snipeit_company resource.
func NewCompanyResource() resource.Resource {
	return &CompanyResource{}
}

// CompanyResource manages a Snipe-IT company.
type CompanyResource struct {
	svc *organizationapi.Service
}

// CompanyResourceModel describes the resource data model.
type CompanyResourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Phone types.String `tfsdk:"phone"`
	Fax   types.String `tfsdk:"fax"`
	Email types.String `tfsdk:"email"`
	Notes types.String `tfsdk:"notes"`
}

func (r *CompanyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_company"
}

func (r *CompanyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a company in Snipe-IT.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the company.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the company. Must be unique.",
				Required:            true,
			},
			"phone": schema.StringAttribute{
				MarkdownDescription: "Phone number of the company.",
				Optional:            true,
			},
			"fax": schema.StringAttribute{
				MarkdownDescription: "Fax number of the company.",
				Optional:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address of the company.",
				Optional:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Free-form notes.",
				Optional:            true,
			},
		},
	}
}

func (r *CompanyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = organizationapi.New(c)
	}
}

func (m *CompanyResourceModel) toBody() map[string]any {
	body := map[string]any{"name": m.Name.ValueString()}
	tfutil.BodyString(body, "phone", m.Phone)
	tfutil.BodyString(body, "fax", m.Fax)
	tfutil.BodyString(body, "email", m.Email)
	tfutil.BodyString(body, "notes", m.Notes)
	return body
}

func (m *CompanyResourceModel) fromAPI(api *organizationapi.Company) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Phone = tfutil.StateStringPtr(api.Phone)
	m.Fax = tfutil.StateStringPtr(api.Fax)
	m.Email = tfutil.StateStringPtr(api.Email)
	m.Notes = tfutil.StateStringPtr(api.Notes)
}

func (r *CompanyResource) read(ctx context.Context, id int64, data *CompanyResourceModel) error {
	api, err := r.svc.GetCompany(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *CompanyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CompanyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateCompany(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create company", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read company after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CompanyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CompanyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read company", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CompanyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CompanyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateCompany(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update company", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read company after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CompanyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CompanyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteCompany(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete company", err.Error())
	}
}

func (r *CompanyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
