data "snipeit_user_me" "me" {}

output "token_owner" {
  value = data.snipeit_user_me.me.username
}
