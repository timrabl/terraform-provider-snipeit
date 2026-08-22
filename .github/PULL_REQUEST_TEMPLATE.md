# What

<!-- One or two sentences. Link the issue if there is one. -->

## Type

- [ ] fix (patch)
- [ ] feat (minor)
- [ ] breaking change (major, describe the migration below)
- [ ] chore / docs / refactor / ci (no release)

## Checklist

- [ ] Conventional commit title (`type(scope): summary`)
- [ ] Unit tests pass (`mise run test`)
- [ ] Acceptance tests pass for touched resources (`mise run testacc`)
- [ ] Generated code is current (`mise run api:generate && mise run generate`, no diff)
- [ ] Spec updated under `api/` if API behavior knowledge changed
