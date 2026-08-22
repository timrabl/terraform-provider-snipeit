resource "snipeit_category" "keyboards" {
  name          = "Keyboards"
  category_type = "accessory"
}

resource "snipeit_accessory" "mx_keys" {
  name         = "Logitech MX Keys"
  qty          = 20
  category_id  = snipeit_category.keyboards.id
  model_number = "920-009415"
  min_amt      = 2
}
