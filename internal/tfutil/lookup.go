// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package tfutil

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

// ListRow is the subset of fields any list endpoint row carries that we match
// lookups against.
type ListRow struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// LookupField declares one alternative lookup attribute of a data source,
// e.g. name, username or email.
type LookupField[M any] struct {
	// Attr is the schema attribute name, used in diagnostics.
	Attr string
	// Get extracts the configured value from the config model.
	Get func(*M) types.String
	// Match extracts the comparable value from a list row.
	Match func(ListRow) string
}

// LookupCfg describes a "find one object by id or unique field" data source.
type LookupCfg[A any, M any] struct {
	// TypeSuffix is appended to the provider type name, e.g. "_company".
	TypeSuffix string
	// Path is the API collection path, e.g. "/companies".
	Path string
	// Schema is the full data source schema.
	Schema dsschema.Schema
	// IDOf extracts the configured id from the config model.
	IDOf func(*M) types.Int64
	// Lookups are the alternative unique-field lookups.
	Lookups []LookupField[M]
	// FromAPI maps the fetched API object into the model.
	FromAPI func(context.Context, *A, *M) error
}

// NewLookupDataSource builds a generic data source that finds a single object
// either by numeric id or by exact match on a unique field via the list
// endpoint.
func NewLookupDataSource[A any, M any](cfg LookupCfg[A, M]) datasource.DataSource {
	return &lookupDataSource[A, M]{cfg: cfg}
}

type lookupDataSource[A any, M any] struct {
	client *client.Client
	cfg    LookupCfg[A, M]
}

var _ datasource.DataSource = &lookupDataSource[struct{}, struct{}]{}

func (d *lookupDataSource[A, M]) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + d.cfg.TypeSuffix
}

func (d *lookupDataSource[A, M]) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = d.cfg.Schema
}

func (d *lookupDataSource[A, M]) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = ClientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

// findByField pages through the list endpoint searching for value and returns
// the id of the exactly-one matching row.
func (d *lookupDataSource[A, M]) findByField(ctx context.Context, field LookupField[M], value string) (int64, error) {
	var list struct {
		Total int64     `json:"total"`
		Rows  []ListRow `json:"rows"`
	}
	path := fmt.Sprintf("%s?limit=500&search=%s", d.cfg.Path, url.QueryEscape(value))
	if err := d.client.Get(ctx, path, &list); err != nil {
		return 0, err
	}
	var matches []ListRow
	for _, row := range list.Rows {
		if field.Match(row) == value {
			matches = append(matches, row)
		}
	}
	switch len(matches) {
	case 0:
		return 0, fmt.Errorf("no object found with %s = %q", field.Attr, value)
	case 1:
		return matches[0].ID, nil
	default:
		return 0, fmt.Errorf("%d objects found with %s = %q, lookup must be unambiguous", len(matches), field.Attr, value)
	}
}

func (d *lookupDataSource[A, M]) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data M
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int64(0)
	if v := d.cfg.IDOf(&data); !v.IsNull() && !v.IsUnknown() {
		id = v.ValueInt64()
	} else {
		for _, field := range d.cfg.Lookups {
			v := field.Get(&data)
			if v.IsNull() || v.IsUnknown() {
				continue
			}
			found, err := d.findByField(ctx, field, v.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Unable to look up object", err.Error())
				return
			}
			id = found
			break
		}
	}
	if id == 0 {
		attrs := "id"
		for _, l := range d.cfg.Lookups {
			attrs += ", " + l.Attr
		}
		resp.Diagnostics.AddError("Missing lookup key",
			fmt.Sprintf("Exactly one of %s must be set.", attrs))
		return
	}

	var api A
	if err := d.client.Get(ctx, fmt.Sprintf("%s/%d", d.cfg.Path, id), &api); err != nil {
		resp.Diagnostics.AddError("Unable to read object", err.Error())
		return
	}
	if err := d.cfg.FromAPI(ctx, &api, &data); err != nil {
		resp.Diagnostics.AddError("Unable to map object", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- schema attribute shorthands for data sources ---------------------------

// DSID is the optional/computed id attribute of a lookup data source.
func DSID(objName string) dsschema.Int64Attribute {
	return dsschema.Int64Attribute{
		MarkdownDescription: "Numeric id of the " + objName + ". Set it to look up by id.",
		Optional:            true,
		Computed:            true,
	}
}

// DSLookupString is an optional/computed unique-field lookup attribute.
func DSLookupString(desc string) dsschema.StringAttribute {
	return dsschema.StringAttribute{MarkdownDescription: desc, Optional: true, Computed: true}
}

// DSString is a computed string attribute.
func DSString(desc string) dsschema.StringAttribute {
	return dsschema.StringAttribute{MarkdownDescription: desc, Computed: true}
}

// DSMoney is a computed money attribute (normalized decimal string).
func DSMoney(desc string) dsschema.StringAttribute {
	return dsschema.StringAttribute{MarkdownDescription: desc, Computed: true, CustomType: MoneyType{}}
}

// DSInt64 is a computed int64 attribute.
func DSInt64(desc string) dsschema.Int64Attribute {
	return dsschema.Int64Attribute{MarkdownDescription: desc, Computed: true}
}

// DSBool is a computed bool attribute.
func DSBool(desc string) dsschema.BoolAttribute {
	return dsschema.BoolAttribute{MarkdownDescription: desc, Computed: true}
}
