resource "snipeit_group" "helpdesk" {
  name = "Helpdesk"
  permissions = {
    "assets.view"     = "1"
    "assets.checkin"  = "1"
    "assets.checkout" = "1"
    "users.view"      = "1"
  }
}
