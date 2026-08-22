// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package assets

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	assetsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/assets"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

// NewManufacturerDataSource returns a new snipeit_manufacturer data source.
func NewManufacturerDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[assetsapi.Manufacturer, ManufacturerResourceModel]{
		TypeSuffix: "_manufacturer",
		Path:       "/manufacturers",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single manufacturer by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":                  tfutil.DSID("manufacturer"),
				"name":                tfutil.DSLookupString("Name of the manufacturer. Set it to look up by exact name."),
				"url":                 tfutil.DSString("Website of the manufacturer."),
				"support_url":         tfutil.DSString("Support website of the manufacturer."),
				"warranty_lookup_url": tfutil.DSString("Warranty lookup URL of the manufacturer."),
				"support_phone":       tfutil.DSString("Support phone number of the manufacturer."),
				"support_email":       tfutil.DSString("Support email address of the manufacturer."),
				"notes":               tfutil.DSString("Free-form notes."),
			},
		},
		IDOf: func(m *ManufacturerResourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[ManufacturerResourceModel]{{
			Attr:  "name",
			Get:   func(m *ManufacturerResourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(_ context.Context, api *assetsapi.Manufacturer, m *ManufacturerResourceModel) error {
			m.fromAPI(api)
			return nil
		},
	})
}

// NewCategoryDataSource returns a new snipeit_category data source.
func NewCategoryDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[assetsapi.Category, CategoryResourceModel]{
		TypeSuffix: "_category",
		Path:       "/categories",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single category by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":                 tfutil.DSID("category"),
				"name":               tfutil.DSLookupString("Name of the category. Set it to look up by exact name."),
				"category_type":      tfutil.DSString("Type of the category (`asset`, `accessory`, `consumable`, `component`, `license`)."),
				"eula_text":          tfutil.DSString("EULA text. Always null: the API does not return it."),
				"use_default_eula":   tfutil.DSBool("Whether the instance-wide default EULA is used."),
				"require_acceptance": tfutil.DSBool("Whether users must confirm acceptance of items in this category."),
				"checkin_email":      tfutil.DSBool("Whether checkin/checkout emails are sent."),
				"notes":              tfutil.DSString("Free-form notes."),
			},
		},
		IDOf: func(m *CategoryResourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[CategoryResourceModel]{{
			Attr:  "name",
			Get:   func(m *CategoryResourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(_ context.Context, api *assetsapi.Category, m *CategoryResourceModel) error {
			m.fromAPI(api)
			return nil
		},
	})
}

// NewStatusLabelDataSource returns a new snipeit_status_label data source.
func NewStatusLabelDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[assetsapi.StatusLabel, StatusLabelResourceModel]{
		TypeSuffix: "_status_label",
		Path:       "/statuslabels",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single status label by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":            tfutil.DSID("status label"),
				"name":          tfutil.DSLookupString("Name of the status label. Set it to look up by exact name."),
				"type":          tfutil.DSString("Meta type of the label (`deployable`, `pending`, `undeployable`, `archived`)."),
				"color":         tfutil.DSString("Hex chart color of the label."),
				"show_in_nav":   tfutil.DSBool("Whether assets with this label appear in the side navigation."),
				"default_label": tfutil.DSBool("Whether this label is pinned to the top of the status dropdown."),
				"notes":         tfutil.DSString("Free-form notes."),
			},
		},
		IDOf: func(m *StatusLabelResourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[StatusLabelResourceModel]{{
			Attr:  "name",
			Get:   func(m *StatusLabelResourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(_ context.Context, api *assetsapi.StatusLabel, m *StatusLabelResourceModel) error {
			m.fromAPI(api)
			return nil
		},
	})
}

// NewModelDataSource returns a new snipeit_model data source.
func NewModelDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[assetsapi.Model, ModelResourceModel]{
		TypeSuffix: "_model",
		Path:       "/models",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single asset model by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":              tfutil.DSID("model"),
				"name":            tfutil.DSLookupString("Name of the model. Set it to look up by exact name."),
				"model_number":    tfutil.DSString("Manufacturer model number."),
				"category_id":     tfutil.DSInt64("Id of the category of this model."),
				"manufacturer_id": tfutil.DSInt64("Id of the manufacturer of this model."),
				"fieldset_id":     tfutil.DSInt64("Id of the custom fieldset attached to this model."),
				"eol":             tfutil.DSInt64("End-of-life in months."),
				"min_amt":         tfutil.DSInt64("Minimum quantity before a low-stock alert triggers."),
				"requestable":     tfutil.DSBool("Whether assets of this model can be requested."),
				"notes":           tfutil.DSString("Free-form notes."),
			},
		},
		IDOf: func(m *ModelResourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[ModelResourceModel]{{
			Attr:  "name",
			Get:   func(m *ModelResourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(_ context.Context, api *assetsapi.Model, m *ModelResourceModel) error {
			m.fromAPI(api)
			return nil
		},
	})
}
