---
page_title: "Using the provider with Terragrunt"
subcategory: ""
---

# Using the provider with Terragrunt

Nothing provider-specific is required, but two patterns make Terragrunt
setups clean. Both are taken from a verified working stack.

## Generate the provider block once

In the root configuration, so no unit ever declares it by hand:

```hcl
# root.hcl
generate "provider" {
  path      = "provider.tf"
  if_exists = "overwrite"
  contents  = <<PROV
terraform {
  required_providers {
    snipeit = {
      source  = "timrabl/snipeit"
      version = "~> 0.1"
    }
  }
}

provider "snipeit" {}
PROV
}
```

Credentials come from the `SNIPEIT_URL` and `SNIPEIT_TOKEN` environment
variables, so no unit carries them either.

## Pass ids between units

Snipe-IT objects reference each other by numeric id, which maps naturally
onto Terragrunt dependencies. A foundation unit owns the shared objects and
exports their ids:

```hcl
# modules/foundation outputs
output "model_id" { value = snipeit_model.laptop.id }
output "status_id" { value = snipeit_status_label.ready.id }
```

A fleet unit consumes them:

```hcl
# live/fleet/terragrunt.hcl
dependency "foundation" {
  config_path = "../foundation"
}

inputs = {
  model_id  = dependency.foundation.outputs.model_id
  status_id = dependency.foundation.outputs.status_id
}
```

The ids are plain numbers, so `mock_outputs` for plan-time validation are
trivial (`model_id = 1`).
