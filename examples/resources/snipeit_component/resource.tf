resource "snipeit_category" "ram" {
  name          = "RAM"
  category_type = "component"
}

resource "snipeit_component" "ddr5" {
  name        = "Kingston 32GB DDR5"
  qty         = 16
  category_id = snipeit_category.ram.id
  min_amt     = 4
}
