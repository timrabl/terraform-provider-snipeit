package inventory_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/timrabl/terraform-provider-snipeit/internal/acctest"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
)

// testAccInvAPIClient builds a raw API client from the acceptance test env.
func testAccInvAPIClient(t *testing.T) *client.Client {
	t.Helper()
	c, err := client.New(client.Config{
		URL:   os.Getenv("SNIPEIT_URL"),
		Token: os.Getenv("SNIPEIT_TOKEN"),
	})
	if err != nil {
		t.Fatalf("building API client: %v", err)
	}
	return c
}

// testAccInvCreateUser creates a throwaway user directly via the API (the
// snipeit_user resource is another domain's concern) and registers its
// deletion as cleanup.
func testAccInvCreateUser(t *testing.T, prefix string) int64 {
	t.Helper()
	// Skip before touching the API client so plain `go test ./...` (no
	// TF_ACC) does not fail on missing environment.
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	c := testAccInvAPIClient(t)
	var created struct {
		ID int64 `json:"id"`
	}
	body := map[string]any{
		"first_name":            "Acc",
		"last_name":             "Test",
		"username":              prefix + "-user",
		"password":              "Str0ngPass!12345",
		"password_confirmation": "Str0ngPass!12345",
	}
	if err := c.Post(context.Background(), "/users", body, &created); err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Delete(context.Background(), fmt.Sprintf("/users/%d", created.ID))
	})
	return created.ID
}

func TestAccHardwareCheckoutResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-hwco")

	config := acc.HardwareBaseConfig(prefix) + fmt.Sprintf(`
resource "snipeit_hardware" "source" {
  asset_tag = "%[1]s-0001"
  model_id  = snipeit_model.test.id
  status_id = snipeit_status_label.test.id
}

resource "snipeit_hardware" "target" {
  asset_tag = "%[1]s-0002"
  model_id  = snipeit_model.test.id
  status_id = snipeit_status_label.test.id
}

resource "snipeit_hardware_checkout" "test" {
  asset_id         = snipeit_hardware.source.id
  checkout_to_type = "asset"
  assigned_id      = snipeit_hardware.target.id
  note             = "checked out by acceptance test"
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
						"snipeit_hardware_checkout.test", "asset_id",
						"snipeit_hardware.source", "id",
					),
					resource.TestCheckResourceAttrPair(
						"snipeit_hardware_checkout.test", "assigned_id",
						"snipeit_hardware.target", "id",
					),
					resource.TestCheckResourceAttr("snipeit_hardware_checkout.test", "checkout_to_type", "asset"),
				),
			},
			// Destroy checks the asset back in; the implicit destroy step of
			// the test framework verifies this does not error.
		},
	})
}

func TestAccAccessoryCheckoutResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-accco")
	userID := testAccInvCreateUser(t, prefix)

	config := fmt.Sprintf(`
resource "snipeit_category" "test" {
  name          = "%[1]s-cat"
  category_type = "accessory"
}

resource "snipeit_accessory" "test" {
  name        = %[1]q
  qty         = 3
  category_id = snipeit_category.test.id
}

resource "snipeit_accessory_checkout" "test" {
  accessory_id = snipeit_accessory.test.id
  user_id      = %[2]d
  note         = "checked out by acceptance test"
}
`, prefix, userID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("snipeit_accessory_checkout.test", "id"),
					resource.TestCheckResourceAttr("snipeit_accessory_checkout.test", "user_id", fmt.Sprint(userID)),
					resource.TestCheckResourceAttrPair(
						"snipeit_accessory_checkout.test", "accessory_id",
						"snipeit_accessory.test", "id",
					),
				),
			},
		},
	})
}

func TestAccComponentCheckoutResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-compco")

	config := acc.HardwareBaseConfig(prefix) + fmt.Sprintf(`
resource "snipeit_category" "comp" {
  name          = "%[1]s-comp-cat"
  category_type = "component"
}

resource "snipeit_component" "test" {
  name        = %[1]q
  qty         = 6
  category_id = snipeit_category.comp.id
}

resource "snipeit_hardware" "target" {
  asset_tag = "%[1]s-0001"
  model_id  = snipeit_model.test.id
  status_id = snipeit_status_label.test.id
}

resource "snipeit_component_checkout" "test" {
  component_id = snipeit_component.test.id
  asset_id     = snipeit_hardware.target.id
  qty          = 2
}
`, prefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("snipeit_component_checkout.test", "id"),
					resource.TestCheckResourceAttr("snipeit_component_checkout.test", "qty", "2"),
					resource.TestCheckResourceAttrPair(
						"snipeit_component_checkout.test", "asset_id",
						"snipeit_hardware.target", "id",
					),
				),
			},
		},
	})
}

func TestAccConsumableCheckoutResource(t *testing.T) {
	prefix := acctest.RandomWithPrefix("tf-acc-consco")
	userID := testAccInvCreateUser(t, prefix)

	config := fmt.Sprintf(`
resource "snipeit_category" "test" {
  name          = "%[1]s-cat"
  category_type = "consumable"
}

resource "snipeit_consumable" "test" {
  name        = %[1]q
  qty         = 4
  category_id = snipeit_category.test.id
}

resource "snipeit_consumable_checkout" "test" {
  consumable_id = snipeit_consumable.test.id
  user_id       = %[2]d
  note          = "consumed by acceptance test"
}
`, prefix, userID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t) },
		ProtoV6ProviderFactories: acc.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("snipeit_consumable_checkout.test", "id"),
					resource.TestCheckResourceAttr("snipeit_consumable_checkout.test", "user_id", fmt.Sprint(userID)),
				),
			},
		},
	})
}
