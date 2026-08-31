# TBC Upgrade Finder

A local browser application that ranks practical **single-item DPS upgrades** for
one TBC character under the exact assumptions of an imported wowsims
configuration.

Simulation and item data are provided by
[wowsims/tbc-new](https://github.com/wowsims/tbc-new).

## Rebuild UI

Run these commands whenever the Svelte source changes. The Vite build writes the
embedded assets consumed by `wowsimcli`:

```bash
cd ui-finder
rtk npm ci
rtk npm run build
```

## Run

```bash
rtk go run ./cmd/wowsimcli rank-upgrades --no-browser
```

The command binds to loopback only (`127.0.0.1:0` by default; override with
`--addr 127.0.0.1:<port>`), prints the resolved URL, and opens the browser unless
`--no-browser` is set. The server keeps jobs and reports in process memory only;
nothing is written to disk and no character data is cached between runs.

## Input

- **Accepted:** wowsims **individual-sim** export links (the URL fragment after
  `#` from an individual sim page, for example
  `https://wowsims.com/tbc/mage/#eJ...`).
- **Rejected before any ranking or simulation runs:** malformed links, raid-sim
  links, links whose settings API version is newer than the pinned simulator
  revision, and unsupported references in the imported settings. Unsupported
  item IDs, random-suffix IDs, gem IDs, and enchant-effect IDs are each checked
  as typed references before ranking starts. Rejections report a specific typed
  code: `invalid_link`, `unsupported_link`, `incompatible_settings`,
  `incompatible_item`, `incompatible_random_suffix`, `incompatible_gem`, or
  `incompatible_enchant`.

## Armory review

After import, review the complete canonical armory before ranking: character
header, all 17 gear slots, item names and phases, gems, enchants, socket-bonus
status, and the deterministic **unbuffed (base + gear)** raw and derived stat
panels. This is a server-calculated armory view, not a decoded-link summary.

The armory excludes buffs, consumes, and talents from its displayed unbuffed
snapshot. It follows the simulator engine for racial effects, stat
dependencies, random suffixes, socket rules, set bonuses, and static gear
effects. The imported baseline remains unchanged while candidates are tested.

## Workflow

1. **Paste** the export link and press *Import settings*.
2. **Review** the armory as described above before anything is ranked.
3. **Select** content filters (maximum phase, optional inclusion of items without
   source metadata) and simulation budgets (screening and confirmation
   iterations).
4. **Run** ranking with *Start ranking*. Progress reports queued/running work
   and completed candidate runs. *Cancel ranking* stops outstanding work and
   leaves the canceled job without a report.
5. **Read** the report and use *Copy JSON* to place a parseable JSON snapshot of
   the full report on the clipboard.

Items without source metadata are excluded and counted by default; enable
*Include unknown-source items* to rank them anyway.

## Results

Each confirmed candidate reports:

- item, target slot (rings/trinkets are evaluated in both slots), and the
  displaced item(s);
- source and phase;
- the exact applied gem/enchant policy for that candidate;
- DPS delta versus the unchanged imported baseline, relative gain, iteration
  count, and the two-sided 95% interval;
- `Too close to call` instead of a precise rank whenever candidates' 95%
  intervals overlap.

Every report also carries the simulator revision, bundled database revision, and
an assumptions fingerprint (SHA-256 over the imported settings, normalized
filters/policy/options, and both revisions) so results can be compared across
runs.

## Semantics

- The imported baseline is never mutated; every candidate starts from a copy.
- Only the submitted gem/enchant policy is applied, and only where legal; no
  alternative gem/enchant combinations are searched.
- A failed candidate simulation is isolated and listed with its reason; it does
  not invalidate other results.
- Candidates whose simulated 95% intervals overlap are marked
  `too close to call` rather than being given a fake precise order.

## Non-goals

The following are explicitly out of scope: scheduled monitoring, MCP transport,
acquisition pricing, multi-item optimization, non-DPS metrics (tank/healer),
accounts, and remote persistence.

## Attribution

Simulation and item data are provided by
[wowsims/tbc-new](https://github.com/wowsims/tbc-new), whose MIT license also
requests this visible attribution.

## Manual smoke test

```bash
rtk go run ./cmd/wowsimcli rank-upgrades --no-browser
```

Then, in a browser at the printed URL:

1. Paste the contents of
   `cmd/wowsimcli/cmd/upgrades/testdata/fixed_individual_link.txt` and import.
2. Confirm the character header and all 17 gear slots render, including at
   least one socket line and enchant line.
3. Confirm **Raw stats**, **Derived percentages**, and visible **unbuffed (base
   + gear)** stat labeling; confirm the attribution link points to
   `https://github.com/wowsims/tbc-new`.
4. Set maximum phase, screening iterations, and confirmation iterations to `1`;
   start ranking and confirm queued/running progress appears before the report
   table completes.
5. Use *Copy JSON* and confirm the copy status reports success (the clipboard
   value should parse as JSON).
6. Start a fresh job and click *Cancel ranking*; confirm **Ranking canceled.**
   appears and no report table remains.
