package customfields_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/timrabl/terraform-provider-snipeit/internal/acctest"
)

func TestAccFieldsetResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-fieldset")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "snipeit_fieldset" "test" {
  name = %q
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_fieldset.test", "name", name),
					resource.TestCheckResourceAttrSet("snipeit_fieldset.test", "id"),
				),
			},
			{
				ResourceName:      "snipeit_fieldset.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
resource "snipeit_fieldset" "test" {
  name = %q
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_fieldset.test", "name", name+"-renamed"),
				),
			},
		},
	})
}

func TestAccFieldResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-field")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "snipeit_field" "test" {
  name      = %q
  element   = "text"
  format    = "IP"
  help_text = "created by acceptance test"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_field.test", "name", name),
					resource.TestCheckResourceAttr("snipeit_field.test", "element", "text"),
					resource.TestCheckResourceAttr("snipeit_field.test", "format", "IP"),
					resource.TestCheckResourceAttrSet("snipeit_field.test", "db_column_name"),
				),
			},
			{
				ResourceName:      "snipeit_field.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Not returned by the API, so import cannot recover them.
				ImportStateVerifyIgnore: []string{"help_text", "show_in_email"},
			},
			{
				Config: fmt.Sprintf(`
resource "snipeit_field" "test" {
  name         = %q
  element      = "listbox"
  field_values = "one\ntwo\nthree"
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("snipeit_field.test", "name", name+"-renamed"),
					resource.TestCheckResourceAttr("snipeit_field.test", "element", "listbox"),
					resource.TestCheckResourceAttr("snipeit_field.test", "field_values", "one\ntwo\nthree"),
				),
			},
		},
	})
}

func TestAccFieldFieldsetAssociationResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-assoc")

	config := fmt.Sprintf(`
resource "snipeit_field" "test" {
  name    = "%[1]s-field"
  element = "text"
}

resource "snipeit_fieldset" "test" {
  name = "%[1]s-fieldset"
}

resource "snipeit_field_fieldset_association" "test" {
  field_id    = snipeit_field.test.id
  fieldset_id = snipeit_fieldset.test.id
}
`, prefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"snipeit_field_fieldset_association.test", "field_id",
						"snipeit_field.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"snipeit_field_fieldset_association.test", "fieldset_id",
						"snipeit_fieldset.test", "id",
					),
					resource.TestMatchResourceAttr(
						"snipeit_field_fieldset_association.test", "id",
						regexp.MustCompile(`^\d+:\d+$`),
					),
				),
			},
			{
				ResourceName:      "snipeit_field_fieldset_association.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFieldsetDataSource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-fsds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "snipeit_field" "test" {
  name    = "%[1]s-field"
  element = "text"
}

resource "snipeit_fieldset" "test" {
  name = "%[1]s-fieldset"
}

resource "snipeit_field_fieldset_association" "test" {
  field_id    = snipeit_field.test.id
  fieldset_id = snipeit_fieldset.test.id
}

data "snipeit_fieldset" "by_name" {
  name       = snipeit_fieldset.test.name
  depends_on = [snipeit_field_fieldset_association.test]
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.snipeit_fieldset.by_name", "id",
						"snipeit_fieldset.test", "id",
					),
					resource.TestCheckResourceAttr("data.snipeit_fieldset.by_name", "fields.#", "1"),
					resource.TestCheckResourceAttr("data.snipeit_fieldset.by_name", "fields.0.name", prefix+"-field"),
				),
			},
		},
	})
}

func TestAccFieldDataSource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-fdds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "snipeit_field" "test" {
  name    = "%[1]s-field"
  element = "text"
  format  = "MAC"
}

data "snipeit_field" "by_id" {
  id = snipeit_field.test.id
}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.snipeit_field.by_id", "name", prefix+"-field"),
					resource.TestCheckResourceAttr("data.snipeit_field.by_id", "format", "MAC"),
					resource.TestCheckResourceAttr("data.snipeit_field.by_id", "element", "text"),
				),
			},
		},
	})
}
