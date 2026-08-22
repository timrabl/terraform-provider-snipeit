# Security policy

## Reporting a vulnerability

Please report vulnerabilities privately via
[GitHub private vulnerability reporting](https://github.com/timrabl/terraform-provider-snipeit/security/advisories/new).
Do not open a public issue for security problems.

You can expect an initial response within a week. Please include a
reproduction and the affected version.

## Scope

This provider talks to a Snipe-IT instance with an API token. Relevant
classes of problems include:

- Token leakage: the token must never appear in state output, logs, or error
  messages (it is marked sensitive in the provider schema).
- TLS: certificate verification is on by default. The `insecure` flag is an
  explicit opt-out intended for development instances only.
- Anything where crafted API responses could corrupt state or execute
  unintended requests.

Vulnerabilities in Snipe-IT itself belong to the
[Snipe-IT project](https://github.com/grokability/snipe-it/security/policy).

## Supported versions

Only the latest release receives security fixes.

## Release integrity

Releases are built by CI from signed commits and the checksum file is signed
with the project's GPG release key (RSA-4096, fingerprint
`7F4C 2D74 9B87 07B8 761D 54DA AA69 D1ED 81B6 4D48`). Terraform and OpenTofu
verify this signature automatically when installing from the registries.
