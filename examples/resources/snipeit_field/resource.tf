resource "snipeit_field" "mac_address" {
  name    = "MAC Address"
  element = "text"
  format  = "MAC"
}

resource "snipeit_field" "ram" {
  name         = "RAM"
  element      = "listbox"
  field_values = "8 GB\n16 GB\n32 GB\n64 GB"
}
