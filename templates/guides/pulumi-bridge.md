---
page_title: "Using the provider from Pulumi"
subcategory: ""
---

# Using the provider from Pulumi

There is no dedicated Pulumi package yet, but Pulumi bridges any registry
provider (the short name resolves through the OpenTofu registry, where this
provider is also listed):

```sh
pulumi package add terraform-provider timrabl/snipeit
```

One thing to know: Pulumi's built-in resource `id` is a string, while the
Snipe-IT reference attributes are numbers. The bridge therefore exposes each
resource's numeric id as an extra output property named after the resource
(`categoryId`, `manufacturerId`, `hardwareId`, ...). Use those for
cross-references:

```yaml
resources:
  cat:
    type: snipeit:Category
    properties:
      name: Laptops
      categoryType: asset
  model:
    type: snipeit:Model
    properties:
      name: ThinkPad T14
      categoryId: ${cat.categoryId} # the numeric id property, not ${cat.id}
```
