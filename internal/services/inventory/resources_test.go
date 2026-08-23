// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package inventory_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/timrabl/terraform-provider-snipeit/internal/acctest"
)

func TestAccAccessoryResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-accessory")

	base := fmt.Sprintf(`
resource "snipeit_category" "test" {
  name          = "%[1]s-cat"
  category_type = "accessory"
}

resource "snipeit_manufacturer" "test" {
  name = "%[1]s-mfg"
}

resource "snipeit_location" "test" {
  name = "%[1]s-loc"
}
`, prefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: base + fmt.Sprintf(`
resource "snipeit_accessory" "test" {
  name            = %[1]q
  qty             = 5
  category_id     = snipeit_category.test.id
  manufacturer_id = snipeit_manufacturer.test.id
  location_id     = snipeit_location.test.id
  model_number    = "ACC-100"
  order_number    = "ORD-1"
  purchase_cost   = "88.10"
  purchase_date   = "2026-01-10"
  min_amt         = 2
  notes           = "created by acceptance test"
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_accessory.test", "name", prefix),
					resource.TestCheckResourceAttr("snipeit_accessory.test", "qty", "5"),
					resource.TestCheckResourceAttr("snipeit_accessory.test", "model_number", "ACC-100"),
					resource.TestCheckResourceAttr("snipeit_accessory.test", "purchase_cost", "88.10"),
					resource.TestCheckResourceAttr("snipeit_accessory.test", "purchase_date", "2026-01-10"),
					resource.TestCheckResourceAttr("snipeit_accessory.test", "min_amt", "2"),
					resource.TestCheckResourceAttrPair(
						"snipeit_accessory.test", "manufacturer_id",
						"snipeit_manufacturer.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"snipeit_accessory.test", "location_id",
						"snipeit_location.test", "id",
					),
				),
			},
			{
				ResourceName:      "snipeit_accessory.test",
				ImportState:       true,
				ImportStateVerify: true,
				// order_number is not returned by the API on Snipe-IT 8.7+, and
				// purchase_cost/purchase_date use a clear-aware mapper that reads
				// null on a fresh import (no prior value), so none of these can be
				// verified by a fresh import.
				ImportStateVerifyIgnore: []string{"order_number", "purchase_cost", "purchase_date"},
			},
			{
				Config: base + fmt.Sprintf(`
resource "snipeit_accessory" "test" {
  name        = "%[1]s-renamed"
  qty         = 8
  category_id = snipeit_category.test.id
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_accessory.test", "name", prefix+"-renamed"),
					resource.TestCheckResourceAttr("snipeit_accessory.test", "qty", "8"),
					resource.TestCheckNoResourceAttr("snipeit_accessory.test", "model_number"),
					resource.TestCheckNoResourceAttr("snipeit_accessory.test", "manufacturer_id"),
				),
			},
			{
				Config: base + fmt.Sprintf(`
resource "snipeit_accessory" "test" {
  name        = "%[1]s-renamed"
  qty         = 8
  category_id = snipeit_category.test.id
}

data "snipeit_accessory" "by_name" {
  name = snipeit_accessory.test.name
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.snipeit_accessory.by_name", "id",
						"snipeit_accessory.test", "id",
					),
					resource.TestCheckResourceAttr("data.snipeit_accessory.by_name", "qty", "8"),
				),
			},
		},
	})
}

func TestAccConsumableResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-consumable")

	base := fmt.Sprintf(`
resource "snipeit_category" "test" {
  name          = "%[1]s-cat"
  category_type = "consumable"
}

resource "snipeit_manufacturer" "test" {
  name = "%[1]s-mfg"
}
`, prefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: base + fmt.Sprintf(`
resource "snipeit_consumable" "test" {
  name            = %[1]q
  qty             = 10
  category_id     = snipeit_category.test.id
  manufacturer_id = snipeit_manufacturer.test.id
  item_no         = "ITEM-9"
  model_number    = "CON-200"
  purchase_cost   = "7.50"
  order_number    = "ORD-2"
  purchase_date   = "2026-02-01"
  min_amt         = 3
  notes           = "created by acceptance test"
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_consumable.test", "name", prefix),
					resource.TestCheckResourceAttr("snipeit_consumable.test", "qty", "10"),
					resource.TestCheckResourceAttr("snipeit_consumable.test", "item_no", "ITEM-9"),
					resource.TestCheckResourceAttr("snipeit_consumable.test", "purchase_cost", "7.50"),
					resource.TestCheckResourceAttr("snipeit_consumable.test", "purchase_date", "2026-02-01"),
					resource.TestCheckResourceAttr("snipeit_consumable.test", "min_amt", "3"),
					resource.TestCheckResourceAttrPair(
						"snipeit_consumable.test", "manufacturer_id",
						"snipeit_manufacturer.test", "id",
					),
				),
			},
			{
				ResourceName:      "snipeit_consumable.test",
				ImportState:       true,
				ImportStateVerify: true,
				// order_number is not returned by the API on Snipe-IT 8.7+, and
				// purchase_cost/purchase_date use a clear-aware mapper that reads
				// null on a fresh import (no prior value), so none of these can be
				// verified by a fresh import.
				ImportStateVerifyIgnore: []string{"order_number", "purchase_cost", "purchase_date"},
			},
			{
				Config: base + fmt.Sprintf(`
resource "snipeit_consumable" "test" {
  name        = "%[1]s-renamed"
  qty         = 12
  category_id = snipeit_category.test.id
}

data "snipeit_consumable" "by_name" {
  name = snipeit_consumable.test.name
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_consumable.test", "name", prefix+"-renamed"),
					resource.TestCheckResourceAttr("snipeit_consumable.test", "qty", "12"),
					resource.TestCheckNoResourceAttr("snipeit_consumable.test", "item_no"),
					resource.TestCheckResourceAttrPair(
						"data.snipeit_consumable.by_name", "id",
						"snipeit_consumable.test", "id",
					),
				),
			},
		},
	})
}

func TestAccComponentResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-component")

	base := fmt.Sprintf(`
resource "snipeit_category" "test" {
  name          = "%[1]s-cat"
  category_type = "component"
}
`, prefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: base + fmt.Sprintf(`
resource "snipeit_component" "test" {
  name          = %[1]q
  qty           = 8
  category_id   = snipeit_category.test.id
  serial        = "SER-77"
  order_number  = "ORD-3"
  purchase_cost = "6543.21"
  purchase_date = "2026-03-01"
  min_amt       = 1
  notes         = "created by acceptance test"
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_component.test", "name", prefix),
					resource.TestCheckResourceAttr("snipeit_component.test", "qty", "8"),
					resource.TestCheckResourceAttr("snipeit_component.test", "serial", "SER-77"),
					resource.TestCheckResourceAttr("snipeit_component.test", "purchase_cost", "6543.21"),
					resource.TestCheckResourceAttr("snipeit_component.test", "purchase_date", "2026-03-01"),
					resource.TestCheckResourceAttr("snipeit_component.test", "min_amt", "1"),
				),
			},
			{
				ResourceName:      "snipeit_component.test",
				ImportState:       true,
				ImportStateVerify: true,
				// order_number is not returned by the API on Snipe-IT 8.7+, and
				// purchase_cost/purchase_date use a clear-aware mapper that reads
				// null on a fresh import (no prior value), so none of these can be
				// verified by a fresh import.
				ImportStateVerifyIgnore: []string{"order_number", "purchase_cost", "purchase_date"},
			},
			{
				Config: base + fmt.Sprintf(`
resource "snipeit_component" "test" {
  name        = "%[1]s-renamed"
  qty         = 6
  category_id = snipeit_category.test.id
}

data "snipeit_component" "by_name" {
  name = snipeit_component.test.name
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_component.test", "name", prefix+"-renamed"),
					resource.TestCheckResourceAttr("snipeit_component.test", "qty", "6"),
					resource.TestCheckNoResourceAttr("snipeit_component.test", "serial"),
					resource.TestCheckResourceAttrPair(
						"data.snipeit_component.by_name", "id",
						"snipeit_component.test", "id",
					),
				),
			},
		},
	})
}
