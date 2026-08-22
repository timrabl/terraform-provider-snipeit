# Check component units out to an asset; destroy checks the same qty back in.
resource "snipeit_component_checkout" "ram_upgrade" {
  component_id = snipeit_component.ddr5.id
  asset_id     = snipeit_hardware.workstation.id
  qty          = 2
}
