resource "snipeit_field" "mac_address" {
  name    = "MAC Address"
  element = "text"
  format  = "MAC"
}

resource "snipeit_fieldset" "laptop" {
  name = "Laptop Details"
}

resource "snipeit_field_fieldset_association" "laptop_mac" {
  field_id    = snipeit_field.mac_address.id
  fieldset_id = snipeit_fieldset.laptop.id
}
