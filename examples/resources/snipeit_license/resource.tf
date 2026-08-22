resource "snipeit_category" "software" {
  name          = "Software"
  category_type = "license"
}

resource "snipeit_license" "office" {
  name            = "Office Suite"
  seats           = 25
  category_id     = snipeit_category.software.id
  serial          = "XXXX-YYYY-ZZZZ"
  license_name    = "Example GmbH"
  license_email   = "licenses@example.com"
  purchase_date   = "2026-01-10"
  expiration_date = "2027-01-10"
  maintained      = true
}
