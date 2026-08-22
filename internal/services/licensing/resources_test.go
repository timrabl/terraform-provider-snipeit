// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package licensing_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/timrabl/terraform-provider-snipeit/internal/acctest"
)

func licenseBaseConfig(prefix string) string {
	return fmt.Sprintf(`
resource "snipeit_category" "lic" {
  name          = "%[1]s-cat"
  category_type = "license"
}
`, prefix)
}

func TestAccLicenseResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-lic")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: licenseBaseConfig(prefix) + fmt.Sprintf(`
resource "snipeit_license" "test" {
  name            = %q
  seats           = 3
  category_id     = snipeit_category.lic.id
  serial          = "AAAA-BBBB-CCCC"
  license_name    = "Example Licensee"
  license_email   = "licenses@example.com"
  purchase_date   = "2026-01-10"
  expiration_date = "2027-01-10"
  order_number    = "ORD-42"
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_license.test", "name", prefix),
					resource.TestCheckResourceAttr("snipeit_license.test", "seats", "3"),
					resource.TestCheckResourceAttr("snipeit_license.test", "serial", "AAAA-BBBB-CCCC"),
					resource.TestCheckResourceAttr("snipeit_license.test", "purchase_date", "2026-01-10"),
					resource.TestCheckResourceAttr("snipeit_license.test", "expiration_date", "2027-01-10"),
					resource.TestCheckResourceAttrPair(
						"snipeit_license.test", "category_id",
						"snipeit_category.lic", "id",
					),
					resource.TestCheckResourceAttrSet("snipeit_license.test", "id"),
				),
			},
			{
				ResourceName:      "snipeit_license.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: licenseBaseConfig(prefix) + fmt.Sprintf(`
resource "snipeit_license" "test" {
  name        = %q
  seats       = 5
  category_id = snipeit_category.lic.id
  maintained  = true
}
`, prefix+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_license.test", "name", prefix+"-renamed"),
					resource.TestCheckResourceAttr("snipeit_license.test", "seats", "5"),
					resource.TestCheckResourceAttr("snipeit_license.test", "maintained", "true"),
					resource.TestCheckNoResourceAttr("snipeit_license.test", "purchase_date"),
					resource.TestCheckNoResourceAttr("snipeit_license.test", "serial"),
				),
			},
		},
	})
}

func TestAccLicenseSeatResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-seat")

	seatBase := acc.HardwareBaseConfig(prefix) + licenseBaseConfig(prefix) + fmt.Sprintf(`
resource "snipeit_license" "test" {
  name        = %q
  seats       = 2
  category_id = snipeit_category.lic.id
}

resource "snipeit_hardware" "seat_target" {
  asset_tag = "%s-hw"
  model_id  = snipeit_model.test.id
  status_id = snipeit_status_label.test.id
}
`, prefix, prefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Assign a seat to an asset.
			{
				Config: seatBase + `
resource "snipeit_license_seat" "test" {
  license_id           = snipeit_license.test.id
  assigned_to_asset_id = snipeit_hardware.seat_target.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"snipeit_license_seat.test", "assigned_to_asset_id",
						"snipeit_hardware.seat_target", "id",
					),
					resource.TestCheckResourceAttrSet("snipeit_license_seat.test", "seat_id"),
					resource.TestCheckNoResourceAttr("snipeit_license_seat.test", "assigned_to_user_id"),
				),
			},
			// Import via "license_id/seat_id".
			{
				ResourceName: "snipeit_license_seat.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["snipeit_license_seat.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					return rs.Primary.Attributes["license_id"] + "/" + rs.Primary.Attributes["seat_id"], nil
				},
				ImportStateVerify: true,
			},
			// Reassign the same seat to a user (admin id 1 always exists).
			{
				Config: seatBase + `
resource "snipeit_license_seat" "test" {
  license_id          = snipeit_license.test.id
  assigned_to_user_id = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_license_seat.test", "assigned_to_user_id", "1"),
					resource.TestCheckNoResourceAttr("snipeit_license_seat.test", "assigned_to_asset_id"),
				),
			},
		},
	})
}

func TestAccLicenseDataSource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-licds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: licenseBaseConfig(prefix) + fmt.Sprintf(`
resource "snipeit_license" "test" {
  name        = %q
  seats       = 2
  category_id = snipeit_category.lic.id
  serial      = "DS-1111"
}

data "snipeit_license" "by_name" {
  name = snipeit_license.test.name
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.snipeit_license.by_name", "id",
						"snipeit_license.test", "id",
					),
					resource.TestCheckResourceAttr("data.snipeit_license.by_name", "seats", "2"),
					resource.TestCheckResourceAttr("data.snipeit_license.by_name", "free_seats_count", "2"),
					resource.TestCheckResourceAttr("data.snipeit_license.by_name", "serial", "DS-1111"),
				),
			},
		},
	})
}
