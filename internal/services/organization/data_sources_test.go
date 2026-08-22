// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package organization_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/timrabl/terraform-provider-snipeit/internal/acctest"
)

// TestAccOrganizationDataSources creates one of each organization entity and
// looks each up again by exact name through its data source.
func TestAccOrganizationDataSources(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-orgds")

	config := fmt.Sprintf(`
resource "snipeit_company" "test" {
  name = "%[1]s-company"
}

resource "snipeit_supplier" "test" {
  name = "%[1]s-supplier"
}

resource "snipeit_department" "test" {
  name       = "%[1]s-department"
  company_id = snipeit_company.test.id
}

resource "snipeit_location" "test" {
  name = "%[1]s-location"
  city = "Rosenheim"
}

data "snipeit_company" "test" {
  name       = snipeit_company.test.name
  depends_on = [snipeit_company.test]
}

data "snipeit_supplier" "test" {
  name       = snipeit_supplier.test.name
  depends_on = [snipeit_supplier.test]
}

data "snipeit_department" "test" {
  name       = snipeit_department.test.name
  depends_on = [snipeit_department.test]
}

data "snipeit_location" "test" {
  name       = snipeit_location.test.name
  depends_on = [snipeit_location.test]
}
`, prefix)

	pairs := [][2]string{
		{"data.snipeit_company.test", "snipeit_company.test"},
		{"data.snipeit_supplier.test", "snipeit_supplier.test"},
		{"data.snipeit_department.test", "snipeit_department.test"},
		{"data.snipeit_location.test", "snipeit_location.test"},
	}
	var checks []resource.TestCheckFunc
	for _, p := range pairs {
		checks = append(checks, resource.TestCheckResourceAttrPair(p[0], "id", p[1], "id"))
	}
	checks = append(checks,
		resource.TestCheckResourceAttr("data.snipeit_location.test", "city", "Rosenheim"),
		resource.TestCheckResourceAttrPair(
			"data.snipeit_department.test", "company_id",
			"snipeit_company.test", "id",
		),
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
