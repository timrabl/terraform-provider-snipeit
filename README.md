# Terraform Provider for Snipe-IT

Even in 2026 there is no modern open source ITAM tool that has truly arrived
in the 21st century. So most companies that don't want to spend thousands of
euros a year end up on [Snipe-IT](https://snipeitapp.com/). Which is honestly
quite good (the UI/UX could use an overhaul, but I am drifting away). Anyway,
it has a quite usable API, see the [quirks](#snipe-it-api-quirks-this-provider-handles)
below. And with an API comes, you might have guessed it, the possibility to
IaC that thing.

Sadly there was no Terraform provider. So I built one :)

Built on the [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework),
covering the complete Snipe-IT REST API, and designed to be bridged to
[Pulumi](https://www.pulumi.com/) via
[pulumi-terraform-bridge](https://github.com/pulumi/pulumi-terraform-bridge)
later.

## Usage

```hcl
terraform {
  required_providers {
    snipeit = {
      source = "timrabl/snipeit"
    }
  }
}

provider "snipeit" {
  url   = "https://snipeit.example.com" # or SNIPEIT_URL
  token = var.snipeit_token             # or SNIPEIT_TOKEN
  # insecure = true                     # skip TLS verification (dev only)
}
```

See [`examples/`](examples/) and the generated [`docs/`](docs/) for the full
reference.

## Resources

24 resources across 7 domains. Import uses the numeric Snipe-IT id unless
noted otherwise.

| Domain | Resource | Manages | Import |
|---|---|---|:---:|
| organization | `snipeit_company` | Companies | ✅ |
| organization | `snipeit_department` | Departments | ✅ |
| organization | `snipeit_location` | Locations | ✅ |
| organization | `snipeit_supplier` | Suppliers | ✅ |
| assets | `snipeit_manufacturer` | Manufacturers | ✅ |
| assets | `snipeit_category` | Categories (asset, accessory, consumable, component, license) | ✅ |
| assets | `snipeit_status_label` | Status labels | ✅ |
| assets | `snipeit_model` | Asset models | ✅ |
| assets | `snipeit_hardware` | Assets | ✅ |
| people | `snipeit_user` | Users (write-only password) | ✅ |
| people | `snipeit_group` | Permission groups | ✅ |
| licensing | `snipeit_license` | Licenses | ✅ |
| licensing | `snipeit_license_seat` | Seat assignment to a user or asset | `license_id/seat_id` |
| inventory | `snipeit_accessory` | Accessories | ✅ |
| inventory | `snipeit_consumable` | Consumables | ✅ |
| inventory | `snipeit_component` | Components | ✅ |
| inventory | `snipeit_hardware_checkout` | Asset checkout to user, asset or location (delete = checkin) | ❌ |
| inventory | `snipeit_accessory_checkout` | Accessory checkout to a user (delete = checkin) | ❌ |
| inventory | `snipeit_component_checkout` | Component checkout to an asset (delete = checkin) | ❌ |
| inventory | `snipeit_consumable_checkout` | Consumable checkout to a user (irreversible, delete is state-only) | ❌ |
| customfields | `snipeit_fieldset` | Custom fieldsets | ✅ |
| customfields | `snipeit_field` | Custom fields | ✅ |
| customfields | `snipeit_field_fieldset_association` | Field / fieldset membership | `field_id/fieldset_id` |
| operations | `snipeit_maintenance` | Asset maintenances | ✅ |

## Data sources

| Data source | Looks up |
|---|---|
| `snipeit_hardware` | One asset by `id`, `asset_tag` or `serial` |
| `snipeit_company`, `snipeit_manufacturer`, `snipeit_category`, `snipeit_supplier`, `snipeit_department`, `snipeit_location`, `snipeit_status_label`, `snipeit_model`, `snipeit_user`, `snipeit_group`, `snipeit_license`, `snipeit_accessory`, `snipeit_consumable`, `snipeit_component`, `snipeit_fieldset`, `snipeit_field` | One object by `id` or exact `name` (users also by `username`/`email`) |
| `snipeit_user_me` | The user owning the API token |
| `snipeit_activity_reports` | Activity log entries, filterable |
| `snipeit_hardware_audit_due` / `snipeit_hardware_audit_overdue` | Assets due / overdue for audit |

### Deliberately not mapped

| Endpoint / attribute | Why not |
|---|---|
| `/settings/backups` (list/download) | Backup files are not declarative infrastructure |
| `/hardware/{id}/restore`, `/users/{id}/restore` | Soft-delete restore is an action, not a desired state |
| `POST /hardware/audit` | Audit submission is an event, not a resource |
| File attachment upload/download | Binary blobs don't round-trip through state sanely |

## Snipe-IT API quirks this provider handles

| Quirk | How the provider deals with it |
|---|---|
| Errors come back as **HTTP 200** with a `{status, messages, payload}` envelope | The client derives success from the envelope, never the HTTP status |
| Create responses carry only a partial payload | Every object is re-read after create/update |
| Rate limiting (HTTP 429, default 120 req/min) | Retried with progressive backoff. For heavy usage raise `API_THROTTLE_PER_MINUTE` on the instance |
| Decorated values (`"24 mo."`, `"36 months"`, `0/1` and `"0"/"1"` booleans) | Normalized by dedicated decoder types |
| Clearing a field on update needs `""` for text but explicit `null` for dates/references | Encoded once in the shared request-body helpers |
| Occupied license seats reject a target-type switch in one PATCH | Release first, then assign |
| Assorted per-entity oddities (write-only fields, renamed keys, broken sub-endpoints) | Documented in the corrected specs under [`api/`](api/) |

## Development

Requirements: Go >= 1.26, Terraform >= 1.5, [mise](https://mise.jdx.dev/),
Docker (any context, local or remote).

| Task | What it does |
|---|---|
| `mise run build` | Compile |
| `mise run test` | Unit tests (no network needed) |
| `mise run testacc` | Acceptance tests against the dev instance |
| `mise run lint` | golangci-lint |
| `mise run generate` | Regenerate the registry docs (tfplugindocs) |
| `mise run api:generate` | Regenerate API models from the corrected specs |
| `mise run dev:up` / `dev:down` / `dev:destroy` | Manage the dev instance |

### Dev instance

`dev/docker-compose.yml` runs a disposable Snipe-IT on whatever docker
context is active. All machine-specific values come from `env.local`
(gitignored), which mise loads for every task:

```sh
cp env.local.example env.local
echo "SNIPEIT_DEV_APP_KEY=base64:$(openssl rand -base64 32)" >> env.local
mise run dev:up
```

First-time setup: open the instance URL (default `http://localhost:8085`),
run the pre-flight wizard, create an admin account and generate an API token
(top-right user menu, *Manage API Keys*). Put it into `env.local` as
`SNIPE_IT_API_TOKEN=...`. From then on `mise run testacc` just works.

### Running the provider from source

Add a [dev override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers)
to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "timrabl/snipeit" = "<your GOBIN>"
  }
  direct {}
}
```

Then `go install .` and run `terraform plan/apply` (skip `terraform init`).

## Architecture

Spec-first, layered by domain. The full recipe lives in
[`docs-internal/ARCHITECTURE.md`](docs-internal/ARCHITECTURE.md).

```
main.go                     provider server entrypoint
api/<domain>.yaml           corrected OpenAPI 3.0 specs, the source of truth,
                            hand-curated from live-verified behavior (the
                            upstream spec is inaccurate, see api/README.md)
internal/client/            hand-written transport core: envelope handling,
                            429 retries, decorated-value decoder types
internal/api/<domain>/      types.gen.go (oapi-codegen, models only) plus a
                            thin typed Service per entity
internal/services/<domain>/ Terraform resources and data sources (mapping
                            only), unit mapping tests, acceptance tests
internal/provider/          slim provider shell registering the domains
internal/tfutil/            shared schema/body/state helpers
internal/acctest/           shared acceptance-test harness
examples/                   example configurations (also used by tfplugindocs)
docs/                       generated registry documentation, do not edit
dev/                        disposable dev instance (docker compose)
```

Domains: `organization`, `assets`, `people`, `licensing`, `inventory`,
`customfields`, `operations`.

Two invariants protect the verified wire behavior:

1. Request bodies are built with the `tfutil.Body*` helpers (empty string
   clears text, explicit `null` clears dates and references, booleans are
   omitted when unset) and never from generated structs.
2. Generated types cover **responses only**, with quirky fields mapped onto
   `internal/client` types (`Ref`, `Date`, `FlexInt`, `FlexBool`) via
   `x-go-type`.

## Testing

| Suite | Command | Needs |
|---|---|---|
| Unit | `mise run test` | Nothing. Transport tests run against `httptest` servers, mapping tests against recorded JSON fixtures |
| Acceptance | `mise run testacc` | The live dev instance and an API token in `env.local`. Full CRUD + import for every resource and data source |
