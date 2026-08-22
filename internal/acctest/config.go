package acctest

import "fmt"

// HardwareBaseConfig returns Terraform configuration provisioning the full
// dependency chain an asset needs: manufacturer -> category -> model plus a
// status label. Shared by every domain whose acceptance tests need an asset.
func HardwareBaseConfig(prefix string) string {
	return fmt.Sprintf(`
resource "snipeit_manufacturer" "test" {
  name = "%[1]s-mfg"
}

resource "snipeit_category" "test" {
  name          = "%[1]s-category"
  category_type = "asset"
}

resource "snipeit_model" "test" {
  name            = "%[1]s-model"
  model_number    = "TST-1000"
  category_id     = snipeit_category.test.id
  manufacturer_id = snipeit_manufacturer.test.id
  eol             = 36
}

resource "snipeit_status_label" "test" {
  name = "%[1]s-status"
  type = "deployable"
}
`, prefix)
}
