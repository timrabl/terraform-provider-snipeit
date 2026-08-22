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
	_ resource.Resource                = &DepartmentResource{}
	_ resource.ResourceWithImportState = &DepartmentResource{}
)

// NewDepartmentResource returns a new snipeit_department resource.
func NewDepartmentResource() resource.Resource {
	return &DepartmentResource{}
}

// DepartmentResource manages a Snipe-IT department.
type DepartmentResource struct {
	svc *organizationapi.Service
}

// DepartmentResourceModel describes the resource data model.
type DepartmentResourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	CompanyID  types.Int64  `tfsdk:"company_id"`
	ManagerID  types.Int64  `tfsdk:"manager_id"`
	LocationID types.Int64  `tfsdk:"location_id"`
	Notes      types.String `tfsdk:"notes"`
}

func (r *DepartmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_department"
}

func (r *DepartmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a department in Snipe-IT.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the department.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the department.",
				Required:            true,
			},
			"company_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the company this department belongs to.",
				Optional:            true,
			},
			"manager_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the user managing this department.",
				Optional:            true,
			},
			"location_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the location of this department.",
				Optional:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Free-form notes.",
				Optional:            true,
			},
		},
	}
}

func (r *DepartmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = organizationapi.New(c)
	}
}

func (m *DepartmentResourceModel) toBody() map[string]any {
	body := map[string]any{"name": m.Name.ValueString()}
	tfutil.BodyNullableInt(body, "company_id", m.CompanyID)
	tfutil.BodyNullableInt(body, "manager_id", m.ManagerID)
	tfutil.BodyNullableInt(body, "location_id", m.LocationID)
	tfutil.BodyString(body, "notes", m.Notes)
	return body
}

func (m *DepartmentResourceModel) fromAPI(api *organizationapi.Department) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.CompanyID = tfutil.StateRefID(api.Company)
	m.ManagerID = tfutil.StateRefID(api.Manager)
	m.LocationID = tfutil.StateRefID(api.Location)
	m.Notes = tfutil.StateStringPtr(api.Notes)
}

func (r *DepartmentResource) read(ctx context.Context, id int64, data *DepartmentResourceModel) error {
	api, err := r.svc.GetDepartment(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *DepartmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DepartmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateDepartment(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create department", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read department after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DepartmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DepartmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read department", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DepartmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DepartmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateDepartment(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update department", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read department after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DepartmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DepartmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteDepartment(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete department", err.Error())
	}
}

func (r *DepartmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
