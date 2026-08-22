// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package customfields

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	customfieldsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/customfields"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &FieldResource{}
	_ resource.ResourceWithImportState = &FieldResource{}
)

// NewFieldResource returns a new snipeit_field resource.
func NewFieldResource() resource.Resource {
	return &FieldResource{}
}

// FieldResource manages a Snipe-IT custom field.
type FieldResource struct {
	svc *customfieldsapi.Service
}

// FieldResourceModel describes the resource data model.
type FieldResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Element      types.String `tfsdk:"element"`
	Format       types.String `tfsdk:"format"`
	FieldValues  types.String `tfsdk:"field_values"`
	HelpText     types.String `tfsdk:"help_text"`
	ShowInEmail  types.Bool   `tfsdk:"show_in_email"`
	DBColumnName types.String `tfsdk:"db_column_name"`
}

// dbColumnNameModifier keeps the known db_column_name from state unless the
// field name changes, in which case the server regenerates the column name and
// the planned value must become unknown.
type dbColumnNameModifier struct{}

func (m dbColumnNameModifier) Description(_ context.Context) string {
	return "Recomputed when the field name changes, otherwise unchanged."
}

func (m dbColumnNameModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m dbColumnNameModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var stateName, planName types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("name"), &stateName)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("name"), &planName)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if stateName.Equal(planName) && !req.StateValue.IsNull() {
		resp.PlanValue = req.StateValue
	}
	// Otherwise leave the planned value unknown so the server-side rename can
	// land without an inconsistency error.
}

func (r *FieldResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_field"
}

func (r *FieldResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom field in Snipe-IT. Add it to a fieldset with " +
			"`snipeit_field_fieldset_association` to use it on asset models.\n\n" +
			"~> Custom regex formats (`CUSTOM REGEX` / `regex:...`) are not supported: the API " +
			"silently discards them (verified against v8.0.4), so they cannot round-trip through Terraform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the field.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the field. Must be unique.",
				Required:            true,
			},
			"element": schema.StringAttribute{
				MarkdownDescription: "Form element of the field. One of `text`, `textarea`, `listbox`, " +
					"`checkbox` or `radio`.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("text", "textarea", "listbox", "checkbox", "radio"),
				},
			},
			"format": schema.StringAttribute{
				MarkdownDescription: "Validation format of the field. One of `ANY`, `ALPHA`, `ALPHA-DASH`, " +
					"`NUMERIC`, `ALPHA-NUMERIC`, `EMAIL`, `DATE`, `URL`, `IP`, `IPV4`, `IPV6`, `MAC` or " +
					"`BOOLEAN`. Defaults to `ANY`.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("ANY", "ALPHA", "ALPHA-DASH", "NUMERIC", "ALPHA-NUMERIC",
						"EMAIL", "DATE", "URL", "IP", "IPV4", "IPV6", "MAC", "BOOLEAN"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"field_values": schema.StringAttribute{
				MarkdownDescription: "Selectable values for `listbox`, `checkbox` and `radio` elements, " +
					"one per line.",
				Optional: true,
			},
			"help_text": schema.StringAttribute{
				MarkdownDescription: "Help text shown below the field. Not returned by the API, so " +
					"external changes are not detected.",
				Optional: true,
			},
			"show_in_email": schema.BoolAttribute{
				MarkdownDescription: "Include this field in checkout/checkin emails. Not returned by the " +
					"API, so external changes are not detected.",
				Optional: true,
			},
			"db_column_name": schema.StringAttribute{
				MarkdownDescription: "Generated database column name of the field. Changes when the " +
					"field is renamed.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{dbColumnNameModifier{}},
			},
		},
	}
}

func (r *FieldResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = customfieldsapi.New(c)
	}
}

func (m *FieldResourceModel) toBody() map[string]any {
	body := map[string]any{
		"name":    m.Name.ValueString(),
		"element": m.Element.ValueString(),
	}
	if !m.Format.IsNull() && !m.Format.IsUnknown() {
		body["format"] = m.Format.ValueString()
	}
	tfutil.BodyString(body, "field_values", m.FieldValues)
	tfutil.BodyString(body, "help_text", m.HelpText)
	tfutil.BodyOptBool(body, "show_in_email", m.ShowInEmail)
	return body
}

// fromAPI maps the GET response. help_text and show_in_email are not part of
// the response, so the configured values in the model are left untouched.
func (m *FieldResourceModel) fromAPI(api *customfieldsapi.Field) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Element = types.StringValue(api.Element)
	m.Format = types.StringValue(api.Format)
	m.FieldValues = tfutil.StateStringPtrKeep(api.FieldValues, m.FieldValues)
	m.DBColumnName = types.StringValue(api.DbColumnName)
}

func (r *FieldResource) read(ctx context.Context, id int64, data *FieldResourceModel) error {
	api, err := r.svc.GetField(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *FieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateField(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create field", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read field after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read field", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateField(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update field", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read field after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteField(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete field", err.Error())
	}
}

func (r *FieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
