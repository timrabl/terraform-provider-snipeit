data "snipeit_fieldset" "laptop" {
  name = "Laptop Details"
}

output "laptop_field_names" {
  value = [for f in data.snipeit_fieldset.laptop.fields : f.name]
}
