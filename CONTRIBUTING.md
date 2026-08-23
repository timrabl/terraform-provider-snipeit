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

## Release flow

- Feature PRs run the fast CI in [`ci.yml`](.github/workflows/ci.yml): format,
  lint, unit tests, security scans, doc drift. The version matrix does **not**
  run here.
- The full Snipe-IT version matrix
  ([`version-matrix.yml`](.github/workflows/version-matrix.yml)) boots one
  disposable instance per Snipe-IT version and runs the acceptance suite
  against each. It is the release gate on the release-please PR, and also runs
  weekly on a schedule (which opens a `matrix-drift` issue on new breakage) and
  on demand via `workflow_dispatch`. The gating versions (`gate: true`) must
  pass for the gate to go green; the others run for visibility but do not gate,
  because the provider has documented, expected drift on them (see
  [`docs-internal/VERSION-COMPATIBILITY.md`](docs-internal/VERSION-COMPATIBILITY.md)).
  Flip a version's gate to `true` once the provider supports it cleanly.
- **Dispatch discipline:** the release-please PR is opened by the bot using the
  default `GITHUB_TOKEN`, and GitHub does not trigger workflows from
  token-created events. So the matrix does **not** start on its own on that PR
  even though it is wired to `pull_request`. Before merging a release PR,
  dispatch the version matrix manually against the PR branch
  (Actions -> Version matrix -> Run workflow) and wait for the gate to go
  green. Do not merge on the strength of the fast CI alone.
- Releasing: once the dispatched matrix is green, merge the release-please PR.
  That tags the release, and
  [`release-please.yml`](.github/workflows/release-please.yml) runs goreleaser
  to publish it. No manual tagging or branching.

## Security issues

Not here. See [SECURITY.md](SECURITY.md).
