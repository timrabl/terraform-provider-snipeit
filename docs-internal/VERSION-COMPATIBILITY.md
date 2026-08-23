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
| `snipeit_manufacturer` | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| `snipeit_category` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_status_label` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_model` (incl. explicit-zero eol) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_hardware` (incl. explicit empty notes) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_hardware_checkout` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_hardware` (data source) | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| entity-by-name lookups (data sources) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| hardware audit due/overdue (data sources) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **people** | | | | | | |
| `snipeit_user` | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| `snipeit_group` | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| user/group lookups (data sources) | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| **licensing** | | | | | | |
| `snipeit_license` (+ data source) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_license_seat` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **inventory** | | | | | | |
| `snipeit_accessory` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| `snipeit_consumable` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| `snipeit_component` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| `snipeit_accessory_checkout` | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_component_checkout` | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| `snipeit_consumable_checkout` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **customfields** | | | | | | |
| `snipeit_fieldset` (+ data source) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_field` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| `snipeit_field` (data source) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `snipeit_field_fieldset_association` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **operations** | | | | | | |
| `snipeit_maintenance` | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| activity reports (data source) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

## Per-version tally (35 acceptance tests)

| Version | Boots | Passing | Read |
|---|:---:|:---:|---|
| 6.0.14 | ✅ | 28 / 35 | oldest tested; several gaps |
| 6.4.2 | ✅ | 33 / 35 | good, two known gaps |
| 7.1.16 | ✅ | 34 / 35 | near-perfect |
| **8.0.4** | ✅ | **35 / 35** | **build target — full support** |
| 8.4.1 | ✅ | 30 / 35 | forward drift starts |
| 8.7.2 | ✅ | 26 / 35 | current latest; most drift |

The shape is the important part: the provider peaks at its build target and
degrades in **both** directions. Backward drift is mild until 6.0; forward
drift is real from 8.4 onward because Snipe-IT's own API changed after 8.0.

## Notable breakages, with the real error

### Backward (older than the build target)

- **`snipeit_manufacturer` — fails on all of 6.0 / 6.4 / 7.1.**
  `Provider produced inconsistent result after apply: .notes: was
  cty.StringVal("...") but now null`. Snipe-IT ≤ 7.x does not echo the
  manufacturer `notes` field back the way 8.x does, so the value the provider
  writes does not survive the re-read. This is the single most backward-visible
  break and the only failure on 7.1.16.
- **`snipeit_company` — fails on 6.0 only.** Same inconsistent-result class on
  a field 6.0 serializes differently.
- **`snipeit_hardware` data source — fails on 6.0 only.** The lookup response
  shape differs on 6.0.
- **`snipeit_user` / `snipeit_group` — fail on 6.0.** `snipe-it API error
  (HTTP 500): Server Error` on group create: the groups/permissions API is
  immature in 6.0. Users fail as a consequence (they reference groups).
- **`snipeit_accessory_checkout` — fails on 6.0 and 6.4.** `Unable to check
  out accessory`: the accessory checkout endpoint/pivot differs before 7.x.

### Forward (newer than the build target)

- **`snipeit_user` / `snipeit_group` — fail on 8.4 and 8.7.** `snipe-it:
  decoding GET /groups/{id} response: json: cannot unmarshal number into ...`.
  The **groups response shape changed in 8.4+**: a field the provider decodes
  as one type now comes back as a number. Group read-after-create fails, and
  users fail with it. This is the most consequential forward break — it is
  broken at *both* ends of the range for different reasons.
- **`snipeit_maintenance` — fails on 8.4 and 8.7.** 8.4: `name: The name field
  is required`; 8.7: `maintenance_type_id: The maintenance type id ...`.
  Maintenances gained required fields (a `name`, and `maintenance_type_id`
  replacing the old free-text type) that the provider does not send.
- **`snipeit_component_checkout` — fails on 8.4 and 8.7.** `The checkout
  succeeded but its pivot row could not be found`. The component checkout /
  assets response shape changed in 8.4+, so the provider cannot locate the
  pivot id it needs for checkin.
- **`snipeit_accessory` / `snipeit_consumable` / `snipeit_component` — fail on
  8.7 only.** `inconsistent result after apply: .order_number: was
  cty.StringVal("...") but now ...`. 8.7 changed how `order_number` is echoed.
  These pass on 8.4, so this drift landed between 8.4 and 8.7.
- **`snipeit_field` — fails on 8.7 only.** `Unable to update field`: the custom
  fields API changed again in 8.7.

## How many versions are cleanly supported

- **Fully supported: 8.0.x** (the build target, 35/35).
- **Effectively supported: 7.1.x** (34/35 — only the manufacturer `notes`
  round-trip), and **6.4.x** with two documented gaps (manufacturer,
  accessory checkout).
- **Usable with caveats: 6.0.x** (28/35) and **8.4.x** (30/35) — most
  resources work, but the listed families are broken.
- **Most drift: 8.7.x** (26/35), the current latest.

Net: the provider cleanly covers the **7.1 – 8.0** band it was written against,
works well back to 6.4, and needs attention to track Snipe-IT 8.4+.

## Implication for the provider (measurement only)

The forward breakages are concentrated and API-driven (groups shape,
maintenance required fields, checkout pivots, `order_number` echo). That is
exactly the situation where **server-version detection + capability gating**
pays off: detect the Snipe-IT version at provider configure time and adapt the
group decode, the maintenance body, and the checkout-pivot lookup per version,
rather than assuming the 8.0 shapes. This document is measurement only; no
gating is implemented here. It is the evidence base for that future decision.
