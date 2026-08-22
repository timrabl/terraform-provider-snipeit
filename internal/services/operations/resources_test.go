// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package operations_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/timrabl/terraform-provider-snipeit/internal/acctest"
)

// maintenanceBaseConfig provisions the asset + supplier chain a maintenance
// needs, on top of the shared hardware dependency chain.
func maintenanceBaseConfig(prefix string) string {
	return acc.HardwareBaseConfig(prefix) + fmt.Sprintf(`
resource "snipeit_supplier" "test" {
  name = "%[1]s-supplier"
}

resource "snipeit_hardware" "test" {
  asset_tag = "%[1]s-0001"
  model_id  = snipeit_model.test.id
  status_id = snipeit_status_label.test.id
}
`, prefix)
}

func TestAccMaintenanceResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-mx")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: maintenanceBaseConfig(prefix) + `
resource "snipeit_maintenance" "test" {
  asset_id         = snipeit_hardware.test.id
  supplier_id      = snipeit_supplier.test.id
  maintenance_type = "Maintenance"
  title            = "Annual service"
  start_date       = "2026-08-01"
  cost             = "1500.50"
  completion_date  = "2026-08-10"
  notes            = "created by acceptance test"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_maintenance.test", "maintenance_type", "Maintenance"),
					resource.TestCheckResourceAttr("snipeit_maintenance.test", "title", "Annual service"),
					resource.TestCheckResourceAttr("snipeit_maintenance.test", "start_date", "2026-08-01"),
					resource.TestCheckResourceAttr("snipeit_maintenance.test", "cost", "1500.50"),
					resource.TestCheckResourceAttr("snipeit_maintenance.test", "completion_date", "2026-08-10"),
					resource.TestCheckResourceAttrPair(
						"snipeit_maintenance.test", "asset_id",
						"snipeit_hardware.test", "id",
					),
				),
			},
			{
				ResourceName:      "snipeit_maintenance.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: maintenanceBaseConfig(prefix) + `
resource "snipeit_maintenance" "test" {
  asset_id         = snipeit_hardware.test.id
  supplier_id      = snipeit_supplier.test.id
  maintenance_type = "Repair"
  title            = "Board swap"
  start_date       = "2026-08-05"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_maintenance.test", "maintenance_type", "Repair"),
					resource.TestCheckResourceAttr("snipeit_maintenance.test", "title", "Board swap"),
					resource.TestCheckNoResourceAttr("snipeit_maintenance.test", "completion_date"),
				),
			},
		},
	})
}

func TestAccActivityReportsDataSource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-act")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Creating the manufacturer guarantees at least one activity row.
				Config: fmt.Sprintf(`
resource "snipeit_manufacturer" "test" {
  name = "%s-mfg"
}

data "snipeit_activity_reports" "recent" {
  limit      = 5
  depends_on = [snipeit_manufacturer.test]
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.snipeit_activity_reports.recent", "total"),
					resource.TestCheckResourceAttrSet("data.snipeit_activity_reports.recent", "rows.0.action_type"),
				),
			},
		},
	})
}

func TestAccHardwareAuditDataSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Smoke test: the lists may legitimately be empty on a fresh
				// instance, so only assert that reads succeed.
				Config: `
data "snipeit_hardware_audit_due" "due" {}

data "snipeit_hardware_audit_overdue" "overdue" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.snipeit_hardware_audit_due.due", "total"),
					resource.TestCheckResourceAttrSet("data.snipeit_hardware_audit_overdue.overdue", "total"),
				),
			},
		},
	})
}
