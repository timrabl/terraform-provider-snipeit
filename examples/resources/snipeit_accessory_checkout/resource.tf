# Check one unit of an accessory out to a user; destroy checks it back in.
resource "snipeit_accessory_checkout" "keyboard_to_alice" {
  accessory_id = snipeit_accessory.mx_keys.id
  user_id      = 42
  note         = "new hire kit"
}
