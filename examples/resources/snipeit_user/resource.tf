resource "snipeit_group" "helpdesk" {
  name = "Helpdesk"
}

resource "snipeit_user" "jdoe" {
  username   = "jdoe"
  first_name = "Jane"
  last_name  = "Doe"
  password   = var.initial_password
  email      = "jane.doe@example.com"
  jobtitle   = "IT Support"
  activated  = true
  groups     = [snipeit_group.helpdesk.id]
}
