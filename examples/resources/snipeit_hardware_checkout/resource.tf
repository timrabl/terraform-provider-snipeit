# Check an asset out to a user; destroying the resource checks it back in.
resource "snipeit_hardware_checkout" "laptop_to_alice" {
  asset_id         = snipeit_hardware.mbp.id
  checkout_to_type = "user"
  assigned_id      = 42 # user id
  note             = "onboarding equipment"
}

# Assets can also be checked out to another asset or a location.
resource "snipeit_hardware_checkout" "dock_to_desk" {
  asset_id         = snipeit_hardware.dock.id
  checkout_to_type = "location"
  assigned_id      = snipeit_location.office.id
}
