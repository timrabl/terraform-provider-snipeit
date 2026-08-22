resource "snipeit_category" "printer_supplies" {
  name          = "Printer Supplies"
  category_type = "consumable"
}

resource "snipeit_consumable" "toner" {
  name        = "HP 26X Toner"
  qty         = 12
  category_id = snipeit_category.printer_supplies.id
  item_no     = "CF226X"
  min_amt     = 3
}
