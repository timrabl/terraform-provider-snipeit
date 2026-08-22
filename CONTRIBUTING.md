# Contributing

Thanks for looking into this. Short version: mise runs everything, the
architecture doc explains where code goes, and conventional commits drive the
releases.

## Setup

Requirements: Go >= 1.26, Terraform >= 1.5, [mise](https://mise.jdx.dev/),
Docker.

```sh
git clone https://github.com/timrabl/terraform-provider-snipeit
cd terraform-provider-snipeit
mise trust
cp env.local.example env.local   # fill in, see the file's comments
mise run dev:up                  # disposable Snipe-IT instance
```

| Task | What it does |
|---|---|
| `mise run build` | Compile |
| `mise run test` | Unit tests, no network needed |
| `mise run testacc` | Acceptance tests against the dev instance |
| `mise run lint` / `fmt` / `license:fix` | Lint, format, license headers |
| `mise run api:generate` | Regenerate API models from the specs in `api/` |
| `mise run generate` | Regenerate the registry docs |

## Where code goes

Read [`docs-internal/ARCHITECTURE.md`](docs-internal/ARCHITECTURE.md) before
adding a resource. The short form: the corrected OpenAPI specs in `api/` are
the source of truth, generated types cover responses only, request bodies
stay map-based, and every entity lives in a domain package under
`internal/services/`.

One hard rule: never model the API from the upstream Snipe-IT spec. Probe
the live dev instance and encode what it actually returns. The spec files
document every known quirk.

## Pull requests

- Conventional commit titles (`fix(assets): ...`, `feat(licensing): ...`).
  The type decides the release: releases are cut automatically by
  release-please from these commits, so choose the type by consumer effect.
- One concern per PR. New resources come with acceptance tests, mapping unit
  tests, an example config, and regenerated docs (CI enforces the last one).
- `mise run test` and the acceptance tests for touched resources must pass.
  CI additionally runs gofmt, addlicense, golangci-lint, gosec, govulncheck,
  trivy and a generated-code drift check.

## Security issues

Not here. See [SECURITY.md](SECURITY.md).
