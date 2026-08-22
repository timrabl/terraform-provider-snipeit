package organization_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/timrabl/terraform-provider-snipeit/internal/acctest"
)

func TestAccCompanyResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-company")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "snipeit_company" "test" {
  name  = %q
  email = "it@example.com"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_company.test", "name", name),
					resource.TestCheckResourceAttr("snipeit_company.test", "email", "it@example.com"),
					resource.TestCheckResourceAttrSet("snipeit_company.test", "id"),
				),
			},
			{
				ResourceName:      "snipeit_company.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
resource "snipeit_company" "test" {
  name = %q
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_company.test", "name", name+"-renamed"),
				),
			},
		},
	})
}

func TestAccSupplierResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-supplier")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "snipeit_supplier" "test" {
  name    = %q
  city    = "Rosenheim"
  country = "DE"
  contact = "Max Mustermann"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_supplier.test", "name", name),
					resource.TestCheckResourceAttr("snipeit_supplier.test", "city", "Rosenheim"),
					resource.TestCheckResourceAttrSet("snipeit_supplier.test", "id"),
				),
			},
			{
				ResourceName:      "snipeit_supplier.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
resource "snipeit_supplier" "test" {
  name = %q
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_supplier.test", "name", name+"-renamed"),
				),
			},
		},
	})
}

func TestAccDepartmentResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-department")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "snipeit_company" "test" {
  name = "%[1]s-company"
}

resource "snipeit_department" "test" {
  name       = %[1]q
  company_id = snipeit_company.test.id
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_department.test", "name", name),
					resource.TestCheckResourceAttrPair(
						"snipeit_department.test", "company_id",
						"snipeit_company.test", "id",
					),
				),
			},
			{
				ResourceName:      "snipeit_department.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLocationResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-location")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "snipeit_location" "test" {
  name    = %q
  city    = "Rosenheim"
  country = "DE"
  zip     = "83022"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_location.test", "name", name),
					resource.TestCheckResourceAttr("snipeit_location.test", "city", "Rosenheim"),
					resource.TestCheckResourceAttrSet("snipeit_location.test", "id"),
				),
			},
			{
				ResourceName:      "snipeit_location.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
resource "snipeit_location" "test" {
  name = %q
  city = "München"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_location.test", "city", "München"),
					resource.TestCheckNoResourceAttr("snipeit_location.test", "zip"),
				),
			},
		},
	})
}
