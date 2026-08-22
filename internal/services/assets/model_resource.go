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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	assetsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/assets"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &ModelResource{}
	_ resource.ResourceWithImportState = &ModelResource{}
)

// NewModelResource returns a new snipeit_model resource.
func NewModelResource() resource.Resource {
	return &ModelResource{}
}

// ModelResource manages a Snipe-IT asset model.
type ModelResource struct {
	svc *assetsapi.Service
}

// ModelResourceModel describes the resource data model.
type ModelResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ModelNumber    types.String `tfsdk:"model_number"`
	CategoryID     types.Int64  `tfsdk:"category_id"`
	ManufacturerID types.Int64  `tfsdk:"manufacturer_id"`
	FieldsetID     types.Int64  `tfsdk:"fieldset_id"`
	EOL            types.Int64  `tfsdk:"eol"`
	MinAmt         types.Int64  `tfsdk:"min_amt"`
	Requestable    types.Bool   `tfsdk:"requestable"`
	Notes          types.String `tfsdk:"notes"`
}

func (r *ModelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (r *ModelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an asset model in Snipe-IT. Assets reference a model, which in " +
			"turn defines category, manufacturer and custom fieldset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the model.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the model.",
				Required:            true,
			},
			"model_number": schema.StringAttribute{
				MarkdownDescription: "Manufacturer model number.",
				Optional:            true,
			},
			"category_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the category of this model (must be an `asset` category).",
				Required:            true,
			},
			"manufacturer_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the manufacturer of this model.",
				Optional:            true,
			},
			"fieldset_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the custom fieldset attached to this model.",
				Optional:            true,
			},
			"eol": schema.Int64Attribute{
				MarkdownDescription: "End-of-life in months.",
				Optional:            true,
			},
			"min_amt": schema.Int64Attribute{
				MarkdownDescription: "Minimum quantity before a low-stock alert triggers.",
				Optional:            true,
			},
			"requestable": schema.BoolAttribute{
				MarkdownDescription: "Whether assets of this model can be requested by users.",
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

func (r *ModelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = assetsapi.New(c)
	}
}

func (m *ModelResourceModel) toBody() map[string]any {
	body := map[string]any{
		"name":        m.Name.ValueString(),
		"category_id": m.CategoryID.ValueInt64(),
	}
	tfutil.BodyString(body, "model_number", m.ModelNumber)
	tfutil.BodyString(body, "notes", m.Notes)
	tfutil.BodyNullableInt(body, "manufacturer_id", m.ManufacturerID)
	tfutil.BodyNullableInt(body, "fieldset_id", m.FieldsetID)
	tfutil.BodyNullableInt(body, "eol", m.EOL)
	tfutil.BodyNullableInt(body, "min_amt", m.MinAmt)
	tfutil.BodyOptBool(body, "requestable", m.Requestable)
	return body
}

func (m *ModelResourceModel) fromAPI(api *assetsapi.Model) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.ModelNumber = tfutil.StateStringPtrKeep(api.ModelNumber, m.ModelNumber)
	m.CategoryID = types.Int64Value(api.Category.IDOrZero())
	m.ManufacturerID = tfutil.StateRefID(api.Manufacturer)
	m.FieldsetID = tfutil.StateRefID(api.Fieldset)
	m.EOL = tfutil.StateOptIntKeep(int64(api.Eol), m.EOL)
	m.MinAmt = tfutil.StateOptIntKeep(int64(api.MinAmt), m.MinAmt)
	m.Requestable = types.BoolValue(bool(api.Requestable))
	m.Notes = tfutil.StateStringPtrKeep(api.Notes, m.Notes)
}

func (r *ModelResource) read(ctx context.Context, id int64, data *ModelResourceModel) error {
	api, err := r.svc.GetModel(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *ModelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ModelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateModel(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create model", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read model after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ModelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read model", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ModelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateModel(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update model", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read model after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ModelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteModel(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete model", err.Error())
	}
}

func (r *ModelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
