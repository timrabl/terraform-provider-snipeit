# Consumption is IRREVERSIBLE: the API has no consumable checkin, destroying
# this resource only removes the record from Terraform state.
resource "snipeit_consumable_checkout" "toner_to_bob" {
  consumable_id = snipeit_consumable.toner.id
  user_id       = 42
  note          = "printer room refill"
}
