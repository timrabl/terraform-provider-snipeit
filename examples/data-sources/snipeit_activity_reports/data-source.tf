data "snipeit_activity_reports" "recent_checkouts" {
  action_type = "checkout"
  limit       = 25
}
