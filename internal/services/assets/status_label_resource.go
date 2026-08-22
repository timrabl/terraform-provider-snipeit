// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package assets

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	assetsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/assets"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &StatusLabelResource{}
	_ resource.ResourceWithImportState = &StatusLabelResource{}
)

// NewStatusLabelResource returns a new snipeit_status_label resource.
func NewStatusLabelResource() resource.Resource {
	return &StatusLabelResource{}
}

// StatusLabelResource manages a Snipe-IT status label.
type StatusLabelResource struct {
	svc *assetsapi.Service
}

// StatusLabelResourceModel describes the resource data model.
type StatusLabelResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Color        types.String `tfsdk:"color"`
	ShowInNav    types.Bool   `tfsdk:"show_in_nav"`
	DefaultLabel types.Bool   `tfsdk:"default_label"`
	Notes        types.String `tfsdk:"notes"`
}

func (r *StatusLabelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_label"
}

func (r *StatusLabelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a status label in Snipe-IT. Status labels describe the lifecycle " +
			"state of assets (deployable, pending, undeployable, archived).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the status label.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the status label. Must be unique.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Meta type of the label. One of `deployable`, `pending`, " +
					"`undeployable` or `archived`.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("deployable", "pending", "undeployable", "archived"),
				},
			},
			"color": schema.StringAttribute{
				MarkdownDescription: "Hex color for charts, e.g. `#00ff00`.",
				Optional:            true,
			},
			"show_in_nav": schema.BoolAttribute{
				MarkdownDescription: "Show assets with this label in the side navigation.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"default_label": schema.BoolAttribute{
				MarkdownDescription: "Pin this label to the top of the status dropdown.",
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

func (r *StatusLabelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = assetsapi.New(c)
	}
}

func (m *StatusLabelResourceModel) toBody() map[string]any {
	body := map[string]any{
		"name": m.Name.ValueString(),
		"type": m.Type.ValueString(),
	}
	tfutil.BodyString(body, "color", m.Color)
	tfutil.BodyString(body, "notes", m.Notes)
	tfutil.BodyOptBool(body, "show_in_nav", m.ShowInNav)
	tfutil.BodyOptBool(body, "default_label", m.DefaultLabel)
	return body
}

func (m *StatusLabelResourceModel) fromAPI(api *assetsapi.StatusLabel) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Type = types.StringValue(strings.ToLower(api.Type))
	m.Color = tfutil.StateStringPtrKeep(api.Color, m.Color)
	m.ShowInNav = types.BoolValue(bool(api.ShowInNav))
	m.DefaultLabel = types.BoolValue(bool(api.DefaultLabel))
	m.Notes = tfutil.StateStringPtrKeep(api.Notes, m.Notes)
}

func (r *StatusLabelResource) read(ctx context.Context, id int64, data *StatusLabelResourceModel) error {
	api, err := r.svc.GetStatusLabel(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *StatusLabelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StatusLabelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateStatusLabel(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create status label", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read status label after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusLabelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StatusLabelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read status label", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusLabelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StatusLabelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateStatusLabel(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update status label", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read status label after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusLabelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StatusLabelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteStatusLabel(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete status label", err.Error())
	}
}

func (r *StatusLabelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
