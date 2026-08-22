# Provider architecture

```
api/<domain>.yaml            corrected OpenAPI 3.0.3 spec (source of truth, live-verified)
api/<domain>.gen.yaml        oapi-codegen config (models only)
internal/api/<domain>/       package <domain>api
  types.gen.go                 generated response models (DO NOT EDIT)
  service.go                   hand-written typed methods over the transport
internal/services/<domain>/  package <domain>
  *_resource.go                TF resources (schema + state mapping only)
  data_sources.go              TF data sources (tfutil lookup engine)
  package.go                   Resources() / DataSources() exports
  mapping_test.go              unit tests (package <domain>, fixtures, no network)
  *_test.go                    acceptance tests (package <domain>_test)
internal/provider/           slim provider: registers each services package
internal/client/             hand-written transport: envelope-on-200, 429 retry,
                             FlexInt/FlexBool/Date/DateTime/Ref (+ unit tests)
internal/tfutil/             shared helpers: body builders, state mappers,
                             ImportNumericID, generic lookup data source engine
internal/acctest/            shared acceptance harness (PreCheck, factories)
```

## Layer responsibilities

- **client** knows HTTP and Snipe-IT's transport quirks. Nothing else does.
- **api/<domain>** knows endpoint paths and response shapes. `types.gen.go` is
  generated from the corrected spec; `service.go` wraps the transport with
  typed Get/Create/Update/Delete/Search methods. Create returns only the new
  id (the API's create payload is partial), so callers re-Get.
- **services/<domain>** knows Terraform: schemas, plan/state handling, and the
  mapping between API models and TF models. No direct HTTP, no path strings.
- **provider** only registers domains and configures the shared client.

## The request-bodies-stay-maps rule

Request bodies are `map[string]any` built with the tfutil body builders, NOT
generated structs. Clear-on-unset semantics are per-field decisions the
generator cannot express:

| Builder                    | null config value means            | use for              |
|----------------------------|------------------------------------|----------------------|
| `tfutil.BodyString`        | send `""` (clears server-side)     | plain strings        |
| `tfutil.BodyNullableString`| send JSON `null`                   | dates ( "" invalid ) |
| `tfutil.BodyNullableInt`   | send JSON `null` (clears refs)     | *_id references      |
| `tfutil.BodyOptBool`       | omit the key                       | Optional+Computed bools |

Unknown values are always omitted (Optional+Computed attrs mid-plan).

## x-go-type mapping for quirky fields

Declare a schema component and map it onto the transport type; generated
models then embed the quirk-tolerant decoder:

| API behavior                                | x-go-type         |
|---------------------------------------------|-------------------|
| Nested `{id, name}` reference               | `client.Ref`      |
| `{datetime, formatted}` timestamp           | `client.DateTime` |
| `{date, formatted}` date                    | `client.Date`     |
| Int as `"24 mo."`/`"24"`/`24`               | `client.FlexInt`  |
| Bool as `0/1/"0"/"1"/true`                  | `client.FlexBool` |

See `api/organization.yaml` `components.schemas.NestedRef` for the exact
`x-go-type` + `x-go-type-import` incantation. Nullable strings are
`type: string, nullable: true` and generate as `*string`.

## Adding a domain (canonical steps, mirror `organization`)

1. Write `api/<domain>.yaml` from live-verified shapes (probe with curl, never
   trust the upstream spec) and `api/<domain>.gen.yaml`.
2. `mise run api:generate`, then check the generated types are what the
   mapping code needs (field pointerness distinguishes null, quirky types
   via x-go-type).
3. Write `internal/api/<domain>/service.go`: `New(c *client.Client) *Service`
   plus Get/Create/Update/Delete/Search per entity.
4. Write `internal/services/<domain>/`: resources + data sources (mapping
   only), `package.go` with `Resources()`/`DataSources()`.
5. Register in `internal/provider/provider.go`: import the package and add
   `rs = append(rs, <domain>.Resources()...)` in Resources plus
   `ds = append(ds, <domain>.DataSources()...)` in DataSources.
6. Tests:
   - `mapping_test.go` (package `<domain>`): fixture JSON → generated type →
     `fromAPI` assertions ("" and null → TF null, zero refs → null), plus
     `toBody` clear-semantics assertions.
   - `*_test.go` (package `<domain>_test`): acceptance tests using
     `internal/acctest` (`acc.PreCheck`, `acc.ProtoV6ProviderFactories`),
     random-prefixed names, create → import-verify → update steps.
7. Delete the migrated files from `internal/provider` and trim any shared
   test there that covered the moved entities.

## Unit vs acceptance tests

- `mise run test`: unit only, no TF_ACC, no network, runs everywhere.
  Covers transport quirks (`internal/client`) and mapping logic
  (`internal/services/*/mapping_test.go`).
- `mise run testacc`: acceptance, real Terraform CLI against the live dev
  instance (TF_ACC=1, SNIPEIT_URL/SNIPEIT_TOKEN). Covers CRUD, import,
  clear-on-update, and data source lookups end-to-end.

## Import-cycle rule

`internal/acctest` imports `internal/provider`, so tests **inside** package
`provider` cannot use it (cycle). Domain acceptance tests avoid this by being
external test packages (`package <domain>_test`).
