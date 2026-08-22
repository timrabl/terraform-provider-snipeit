// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package customfields

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	customfieldsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/customfields"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ datasource.DataSource = &FieldsetDataSource{}
	_ datasource.DataSource = &FieldDataSource{}
)

// NewFieldsetDataSource returns a new snipeit_fieldset data source.
func NewFieldsetDataSource() datasource.DataSource {
	return &FieldsetDataSource{}
}

// FieldsetDataSource looks up a fieldset by id or name, including its fields.
type FieldsetDataSource struct {
	svc *customfieldsapi.Service
}

// FieldsetDataSourceFieldModel is one field row inside the fieldset.
type FieldsetDataSourceFieldModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	DBColumnName types.String `tfsdk:"db_column_name"`
	Element      types.String `tfsdk:"element"`
	Format       types.String `tfsdk:"format"`
	Required     types.Bool   `tfsdk:"required"`
}

// FieldsetDataSourceModel describes the data source data model.
type FieldsetDataSourceModel struct {
	ID     types.Int64                    `tfsdk:"id"`
	Name   types.String                   `tfsdk:"name"`
	Fields []FieldsetDataSourceFieldModel `tfsdk:"fields"`
}

func (d *FieldsetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fieldset"
}

func (d *FieldsetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a fieldset by `id` or `name`, including the custom fields it contains.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the fieldset.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the fieldset.",
				Optional:            true,
				Computed:            true,
			},
			"fields": schema.ListNestedAttribute{
				MarkdownDescription: "Custom fields contained in the fieldset.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.Int64Attribute{MarkdownDescription: "Field id.", Computed: true},
						"name":           schema.StringAttribute{MarkdownDescription: "Field name.", Computed: true},
						"db_column_name": schema.StringAttribute{MarkdownDescription: "Database column name.", Computed: true},
						"element":        schema.StringAttribute{MarkdownDescription: "Form element type.", Computed: true},
						"format":         schema.StringAttribute{MarkdownDescription: "Validation format.", Computed: true},
						"required":       schema.BoolAttribute{MarkdownDescription: "Whether the field is required in this fieldset.", Computed: true},
					},
				},
			},
		},
	}
}

func (d *FieldsetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		d.svc = customfieldsapi.New(c)
	}
}

func (d *FieldsetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FieldsetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var api *customfieldsapi.Fieldset
	switch {
	case !data.ID.IsNull():
		got, err := d.svc.GetFieldset(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Unable to look up fieldset", err.Error())
			return
		}
		api = got
	case !data.Name.IsNull():
		list, err := d.svc.ListFieldsets(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list fieldsets", err.Error())
			return
		}
		for i := range list.Rows {
			if list.Rows[i].Name == data.Name.ValueString() {
				api = &list.Rows[i]
				break
			}
		}
		if api == nil {
			resp.Diagnostics.AddError("Unable to look up fieldset",
				fmt.Sprintf("No fieldset named %q found.", data.Name.ValueString()))
			return
		}
	default:
		resp.Diagnostics.AddError("Missing lookup key", "Either id or name must be set.")
		return
	}

	data.ID = types.Int64Value(api.Id)
	data.Name = types.StringValue(api.Name)
	data.Fields = make([]FieldsetDataSourceFieldModel, 0, len(api.Fields.Rows))
	for _, f := range api.Fields.Rows {
		data.Fields = append(data.Fields, FieldsetDataSourceFieldModel{
			ID:           types.Int64Value(f.Id),
			Name:         types.StringValue(f.Name),
			DBColumnName: types.StringValue(f.DbColumnName),
			Element:      types.StringValue(f.Element),
			Format:       types.StringValue(f.Format),
			Required:     types.BoolValue(bool(f.Required)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// NewFieldDataSource returns a new snipeit_field data source.
func NewFieldDataSource() datasource.DataSource {
	return &FieldDataSource{}
}

// FieldDataSource looks up a custom field by id or name.
type FieldDataSource struct {
	svc *customfieldsapi.Service
}

// FieldDataSourceModel describes the data source data model.
type FieldDataSourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	DBColumnName types.String `tfsdk:"db_column_name"`
	Element      types.String `tfsdk:"element"`
	Format       types.String `tfsdk:"format"`
	FieldValues  types.String `tfsdk:"field_values"`
}

func (d *FieldDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_field"
}

func (d *FieldDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a custom field by `id` or `name`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the field.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the field.",
				Optional:            true,
				Computed:            true,
			},
			"db_column_name": schema.StringAttribute{
				MarkdownDescription: "Generated database column name of the field.",
				Computed:            true,
			},
			"element": schema.StringAttribute{
				MarkdownDescription: "Form element type of the field.",
				Computed:            true,
			},
			"format": schema.StringAttribute{
				MarkdownDescription: "Validation format of the field.",
				Computed:            true,
			},
			"field_values": schema.StringAttribute{
				MarkdownDescription: "Selectable values, one per line.",
				Computed:            true,
			},
		},
	}
}

func (d *FieldDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		d.svc = customfieldsapi.New(c)
	}
}

func (d *FieldDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FieldDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var api *customfieldsapi.Field
	switch {
	case !data.ID.IsNull():
		got, err := d.svc.GetField(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Unable to look up field", err.Error())
			return
		}
		api = got
	case !data.Name.IsNull():
		list, err := d.svc.ListFields(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list fields", err.Error())
			return
		}
		for i := range list.Rows {
			if list.Rows[i].Name == data.Name.ValueString() {
				api = &list.Rows[i]
				break
			}
		}
		if api == nil {
			resp.Diagnostics.AddError("Unable to look up field",
				fmt.Sprintf("No field named %q found.", data.Name.ValueString()))
			return
		}
	default:
		resp.Diagnostics.AddError("Missing lookup key", "Either id or name must be set.")
		return
	}

	data.ID = types.Int64Value(api.Id)
	data.Name = types.StringValue(api.Name)
	data.DBColumnName = types.StringValue(api.DbColumnName)
	data.Element = types.StringValue(api.Element)
	data.Format = types.StringValue(api.Format)
	data.FieldValues = tfutil.StateStringPtr(api.FieldValues)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
