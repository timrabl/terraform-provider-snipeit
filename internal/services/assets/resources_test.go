// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package assets_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/timrabl/terraform-provider-snipeit/internal/acctest"
)

func TestAccManufacturerResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-mfg")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(`
resource "snipeit_manufacturer" "test" {
  name          = %q
  url           = "https://example.com"
  support_email = "support@example.com"
  notes         = "created by terraform acceptance tests"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_manufacturer.test", "name", name),
					resource.TestCheckResourceAttr("snipeit_manufacturer.test", "url", "https://example.com"),
					resource.TestCheckResourceAttr("snipeit_manufacturer.test", "support_email", "support@example.com"),
					resource.TestCheckResourceAttrSet("snipeit_manufacturer.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "snipeit_manufacturer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(`
resource "snipeit_manufacturer" "test" {
  name = %q
  url  = "https://example.org"
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_manufacturer.test", "name", name+"-renamed"),
					resource.TestCheckResourceAttr("snipeit_manufacturer.test", "url", "https://example.org"),
					resource.TestCheckNoResourceAttr("snipeit_manufacturer.test", "support_email"),
				),
			},
		},
	})
}

func TestAccCategoryResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-category")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "snipeit_category" "test" {
  name          = %q
  category_type = "asset"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_category.test", "name", name),
					resource.TestCheckResourceAttr("snipeit_category.test", "category_type", "asset"),
					resource.TestCheckResourceAttrSet("snipeit_category.test", "id"),
				),
			},
			{
				ResourceName:      "snipeit_category.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
resource "snipeit_category" "test" {
  name               = %q
  category_type      = "asset"
  require_acceptance = true
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_category.test", "name", name+"-renamed"),
					resource.TestCheckResourceAttr("snipeit_category.test", "require_acceptance", "true"),
				),
			},
		},
	})
}

func TestAccStatusLabelResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-status")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "snipeit_status_label" "test" {
  name = %q
  type = "deployable"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_status_label.test", "name", name),
					resource.TestCheckResourceAttr("snipeit_status_label.test", "type", "deployable"),
					resource.TestCheckResourceAttrSet("snipeit_status_label.test", "id"),
				),
			},
			{
				ResourceName:      "snipeit_status_label.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
resource "snipeit_status_label" "test" {
  name  = %q
  type  = "pending"
  color = "#ff0000"
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_status_label.test", "name", name+"-renamed"),
					resource.TestCheckResourceAttr("snipeit_status_label.test", "type", "pending"),
					resource.TestCheckResourceAttr("snipeit_status_label.test", "color", "#ff0000"),
				),
			},
		},
	})
}

func TestAccModelResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-model")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acc.HardwareBaseConfig(prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_model.test", "name", prefix+"-model"),
					resource.TestCheckResourceAttr("snipeit_model.test", "model_number", "TST-1000"),
					resource.TestCheckResourceAttr("snipeit_model.test", "eol", "36"),
					resource.TestCheckResourceAttrPair(
						"snipeit_model.test", "category_id",
						"snipeit_category.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"snipeit_model.test", "manufacturer_id",
						"snipeit_manufacturer.test", "id",
					),
				),
			},
			{
				ResourceName:      "snipeit_model.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccHardwareResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-hw")
	tag := prefix + "-0001"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acc.HardwareBaseConfig(prefix) + fmt.Sprintf(`
resource "snipeit_hardware" "test" {
  asset_tag     = %q
  model_id      = snipeit_model.test.id
  status_id     = snipeit_status_label.test.id
  name          = "Test Asset"
  serial        = "SN-%s"
  purchase_date = "2026-01-15"
}
`, tag, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_hardware.test", "asset_tag", tag),
					resource.TestCheckResourceAttr("snipeit_hardware.test", "name", "Test Asset"),
					resource.TestCheckResourceAttr("snipeit_hardware.test", "serial", "SN-"+prefix),
					resource.TestCheckResourceAttr("snipeit_hardware.test", "purchase_date", "2026-01-15"),
					resource.TestCheckResourceAttrPair(
						"snipeit_hardware.test", "model_id",
						"snipeit_model.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"snipeit_hardware.test", "status_id",
						"snipeit_status_label.test", "id",
					),
				),
			},
			{
				ResourceName:      "snipeit_hardware.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acc.HardwareBaseConfig(prefix) + fmt.Sprintf(`
resource "snipeit_hardware" "test" {
  asset_tag = %q
  model_id  = snipeit_model.test.id
  status_id = snipeit_status_label.test.id
  name      = "Test Asset Renamed"
  notes     = "updated by acceptance test"
}
`, tag),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_hardware.test", "name", "Test Asset Renamed"),
					resource.TestCheckResourceAttr("snipeit_hardware.test", "notes", "updated by acceptance test"),
					resource.TestCheckNoResourceAttr("snipeit_hardware.test", "serial"),
				),
			},
		},
	})
}

func TestAccHardwareDataSource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-hwds")
	tag := prefix + "-0001"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acc.HardwareBaseConfig(prefix) + fmt.Sprintf(`
resource "snipeit_hardware" "test" {
  asset_tag = %q
  model_id  = snipeit_model.test.id
  status_id = snipeit_status_label.test.id
  name      = "Data Source Asset"
}

data "snipeit_hardware" "by_tag" {
  asset_tag = snipeit_hardware.test.asset_tag
}
`, tag),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.snipeit_hardware.by_tag", "id",
						"snipeit_hardware.test", "id",
					),
					resource.TestCheckResourceAttr("data.snipeit_hardware.by_tag", "name", "Data Source Asset"),
				),
			},
		},
	})
}

// TestAccEntityByNameDataSources creates one of each assets-domain entity and
// looks each up again by exact name through its data source.
func TestAccEntityByNameDataSources(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-ds")

	config := acc.HardwareBaseConfig(prefix) + `
data "snipeit_manufacturer" "test" {
  name       = snipeit_manufacturer.test.name
  depends_on = [snipeit_manufacturer.test]
}

data "snipeit_category" "test" {
  name       = snipeit_category.test.name
  depends_on = [snipeit_category.test]
}

data "snipeit_status_label" "test" {
  name       = snipeit_status_label.test.name
  depends_on = [snipeit_status_label.test]
}

data "snipeit_model" "test" {
  name       = snipeit_model.test.name
  depends_on = [snipeit_model.test]
}
`

	pairs := [][2]string{
		{"data.snipeit_manufacturer.test", "snipeit_manufacturer.test"},
		{"data.snipeit_category.test", "snipeit_category.test"},
		{"data.snipeit_status_label.test", "snipeit_status_label.test"},
		{"data.snipeit_model.test", "snipeit_model.test"},
	}
	var checks []resource.TestCheckFunc
	for _, p := range pairs {
		checks = append(checks, resource.TestCheckResourceAttrPair(p[0], "id", p[1], "id"))
	}
	checks = append(checks,
		resource.TestCheckResourceAttr("data.snipeit_category.test", "category_type", "asset"),
		resource.TestCheckResourceAttr("data.snipeit_status_label.test", "type", "deployable"),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.ComposeAggregateTestCheckFunc(checks...),
			},
		},
	})
}
