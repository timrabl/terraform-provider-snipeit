# Snipe-IT version compatibility

The provider is built and verified against **Snipe-IT v8.0.4**. It is not
rebuilt per Snipe-IT version. This matrix records what the acceptance suite
actually does against a spread of real Snipe-IT versions, so users can see
where each resource holds up.

Method: each version was booted from `dev/docker-compose.yml` (only the app
image tag changed), seeded headless with an admin + API token, and the full
acceptance suite (35 acceptance tests) was run against it. `✅` = the test
passed end to end (create → read → update → import → destroy where
applicable); `❌` = it failed. Failures are explained below with the real
error, because that is what tells you where the API drifted.

Last run: 2026-08-23.

## Matrix

| Resource / data source | 6.0.14 | 6.4.2 | 7.1.16 | 8.0.4 | 8.4.1 | 8.7.2 |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| **organization** | | | | | | |
| `snipeit_company` | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_department` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_location` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_supplier` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| organization lookups (data sources) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **assets** | | | | | | |
| `snipeit_manufacturer` | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_category` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_status_label` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_model` (incl. explicit-zero eol) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_hardware` (incl. explicit empty notes) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_hardware_checkout` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_hardware` (data source) | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| entity-by-name lookups (data sources) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| hardware audit due/overdue (data sources) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **people** | | | | | | |
| `snipeit_user` | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ |
| `snipeit_group` | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ |
| user/group lookups (data sources) | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ |
| **licensing** | | | | | | |
| `snipeit_license` (+ data source) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_license_seat` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **inventory** | | | | | | |
| `snipeit_accessory` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_consumable` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_component` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_accessory_checkout` | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_component_checkout` | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| `snipeit_consumable_checkout` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **customfields** | | | | | | |
| `snipeit_fieldset` (+ data source) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_field` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_field` (data source) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_field_fieldset_association` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **operations** | | | | | | |
| `snipeit_maintenance` | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| activity reports (data source) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

## Per-version tally (35 acceptance tests)

| Version | Boots | Passing | Read |
|---|:---:|:---:|---|
| 6.0.14 | ✅ | 28 / 35 | oldest tested; several gaps |
| 6.4.2 | ✅ | 34 / 35 | one known gap (accessory checkout) |
| 7.1.16 | ✅ | **35 / 35** | fully tracked (manufacturer notes fixed) |
| **8.0.4** | ✅ | **35 / 35** | **build target — full support** |
| 8.4.1 | ✅ | 30 / 35 | forward drift starts |
| 8.7.2 | ✅ | **35 / 35** | current latest; fully tracked (see below) |

The provider was written against 8.0.x and now also tracks 8.7.x (the current
latest) and 7.1.x at full parity; the remaining drift is backward, mild until
6.0. Forward drift from 8.4+ was real (Snipe-IT's own API changed after 8.0)
but is now absorbed by version gating and tolerant-read mappers.

## Notable breakages, with the real error

### Backward (older than the build target)

- **`snipeit_manufacturer` — was failing on 6.0 / 6.4 / 7.1, now fixed on 6.4
  and 7.1.** Snipe-IT ≤ 7.x never echoes the manufacturer `notes` field back
  (the API transformer omits it), so it did not survive a re-read (fixed
  earlier with `StateStringPtrPreserve`) or a fresh import. The import step now
  ignores `notes` the same way `order_number` is ignored on the inventory
  resources, so 7.1.16 is fully clean and 6.4.2 is left with only the
  accessory-checkout gap.
- **`snipeit_company` — fails on 6.0 only.** Same inconsistent-result class on
  a field 6.0 serializes differently.
- **`snipeit_hardware` data source — fails on 6.0 only.** The lookup response
  shape differs on 6.0.
- **`snipeit_user` / `snipeit_group` — fail on 6.0.** `snipe-it API error
  (HTTP 500): Server Error` on group create: the groups/permissions API is
  immature in 6.0. Users fail as a consequence (they reference groups).
- **`snipeit_accessory_checkout` — still fails on 6.0 and 6.4.** On Snipe-IT
  < 7.0 the checkout body expects `assigned_to`, not `assigned_user` (which
  7.x+ requires); the provider sends `assigned_user`, so the server returns
  "That user is invalid." Fixing it needs a version gate (send `assigned_to`
  below 7.0), which would make 6.4.2 fully clean. Left open for now.

### Forward (newer than the build target)

All the forward breakages that were open on 8.7.2 are now tracked, and 8.7.2
runs the full suite 35/35. The history, and how each was resolved:

- **`snipeit_user` / `snipeit_group` — was failing on 8.4/8.7, now fixed.** The
  groups response shape changed in 8.4+ (permission values became numbers);
  the `FlexString` tolerant decode plus version gating (v0.3.0) resolved it.
- **`snipeit_maintenance` — was failing on 8.4/8.7, now fixed.** Maintenances
  gained required fields (`name`, and `maintenance_type_id` replacing the
  free-text type); the maintenance body is now version-gated (v0.3.0).
- **`snipeit_component_checkout` — was failing on 8.4/8.7, now fixed.** 8.7
  dropped the top-level `id` from the `/components/{id}/assets` rows and moved
  the asset id into a nested `name` object; a new `client.FlexRef` tolerant
  type reads the asset id from either shape (tolerant-read, no gating).
- **`snipeit_accessory` / `snipeit_consumable` / `snipeit_component` — was
  failing on 8.7, now fixed.** Two distinct 8.7 drifts: `order_number` stopped
  echoing back (handled earlier by `StateStringPtrPreserve`), and clearing
  `purchase_cost` / `purchase_date` on update is silently ignored by the
  server, which keeps echoing the old value. The clear-aware state mappers keep
  the field null once the user has cleared it (tolerant-read, no gating). Note:
  on 8.7 these two fields are effectively immutable after create — the update
  endpoint ignores changes to them, not only clears — so changing (rather than
  clearing) an already-set value is not honored on 8.7.
- **`snipeit_field` — was failing on 8.7, now fixed.** 8.7 validates the field
  `element` against its `format`: a `listbox` is only valid with a compatible
  format. Switching a text/IP field to a listbox must now also set a compatible
  format; the provider surfaces the server's validation error verbatim when it
  does not.
- **`snipeit_hardware_checkout` — was failing on 8.7, now fixed.** Not a
  checkout drift: 8.7 returns HTTP 500 on *concurrent* `POST /hardware`, a
  server-side race in asset creation. The checkout resource is unaffected; the
  acceptance test now serializes its two asset creates so it exercises the
  checkout rather than the unrelated concurrency bug.

Not re-measured this round: **8.4.1** and **6.0.14** are not part of the gating
version-matrix (which runs 6.4.2 / 7.1.16 / 8.0.4 / 8.7.2), so their columns
above predate the v0.3.0 gating work and the 8.7 fixes; several of their `❌`
cells are likely stale.

## How many versions are cleanly supported

- **Fully supported (35/35): 8.0.x** (the build target), **8.7.x** (current
  latest), and **7.1.x**.
- **Effectively supported: 6.4.x** with one documented gap (accessory
  checkout, which differs before 7.x).
- **Not in the gating matrix: 6.0.x and 8.4.x** — not re-measured this round;
  their rows above are historical.

Net: the provider cleanly covers **7.1 through 8.7** (the four gating-matrix
versions minus 6.4), and works well back to 6.4.

## Implication for the provider

The forward breakages were concentrated and API-driven (groups shape,
maintenance required fields, checkout pivots, `order_number` echo,
`purchase_cost` clear, field element/format validation). Server-version
detection plus capability gating (`internal/snipeitversion`) is now in place:
the client detects the Snipe-IT version at configure time, resources gate the
genuinely different requests (group decode, maintenance body) on it, and pure
response-shape drift is absorbed with tolerant-read state mappers and custom
types (`FlexString`, `FlexRef`, `StateStringPtrPreserve`, the clear-aware money
and date mappers) that need no gating. The version matrix gates 7.1.16, 8.0.4
and 8.7.2 as the versions the suite passes 35/35.
