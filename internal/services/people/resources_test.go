// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package people_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/timrabl/terraform-provider-snipeit/internal/acctest"
)

func TestAccUserResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-user")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(`
resource "snipeit_group" "test" {
  name = "%[1]s-group"
}

resource "snipeit_user" "test" {
  username   = %[1]q
  first_name = "Acc"
  last_name  = "Test"
  password   = "Sup3rS3cret!%[1]s"
  email      = "%[1]s@example.com"
  jobtitle   = "Tester"
  activated  = true
  groups     = [snipeit_group.test.id]
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_user.test", "username", name),
					resource.TestCheckResourceAttr("snipeit_user.test", "first_name", "Acc"),
					resource.TestCheckResourceAttr("snipeit_user.test", "last_name", "Test"),
					resource.TestCheckResourceAttr("snipeit_user.test", "email", name+"@example.com"),
					resource.TestCheckResourceAttr("snipeit_user.test", "activated", "true"),
					resource.TestCheckResourceAttr("snipeit_user.test", "groups.#", "1"),
					resource.TestCheckResourceAttrSet("snipeit_user.test", "id"),
				),
			},
			// ImportState testing (password is write-only and cannot be verified)
			{
				ResourceName:            "snipeit_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Update and Read testing: rename, drop group membership and email
			{
				Config: fmt.Sprintf(`
resource "snipeit_group" "test" {
  name = "%[1]s-group"
}

resource "snipeit_user" "test" {
  username   = %[1]q
  first_name = "Acc"
  last_name  = "Renamed"
  password   = "Sup3rS3cret!%[1]s"
  jobtitle   = "Senior Tester"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_user.test", "last_name", "Renamed"),
					resource.TestCheckResourceAttr("snipeit_user.test", "jobtitle", "Senior Tester"),
					resource.TestCheckNoResourceAttr("snipeit_user.test", "email"),
					resource.TestCheckNoResourceAttr("snipeit_user.test", "groups"),
				),
			},
		},
	})
}

func TestAccGroupResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-group")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(`
resource "snipeit_group" "test" {
  name = %q
  permissions = {
    "assets.view"   = "1"
    "assets.create" = "0"
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_group.test", "name", name),
					resource.TestCheckResourceAttr("snipeit_group.test", "permissions.%", "2"),
					resource.TestCheckResourceAttr("snipeit_group.test", "permissions.assets.view", "1"),
					resource.TestCheckResourceAttrSet("snipeit_group.test", "id"),
				),
			},
			// ImportState testing (permissions intentionally start null after
			// import, see the schema description)
			{
				ResourceName:            "snipeit_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"permissions"},
			},
			// Update: rename and replace the permissions map (exercises the
			// v8.0.x update-500 workaround)
			{
				Config: fmt.Sprintf(`
resource "snipeit_group" "test" {
  name = %q
  permissions = {
    "assets.view"  = "1"
    "reports.view" = "1"
  }
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_group.test", "name", name+"-renamed"),
					resource.TestCheckResourceAttr("snipeit_group.test", "permissions.%", "2"),
					resource.TestCheckResourceAttr("snipeit_group.test", "permissions.reports.view", "1"),
				),
			},
		},
	})
}
