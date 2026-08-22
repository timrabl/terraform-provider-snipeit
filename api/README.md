# Corrected Snipe-IT API specs

These OpenAPI 3.0.3 documents are the **source of truth** for the generated
API model types under `internal/api/<domain>/types.gen.go`.

## Why we maintain our own specs

The upstream spec (`dev/snipeit-openapi.json`, published on
snipe-it.readme.io) is hand-maintained and **verifiably wrong** in ways that
would break a generated client: paths with literal `:id`, missing supplier
mutations, wrong verbs for maintenances/fieldsets, custom fields returned
under `type` instead of `element`, write-only fields not marked, decorated
scalar values (`"24 mo."`, `0/1`-as-string booleans), and the
errors-as-HTTP-200 envelope. Every schema in this directory is derived from
**live-verified** behavior against a real instance (v8.0.4), not from the
upstream document. Treat upstream as a reference only.

## Conventions

- One spec per domain: `<domain>.yaml` + oapi-codegen config
  `<domain>.gen.yaml` generating models-only into
  `internal/api/<domain>/types.gen.go` (package `<domain>api`).
- Only **response** (GET-detail and list) shapes are modeled as schemas.
  Request bodies stay hand-built `map[string]any` in the service layer
  because clear-on-unset semantics (empty string vs explicit null vs omit)
  are per-field decisions the generator cannot express.
- Quirky field types map onto the hand-written transport types via
  `x-go-type` / `x-go-type-import`:

  | API behavior                                   | Schema                | x-go-type         |
  |------------------------------------------------|-----------------------|-------------------|
  | Nested `{id, name}` reference                  | `$ref: NestedRef`     | `client.Ref`      |
  | `{datetime, formatted}` timestamp              | `$ref: DateTimeObj`   | `client.DateTime` |
  | `{date, formatted}` date                       | `$ref: DateObj`       | `client.Date`     |
  | Int that may arrive as `"24 mo."`/`"24"`/24    | `$ref: FlexInt`       | `client.FlexInt`  |
  | Bool that may arrive as `0/1/"0"/"1"/bool`     | `$ref: FlexBool`      | `client.FlexBool` |

- Nullable strings are `type: string, nullable: true` → generated as
  `*string` (empty string and null are both treated as "unset" by the
  provider's state mappers).
- The mutation envelope (`{status, messages, payload}` with HTTP 200 even on
  errors) is handled entirely by `internal/client`; it is documented in each
  spec's `info.description` but not modeled as response schemas.

## Regenerating

```sh
mise run api:generate      # all domains
# or a single domain:
go tool oapi-codegen -config api/<domain>.gen.yaml api/<domain>.yaml
```

Generated files carry a "DO NOT EDIT" header; change the spec, regenerate,
and adjust the hand-written `service.go` beside the generated types.

## Editing rules

1. Never copy schemas from the upstream spec unverified. Probe the live dev
   instance (`curl` with the token from `env.local`) and model what the API
   actually returns.
2. A field the API returns but the provider does not map may be omitted from
   the schema; specs document what we consume, not everything that exists.
3. When the API lies per-version, note it in the field's `description`.
