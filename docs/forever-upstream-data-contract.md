# Forever upstream data and validation contract

Date: 2026-10-25
Approved: 2026-10-25
Status: Phase 1 design artifact. It binds the manual-editor and hosted-armory implementation
plans; it does not authorize implementation, and it is not legal advice.

Related: [armory specification](superpowers/specs/2026-10-01-forever-armory-design.md),
[WoWSims Forever review](forever-armory-wowsims-review.md),
[delivery roadmap](plans/2026-10-20-forever-armory-roadmap.md),
[earlier data/import research](forever-data-import-research.md).

## 1. Purpose

This contract decides four things the specification leaves open:

1. Which upstream payload is the Forever catalog, pinned to an immutable identity.
2. How each identifier domain is keyed, so a saved build never depends on array position.
3. Which validation rules our editor enforces, and which it cannot enforce from this payload.
4. How imports are mapped, validated and rejected without partial application.

It does not decide storage schema, authentication, publication transport, UI layout, or
deployment. Those belong to the Phase 2 and Phase 3 plans.

The contract is written to survive a change of dataset. Section 3 records the reuse decision;
section 4 defines our own domain types. An adapter maps the pinned upstream payload into them,
so replacing the dataset does not change the domain or the UI.

## 2. Pinned source

| Item | Value |
| --- | --- |
| Repository | `https://github.com/wowsims/forever` |
| Commit | `0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4` |
| Licence | MIT (software and associated documentation) |
| Main catalog | `assets/database/db.json`, `sha256 9fa8335e7e97f824373ad8225b6c2d82e1a2915b0e71a6f05e2312add442e9e3` |
| Leftover catalog | `assets/database/leftover_db.json`, `sha256 eeb3ff7788226b9739e470c27d8a9a4035dbae5b5ad5f5d781483eeef0ce4685` |
| Talent structure | `ui/sim/talents/trees/*.json`, `ui/sim/talents/config.ts` |
| Schema | `proto/ui.proto`, `proto/db.proto`, `proto/common.proto` |

Observed coverage, recorded by the October 20 review and not recomputed here:

| Dataset | Observed |
| --- | --- |
| Main equipment rows | 8,599 |
| Leftover equipment rows | 1,186 |
| Enchant rows | 246, representing 213 distinct effect IDs (main); 20 rows (leftover) |
| Talent records | 465 across nine classes and 27 trees; 72 emitted prerequisite locations |
| Items with source metadata | 2,119 of 8,599; 6,480 without |
| Main item-level range | 1–100; every main item is phase 1 |

Both catalogs are in scope: the leftover partition holds non-simmable rows that a picker may
still need for imports. Treat "leftover" as a provenance label, not as evidence that an item
does not exist in game.

### 2.1 Build provenance is not a catalog revision

The captured `assets/db_inputs/spell_store_inputs.json` groups identify build `1.60.1.70450`
with two hotfix pushes: `113005` (items, Mage, shared) and `112887` (Druid, Hunter, Paladin,
Priest, racials, Rogue, Shaman, Warlock, Warrior). Partial group updates are intentional.

No payload file carries an immutable catalog revision. We therefore define one:

```text
catalog_revision = "wowsims/forever@0cfb1a0:" + sha256(db.json) + ":" + sha256(leftover_db.json)
```

The build string and hotfix pushes are stored as provenance metadata for display and debugging.
They are never used as identity and never compared for compatibility. A patch that changes only
one group still produces a new manifest.

### 2.2 Native API version is separate again

`IndividualSimSettings` carries an `api_version`. A live Forever export observed on October 20
reported value `2`. This value is independent of the build string and of the catalog revision.

Rule: an adapter accepts exactly the constant declared by the pinned commit and rejects every
other value, including lower ones. The current TBC importer's upper-bound check
(`version <= ours`) is the specific bug that let a Forever payload reach TBC enrichment and
panic.

### 2.3 Evidence status

Facts in this document are labelled as either:

- **Verified** — read directly from the pinned commit (proto and TypeScript sources) during this
  work.
- **Recorded** — taken from the October 20 review's inspection of the same commit, not re-derived.
- **Open** — not established; listed in section 11 with the check that would settle it.

## 3. Reuse decision

### 3.1 Decision

Adopt the pinned WoWSims Forever catalog and talent structure as the versioned data source, and
display its item/enchant/talent names, icons, and stat values. The repository's MIT licence
covers its software and documentation; it does not by itself resolve rights in client-derived
or third-party data. No provider grant was obtained. We proceed on a documented-risk basis.

### 3.2 What is reused

- Numeric and structural item fields: ID, name, icon, type, armor/weapon/hand/ranged types,
  stats, pseudo-stats, item level, quality, unique, limit category, class allowlist, required
  profession, faction restriction, set ID/name, scaling options, random-suffix options, sources,
  item effects, weapon damage/speed.
- Enchant fields: effect, item and spell IDs, name, icon, applicable types, stats, quality,
  phase, carrier item level, class allowlist, required profession, proc kind.
- Talent structure: definition ID, rank spell ID, display name, max points, grid location,
  prerequisite location, per-class tree names and ordering.

### 3.3 What is not reused

- The upstream simulator, rotations, optimizers, EP or combat calculations.
- The upstream React frontend, its state model, or its saved-record format.
- The TBC application's `Import -> EnrichArmory -> ComputeStats` path, which must never receive
  a Forever payload.

### 3.4 Obligations this creates

1. Keep the upstream MIT notice and a visible attribution link to `wowsims/forever`.
2. Record the pinned commit and payload hashes in the repository and in the running service.
3. Do not present the data as Blizzard-verified or endorsed.
4. Keep the catalog behind an adapter so a licensed, self-generated, or provider-supplied
   dataset can replace it without changing the domain, validation rules, or UI contracts.
5. Provide a documented removal path: replacing the dataset must not require a data migration
   of user builds beyond re-resolving identifiers that exist in the new dataset.

### 3.5 Residual risk

This is an accepted, recorded risk, not a cleared legal position. If reuse is later found to be
unavailable, the fallback is a numeric/structural dataset with generic presentation. The
contract's identifier rules and validation semantics are deliberately dataset-neutral so that
fallback remains possible.

## 4. Canonical model

### 4.1 Character

| Field | Rule |
| --- | --- |
| `character_id` | Opaque, server-generated, stable for the character's life. Never derived from name/realm. |
| `name` | Display only; user-supplied; not unique; sanitized on render. |
| `class` | One of the nine upstream classes, stored by our own class key. |
| `race` | Our race key, with faction derived from it. |
| `level` | Integer 1–60. See section 6.4 for what this can and cannot validate. |
| `region`, `realm` | Optional display metadata; never identity. |
| `created_at`, `updated_at` | Server timestamps. |

### 4.2 Build

| Field | Rule |
| --- | --- |
| `build_id` | Opaque, stable across renames. |
| `character_id` | Owner character. |
| `kind` | `current` or `plan`. Exactly one `current` per character. |
| `name` | Plans only; editable; uniqueness within a character is a UI convenience, not identity. |
| `catalog_revision` | Manifest string from section 2.1, stored at write time. |
| `equipment` | Exactly the 17 canonical slots of section 4.3; a slot is either empty or an item instance. |
| `talents` | Map from talent identity to points, section 4.5. Empty map is distinct from "not supplied". |
| `source` | `manual`, `addon`, `sixty`, or `wowsims`; plus import time when applicable, kept separate from `updated_at`. |
| `published` | Plans only; current is publishable; default false. |

### 4.3 Canonical equipment slots

Seventeen statistical slots in display order:

```text
Head, Neck, Shoulder, Back, Chest, Wrist,
Main Hand, Off Hand, Hands, Waist, Legs, Feet,
Finger 1, Finger 2, Trinket 1, Trinket 2, Ranged
```

This order is ours. The mapping to the upstream `ItemSlot` enum is explicit and lives in the
adapter, not in storage. **Open:** confirm the pinned `ItemSlot` enum values and that all
seventeen are populated by the Forever payload; the review established seventeen statistical
slots and no shirt/tabard slots.

### 4.4 Item instance

An equipped item is an instance, not a bare ID:

| Field | Rule |
| --- | --- |
| `item_id` | Upstream `UIItem.id`, resolved inside `catalog_revision`. |
| `variant_ilvl` | Required when the catalog row has a non-empty `scaling_options` map whose keys are alternative item levels; otherwise `0`. |
| `random_suffix_id` | `0` when none; otherwise must appear in that item's `random_suffix_options`. |
| `enchant` | Optional; section 4.6. |

`(item_id, variant_ilvl, random_suffix_id)` is the instance key. Two instances with the same key
are the same item and cannot both be equipped when the catalog row is `unique` or shares a
`limit_category`.

**Open:** confirm whether the Forever picker must expose `scaling_options` variants, and how
`rand_prop_points` scales suffix stats at a chosen variant. The TBC implementation's suffix
scaling (`core.ItemEquipmentBaseStats`) is not evidence for Forever.

### 4.5 Talent allocation

Store allocations by identity, never by position:

```text
talents: { <talent_definition_id>: <points> }
```

`talent_definition_id` is the `definitionId` from the pinned tree JSON; the character's class
scopes the lookup. Where a record has no `definitionId`, fall back to
`<class>:<treeIndex>:<rowIdx>,<colIdx>` (its grid location) and record that fallback in the
build's provenance. `spellId` is a secondary cross-reference only: every rank of a talent
reports the same spell, so it identifies a talent but not a rank.

Rules that follow from storing identity:

- A positional `talents_string` is an interchange encoding, decoded only at the adapter edge
  against the pinned config order.
- A build written under one `catalog_revision` cannot be reinterpreted under another by reusing
  positions. Identifiers are re-resolved; unresolvable identifiers are flagged, never remapped.
- The upstream codec clamps ranks and trims trailing zeros. Our decoder must do neither
  silently: it validates first and rejects malformed input.

### 4.6 Enchant

Upstream states the rule directly: no single ID identifies an enchant. Uniqueness requires
either effect ID + slot, or effect ID + item/spell ID. Store all three, key by effect plus slot:

```text
enchant: { effect_id, item_id, spell_id, slot_type }
```

`slot_type` is the enchant's applicable item type. A build stores the enchant on the equipment
slot, so the slot is already known; validation rejects an enchant whose `type` and `extra_types`
do not admit that slot. `lvl` is the carrier item's level, not a character requirement.

## 5. Identifier domains

| Domain | Canonical identifier | Reject when |
| --- | --- | --- |
| Item | `item_id` + `catalog_revision` (+ variant/suffix) | ID resolves to no row; suffix not in that row's options; variant key absent from `scaling_options` |
| Random suffix | `random_suffix_id` | Not present in `UIDatabase.random_suffixes` or not offered by the item |
| Enchant | `effect_id` + slot (+ item/spell cross-reference) | Effect resolves to no row; resolves to more than one row for that slot; type mismatch |
| Talent | `definition_id` (+ class) | Definition absent from the pinned class trees; points exceed `max_points` |
| Source | `UIItemSource` oneof, normalized to our source kinds | Unknown oneof arm; missing source is `unknown`, not "none" |
| Set | `set_id` | Set ID present without a name, or name without ID — surface as partial, not guessed |

Two substitutions are forbidden because they silently change meaning:

- Enchant `spell_id` used as `effect_id`, or the reverse.
- Talent rank digit position used as talent identity.

## 6. Validation rules

Validation is split into rules we enforce (fail the edit/import) and gaps we display as
unverified (section 9). A save never proceeds on a failed rule.

### 6.1 Equipment

| Rule | Source | Behavior |
| --- | --- | --- |
| Class allowlist | `UIItem.class_allowlist` (empty means no restriction) | Reject item for that character class. |
| Armor proficiency | `armor_type` vs class | Reject. Cloth is universal; leather/mail/plate by class. |
| Weapon proficiency | `weapon_type` vs class | Reject. |
| Ranged proficiency | `ranged_weapon_type` vs class | Reject. |
| Slot eligibility | `type`, `hand_type` | Reject an item that cannot occupy the target slot; one-hand weapons may occupy both melee slots only for dual-wield classes. |
| Unique | `unique` | Reject equipping the same instance twice. |
| Limit category | `limit_category` | Reject a second item sharing a category count already reached. |
| Faction | `faction_restriction` | Reject on mismatch. |
| Profession | `required_profession` | Reject unless the character has it. |
| Weapon combination | `hand_type` | A two-hander empties the off hand; an off-hand cannot coexist with a two-handed main hand. |

Two deliberate departures from upstream behavior:

1. Upstream's setters silently clear conflicting weapons and unique items. We preserve the
   user's existing choice and reject or flag the new one instead. "Preserve and flag" is a
   specification requirement; the setters are not copied.
2. `limit_category` and `unique` are validated across the whole build, not just at the moment of
   assignment, because imports can arrive with conflicts already present.

The class capability inputs behind these rules — armor types, weapon types, ranged types and
dual-wield — must be read from the pinned upstream capability data and verified for Forever. The
TBC tables in `cmd/wowsimcli/cmd/upgrades/candidates.go` are not a valid source for them.

### 6.2 Enchant

- The enchant's `type` or `extra_types` must admit the target slot.
- Class allowlist and required profession, when present, must match the character.
- Applied only to a slot that holds an item; an enchant on an empty slot is invalid.

### 6.3 Talents

The pinned implementation states the semantics precisely; we adopt them as the contract:

| Rule | Definition |
| --- | --- |
| Budget | `level - 9` points for levels 10–60 (51 at 60). Below level 10, zero. |
| Tier gate | A talent at zero-based row index `r` (`rowIdx`) requires `5 * r` points in rows above it within the same tree. |
| Prerequisite | A talent with `prereqLocation` requires that talent at maximum rank and not broken. |
| Rank cap | Points per talent never exceed `maxPoints`. |
| Broken talents | A talent whose gate or prerequisite fails keeps its points but unlocks nothing; it does not count toward the row total; its children are broken too. |
| Add vs remove | Points may always be removed, even when that breaks talents. Points may only be added where the result is not broken for that talent. |

Broken allocations are legal to store and display, with the broken set marked. This matches
upstream and satisfies the specification's "retain and flag existing incompatible choices"
requirement. Our validator differs from upstream in one respect: upstream's positional decoder
clamps ranks and silently converts malformed input. Our decoders validate against identity and
reject rather than clamp.

### 6.4 Character level

`level` drives the talent budget and is stored for display. It cannot validate equipment
legality: the pinned `UIItem` carries no required-character-level field. Upstream uses required
level while generating the catalog but does not carry it into published rows, and `ilvl` is not
a character requirement.

Consequence for the approved scope: **lower-level equipment planning cannot be validated from
this payload.** Options:

1. Support level 60 equipment planning only in the first version, while still allowing lower
   levels for identity and talent budgeting (recommended).
2. Obtain a required-level source before promising lower-level planning.

The specification is revised to the narrower promise pending option 2. An item whose required
level is unknown must not be presented as legal for a lower-level character; if lower levels are
offered at all before this is resolved, that state is shown explicitly as unverified.

## 7. Versioning and migration

1. Every build stores the `catalog_revision` it was written under.
2. On read, identifiers are resolved against the stored revision; if that revision is no longer
   served, the build is read-only and flagged until it is remapped.
3. Remapping is by identity only. An identifier present in the new revision keeps its meaning;
   one that is absent is flagged as unsupported and retained, not deleted.
4. Positional talent strings are decoded only against the revision they were produced under.
5. A schema-version field on our own stored builds is separate from `catalog_revision` and from
   any native `api_version`.

Nothing in this section requires rewriting user data on a catalog bump. It requires resolving
and flagging.

## 8. Imports

### 8.1 Adapter contract

Every route, without exception:

```text
parse -> resolve identifiers -> validate -> present diagnostics -> apply atomically
```

- A failure at any step leaves the target build unchanged and the user's edits intact.
- Partial application is forbidden. This is a deliberate difference from upstream finalization,
  which applies gear even when talents or suffixes were dropped.
- Each route declares the catalog revision and native protocol version it accepts, and rejects
  anything else with a typed error.

### 8.2 Route status

| Route | Status | Mapping notes |
| --- | --- | --- |
| Manual | In scope | Foundation. No external contract. |
| Native WoWSims JSON/link | **Approved, gated (Phase 4)** | Inbound only. Adopt only with exact native API version and a compatible catalog/talent revision. The snapshot link is a payload, not a hosted record; opening it must never imply ownership. |
| Addon (equipped character) | **Gated** | Fixed 17-slot order, item links, positional talent string. No real Forever fixture and no verified client/addon version yet. Talent ordering must be proven equivalent to the pinned config order before any conversion. |
| Sixty Upgrades text/file | **Partial** | Gear shape verified. Enchant effect and spell IDs are distinct domains. Talent records exist in user exports, but the observed `talents[].id` (e.g. `105848`) matches neither the definition ID (`135378`) nor the rank spell (`14523`) for the same talent. Until that domain is resolved, a payload containing talent records is **rejected as a whole** rather than imported gear-only: silently dropping talents would violate the atomic contract. |
| Sixty Upgrades URL | **Blocked** | No provider-supported machine-export contract. Do not scrape. |

### 8.3 Normalization from import payloads

- Slots arrive as provider labels (`HEAD`, `FINGER_1`, `MAIN_HAND`, ...). Map through an explicit
  table to our canonical slots; never rely on array position.
- Item and enchant IDs are resolved against the pinned catalog. Unknown IDs reject the payload.
- Missing fields are distinct from empty fields. An absent talent array means "not supplied";
  an empty array means "supplied and empty".
- Source and import time are recorded independently of edit time.
- Provider stat tables, EP, buffs, consumables, and phases in imported payloads are not
  interpreted as our character state or as catalog versions.

## 9. What this data cannot support

These are explicit product states, not implementation gaps to paper over:

| Area | State | Display rule |
| --- | --- | --- |
| Required character level | Absent from `UIItem` | Unknown; never inferred from `ilvl`. |
| Whole-character stats | No verified formula coverage | Unavailable with a named reason. Never sum item stats as a total. |
| Gems and sockets | Reserved/removed in the pinned proto (`UIDatabase` and `UIItem` reserve the fields) | No socket UI in the first version. |
| Phase filtering | Every main item is phase 1 | Omit the control; it cannot discriminate. |
| Set bonuses | Set ID/name present, bonus effects not modeled | Show set membership only. |
| Proc/use descriptions | Not carried as text | Do not invent; show the effect fields the catalog has. |
| Lower-level equipment legality | See section 6.4 | Level 60 equipment planning only pending a required-level source. |

## 10. Fixtures and verification

Evidence that exists today:

- `docs/superpowers/experiments/forever-fixtures/catalog-sample.json`: attributed, CC BY 4.0,
  17-record numeric/structural Warrior talent projection. It is not a catalog and not a
  production dataset.
- Recorded representative identities: item `260212` (Bad Hat, head/cloth, ilvl 56), item `12640`,
  enchant effect `8219` / carrier item `249503` / spell `1249058` (neck), talent Silent Resolve
  (`14523`, definition `135378`, max 3).

Fixtures to produce for the implementation plan, generated from the pinned payload by a
checked-in script rather than committed wholesale:

1. One record per canonical slot, including one item with `scaling_options` and one with
   `random_suffix_options`.
2. One enchant per applicable slot type, including one with both item and spell IDs and one with
   only an effect ID.
3. One full class tree set with a tier-gated talent, a prerequisite, and a max-rank-1 talent.
4. Deliberately invalid cases: unknown item, unknown enchant effect, ambiguous enchant,
   suffix not offered, over-budget allocation, prerequisite at less than max rank.
5. One rejected native payload with the wrong API version, asserting no state change.

Required tests:

- Catalog adapter round-trips every identifier and rejects unresolved references.
- Enchant uniqueness resolves by effect + slot and refuses ambiguous rows.
- Talent decode/encode round-trips by identity; malformed strings reject rather than clamp.
- Tier, prerequisite, budget and broken-talent rules match the table in section 6.3.
- Equipment conflicts preserve existing gear instead of clearing it.
- Import failure leaves the previous build byte-identical.
- A build pinned to an older catalog revision opens read-only, flags missing identifiers, and
  never reinterprets a positional string.

## 11. Open items

| # | Question | Effect | How to settle |
| --- | --- | --- | --- |
| 1 | Exact pinned `ItemSlot` enum values and whether all 17 are populated | Slot mapping | Read `proto/common.proto` at the pinned commit; assert in an adapter test |
| 2 | Whether the picker must expose `scaling_options` variants, and suffix scaling at a variant | Item instance model | Inspect payload rows with non-empty `scaling_options`; one representative fixture |
| 3 | Talent records without `definitionId` | Fallback identity | Count records missing it across all nine class trees |
| 4 | Sixty `talents[].id` domain | Sixty import route | Crosswalk against a provider statement or a broader sample; until then keep the route rejected |
| 5 | Addon positional order versus pinned config order | Addon import route | Real fixture plus a documented ordering check; do not infer from code |
| 6 | Required character level | Lower-level support | Source separately or keep the level-60 restriction |
| 7 | Whether the pinned commit's native `api_version` matches the observed live value `2` | Native interchange | Read the constant at the pinned commit; test exact-match rejection |
| 8 | Leftover catalog inclusion in the picker | Coverage | Decide whether the picker searches both partitions or resolves leftovers only on import |

None of these blocks writing the Phase 2 manual-editor plan. Items 1, 2, 3 and 7 block the
catalog adapter's first implementation slice; items 4, 5 and 6 gate their respective routes, not
the core editor.
