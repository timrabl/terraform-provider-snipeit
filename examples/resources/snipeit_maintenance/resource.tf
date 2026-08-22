resource "snipeit_maintenance" "annual_service" {
  asset_id         = snipeit_hardware.mbp.id
  supplier_id      = snipeit_supplier.repair_shop.id
  maintenance_type = "Maintenance"
  title            = "Annual service"
  start_date       = "2026-08-01"
  completion_date  = "2026-08-10"
  is_warranty      = false
  notes            = "Fan cleaning and thermal paste."
}
