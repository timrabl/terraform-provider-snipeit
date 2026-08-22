---
page_title: "Getting started: bootstrap a company"
subcategory: ""
---

# Getting started: bootstrap a company

A compact, realistic starting point: offices, departments, staff, a laptop
fleet checked out to its owners, and license seats. Model people and devices
as maps, then let `for_each` do the work.

```hcl
variable "initial_password" {
  type      = string
  sensitive = true
}

locals {
  staff = {
    "anna.admin" = { first = "Anna", last = "Admin", dept = "IT" }
    "ben.seller" = { first = "Ben", last = "Seller", dept = "Sales" }
  }
  laptops = {
    "LT-001" = "anna.admin"
    "LT-002" = "ben.seller"
  }
}

resource "snipeit_company" "co" {
  name = "Example GmbH"
}

resource "snipeit_department" "dept" {
  for_each   = toset(["IT", "Sales"])
  name       = each.key
  company_id = snipeit_company.co.id
}

resource "snipeit_manufacturer" "lenovo" {
  name = "Lenovo"
}

resource "snipeit_category" "laptops" {
  name          = "Laptops"
  category_type = "asset"
}

resource "snipeit_status_label" "in_use" {
  name = "In Use"
  type = "deployable"
}

resource "snipeit_model" "t14" {
  name            = "ThinkPad T14"
  category_id     = snipeit_category.laptops.id
  manufacturer_id = snipeit_manufacturer.lenovo.id
}

resource "snipeit_user" "staff" {
  for_each      = local.staff
  username      = each.key
  first_name    = each.value.first
  last_name     = each.value.last
  password      = var.initial_password
  company_id    = snipeit_company.co.id
  department_id = snipeit_department.dept[each.value.dept].id
}

resource "snipeit_hardware" "laptop" {
  for_each  = local.laptops
  asset_tag = each.key
  model_id  = snipeit_model.t14.id
  status_id = snipeit_status_label.in_use.id
}

resource "snipeit_hardware_checkout" "assign" {
  for_each         = local.laptops
  asset_id         = snipeit_hardware.laptop[each.key].id
  checkout_to_type = "user"
  assigned_id      = snipeit_user.staff[each.value].id
}
```

Onboarding a new employee with a laptop is then one entry in each map.
Offboarding is deleting them again: the checkout is checked in before the
user and the asset are removed.

Two things worth knowing up front:

- `password` is write-only. The API never returns it, the provider keeps the
  configured value in state, and after `terraform import` it is unset.
- Deleting a checked-out consumable cannot be undone server-side, and status
  labels of type `pending` make their assets unavailable for checkout.
