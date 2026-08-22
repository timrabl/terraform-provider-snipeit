package people_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/timrabl/terraform-provider-snipeit/internal/acctest"
)

func TestAccUserAndGroupDataSources(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-uds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "snipeit_group" "test" {
  name = "%[1]s-group"
}

resource "snipeit_user" "test" {
  username   = %[1]q
  first_name = "Lookup"
  password   = "Sup3rS3cret!%[1]s"
  email      = "%[1]s@example.com"
  groups     = [snipeit_group.test.id]
}

data "snipeit_user" "by_username" {
  username   = snipeit_user.test.username
  depends_on = [snipeit_user.test]
}

data "snipeit_user" "by_email" {
  email      = snipeit_user.test.email
  depends_on = [snipeit_user.test]
}

data "snipeit_group" "test" {
  name       = snipeit_group.test.name
  depends_on = [snipeit_group.test]
}

data "snipeit_user_me" "me" {}
`, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.snipeit_user.by_username", "id", "snipeit_user.test", "id"),
					resource.TestCheckResourceAttrPair("data.snipeit_user.by_email", "id", "snipeit_user.test", "id"),
					resource.TestCheckResourceAttrPair("data.snipeit_group.test", "id", "snipeit_group.test", "id"),
					resource.TestCheckResourceAttr("data.snipeit_user.by_username", "groups.#", "1"),
					resource.TestCheckResourceAttrSet("data.snipeit_user_me.me", "id"),
					resource.TestCheckResourceAttrSet("data.snipeit_user_me.me", "username"),
				),
			},
		},
	})
}
