// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	inventoryapi "github.com/timrabl/terraform-provider-snipeit/internal/api/inventory"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

// AccessoryDataSourceModel describes the snipeit_accessory data source model.
type AccessoryDataSourceModel struct {
	ID           types.Int64       `tfsdk:"id"`
	Name         types.String      `tfsdk:"name"`
	CategoryID   types.Int64       `tfsdk:"category_id"`
	Qty          types.Int64       `tfsdk:"qty"`
	ModelNumber  types.String      `tfsdk:"model_number"`
	Notes        types.String      `tfsdk:"notes"`
	PurchaseCost tfutil.MoneyValue `tfsdk:"purchase_cost"`
}

// NewAccessoryDataSource returns a new snipeit_accessory data source.
func NewAccessoryDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[inventoryapi.Accessory, AccessoryDataSourceModel]{
		TypeSuffix: "_accessory",
		Path:       "/accessories",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single accessory by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":            tfutil.DSID("accessory"),
				"name":          tfutil.DSLookupString("Exact name of the accessory. Set it to look up by name."),
				"category_id":   tfutil.DSInt64("Id of the category."),
				"qty":           tfutil.DSInt64("Total quantity."),
				"model_number":  tfutil.DSString("Manufacturer model number."),
				"notes":         tfutil.DSString("Free-form notes."),
				"purchase_cost": tfutil.DSMoney("Purchase cost as a plain decimal string."),
			},
		},
		IDOf: func(m *AccessoryDataSourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[AccessoryDataSourceModel]{{
			Attr:  "name",
			Get:   func(m *AccessoryDataSourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(_ context.Context, api *inventoryapi.Accessory, m *AccessoryDataSourceModel) error {
			m.ID = types.Int64Value(api.Id)
			m.Name = types.StringValue(api.Name)
			m.CategoryID = types.Int64Value(api.Category.IDOrZero())
			m.Qty = types.Int64Value(int64(api.Qty))
			m.ModelNumber = tfutil.StateStringPtr(api.ModelNumber)
			m.Notes = tfutil.StateStringPtr(api.Notes)
			m.PurchaseCost = tfutil.StateMoneyPtr(api.PurchaseCost)
			return nil
		},
	})
}

// ConsumableDataSourceModel describes the snipeit_consumable data source model.
type ConsumableDataSourceModel struct {
	ID           types.Int64       `tfsdk:"id"`
	Name         types.String      `tfsdk:"name"`
	CategoryID   types.Int64       `tfsdk:"category_id"`
	Qty          types.Int64       `tfsdk:"qty"`
	ItemNo       types.String      `tfsdk:"item_no"`
	Notes        types.String      `tfsdk:"notes"`
	PurchaseCost tfutil.MoneyValue `tfsdk:"purchase_cost"`
}

// NewConsumableDataSource returns a new snipeit_consumable data source.
func NewConsumableDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[inventoryapi.Consumable, ConsumableDataSourceModel]{
		TypeSuffix: "_consumable",
		Path:       "/consumables",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single consumable by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":            tfutil.DSID("consumable"),
				"name":          tfutil.DSLookupString("Exact name of the consumable. Set it to look up by name."),
				"category_id":   tfutil.DSInt64("Id of the category."),
				"qty":           tfutil.DSInt64("Total quantity."),
				"item_no":       tfutil.DSString("Item number."),
				"notes":         tfutil.DSString("Free-form notes."),
				"purchase_cost": tfutil.DSMoney("Purchase cost as a plain decimal string."),
			},
		},
		IDOf: func(m *ConsumableDataSourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[ConsumableDataSourceModel]{{
			Attr:  "name",
			Get:   func(m *ConsumableDataSourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(_ context.Context, api *inventoryapi.Consumable, m *ConsumableDataSourceModel) error {
			m.ID = types.Int64Value(api.Id)
			m.Name = types.StringValue(api.Name)
			m.CategoryID = types.Int64Value(api.Category.IDOrZero())
			m.Qty = types.Int64Value(int64(api.Qty))
			m.ItemNo = tfutil.StateStringPtr(api.ItemNo)
			m.Notes = tfutil.StateStringPtr(api.Notes)
			m.PurchaseCost = tfutil.StateMoneyPtr(api.PurchaseCost)
			return nil
		},
	})
}

// ComponentDataSourceModel describes the snipeit_component data source model.
type ComponentDataSourceModel struct {
	ID           types.Int64       `tfsdk:"id"`
	Name         types.String      `tfsdk:"name"`
	CategoryID   types.Int64       `tfsdk:"category_id"`
	Qty          types.Int64       `tfsdk:"qty"`
	Serial       types.String      `tfsdk:"serial"`
	Notes        types.String      `tfsdk:"notes"`
	PurchaseCost tfutil.MoneyValue `tfsdk:"purchase_cost"`
}

// NewComponentDataSource returns a new snipeit_component data source.
func NewComponentDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[inventoryapi.Component, ComponentDataSourceModel]{
		TypeSuffix: "_component",
		Path:       "/components",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single component by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":            tfutil.DSID("component"),
				"name":          tfutil.DSLookupString("Exact name of the component. Set it to look up by name."),
				"category_id":   tfutil.DSInt64("Id of the category."),
				"qty":           tfutil.DSInt64("Total quantity."),
				"serial":        tfutil.DSString("Serial number."),
				"notes":         tfutil.DSString("Free-form notes."),
				"purchase_cost": tfutil.DSMoney("Purchase cost as a plain decimal string."),
			},
		},
		IDOf: func(m *ComponentDataSourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[ComponentDataSourceModel]{{
			Attr:  "name",
			Get:   func(m *ComponentDataSourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(_ context.Context, api *inventoryapi.Component, m *ComponentDataSourceModel) error {
			m.ID = types.Int64Value(api.Id)
			m.Name = types.StringValue(api.Name)
			m.CategoryID = types.Int64Value(api.Category.IDOrZero())
			m.Qty = types.Int64Value(int64(api.Qty))
			m.Serial = tfutil.StateStringPtr(api.Serial)
			m.Notes = tfutil.StateStringPtr(api.Notes)
			m.PurchaseCost = tfutil.StateMoneyPtr(api.PurchaseCost)
			return nil
		},
	})
}
