resource "snipeit_category" "laptops" {
  name          = "Laptops"
  category_type = "asset"
}

resource "snipeit_manufacturer" "apple" {
  name = "Apple"
}

resource "snipeit_model" "mbp14" {
  name            = "MacBook Pro 14\""
  model_number    = "A2779"
  category_id     = snipeit_category.laptops.id
  manufacturer_id = snipeit_manufacturer.apple.id
  eol             = 48
}

resource "snipeit_status_label" "ready" {
  name = "Ready to Deploy"
  type = "deployable"
}

resource "snipeit_hardware" "mbp" {
  asset_tag     = "IT-0001"
  model_id      = snipeit_model.mbp14.id
  status_id     = snipeit_status_label.ready.id
  serial        = "C02XXXXXXX"
  purchase_date = "2026-01-15"
}
