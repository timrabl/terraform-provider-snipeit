data "snipeit_hardware_audit_due" "due" {}

output "assets_due_for_audit" {
  value = [for a in data.snipeit_hardware_audit_due.due.assets : a.asset_tag]
}
