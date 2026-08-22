# Assign one seat of a license to a user...
resource "snipeit_license_seat" "for_user" {
  license_id          = snipeit_license.office.id
  assigned_to_user_id = 42
}

# ...and another seat to an asset.
resource "snipeit_license_seat" "for_asset" {
  license_id           = snipeit_license.office.id
  assigned_to_asset_id = snipeit_hardware.mbp.id
}
