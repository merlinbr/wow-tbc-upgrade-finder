# Forever data/import feasibility — experiment log

Plan: `docs/plans/2026-10-01-forever-data-import-feasibility.md`.
Execution date: 2026-10-02 UTC (the filename retains the plan date).
Research-only environment: Windows, Python 3.14, Git, existing Node/Playwright installation from the original checkout. No dependencies installed, client files inspected, accounts accessed, provider messages sent, or application code changed.

## Baseline and execution mode

- Worktree: `C:/Users/merli/worktrees/wow-upgrade-agent/forever-feasibility`.
- Branch: `research/forever-feasibility`.
- Baseline HEAD: `6c5b82653d64e0ab4251c9f0c95d2866d8cc553e`.
- `git status --short` was empty before inline work.
- The user explicitly approved inline execution after the delegate runner failed. No delegate did research; temporary authored agent definitions were removed.
- The original checkout's pre-existing changes remain outside this worktree.
- Read-only context: approved design, README, operator guide, prior API/asset rights research, armory/gear/talent components and the TBC importer. The current UI references TBC trees and ZAM icons and the importer validates against a bundled TBC database; none were treated as Forever data.

## Task 1 — primary-source access and permission evidence

### Direct HTTP check

Executed at `2026-10-02T17:07:20.542024+00:00` with Python urllib and a descriptive research User-Agent. The response was examined in memory; no proprietary catalog was saved to the repository.

```python
import urllib.request
request = urllib.request.Request(
    url,
    headers={'User-Agent': 'ForeverArmoryFeasibility/1.0 (public documentation inspection)'},
)
with urllib.request.urlopen(request, timeout=20) as response:
    body = response.read(600000)
    print(url, response.status, response.headers.get('content-type'), len(body))
```

| URL | Observation |
| --- | --- |
| `https://sixtyupgrades.com/forever/terms` | 200, text/html, 1,820 bytes; SPA shell, not rendered legal text |
| `https://www.wowhead.com/forever` | 200, text/html, 55,946 bytes; no usable catalog export established |
| `https://wowforever.space/data-updates/` | 200, HTML, 25,340 bytes; versioned September 15 demo snapshot |
| `https://forever-codex.com/items/` | 200, HTML, 198,057 bytes; public UI, not a granted bulk export |
| `https://api.github.com/repos/ElliotWood/Forever` | 200, JSON, 15,245 bytes; public repository metadata |
| `https://www.curseforge.com/wow/addons/sixty-upgrades-exporter` | 403; stopped, did not bypass |

A subsequent default urllib request to Wowhead returned 403. Log both outcomes; do not claim reliable automated access or retry with evasion. Content extraction also reported Wowhead's page as dynamically rendered. Guessed Wowhead/ZAM `/terms-of-service` paths returned 404 and were not used to infer permission.

### Rendered Sixty Upgrades legal page

Used existing `@playwright/test` from the original checkout, not an installed dependency in this worktree. Launched an isolated headless Chromium context with no user's cookies or storage. Only public terms and homepage were opened; no sign-in/editor state was accessed or persisted.

```javascript
const {chromium} = require('C:/Users/merli/Documents/Projects/wow-upgrade-agent/ui-finder/node_modules/@playwright/test');
(async () => {
  const browser = await chromium.launch({headless: true});
  try {
    const page = await browser.newPage();
    const response = await page.goto('https://sixtyupgrades.com/forever/terms', {
      waitUntil: 'domcontentloaded', timeout: 30000,
    });
    await page.waitForTimeout(2000);
    console.log(response.status(), await page.locator('body').innerText());
  } finally {
    await browser.close();
  }
})().catch(error => { console.error(error.message); process.exitCode = 1; });
```

Observed HTTP 200 and rendered §2 “USE LICENSE”: temporary personal non-commercial transitory viewing; prohibitions on copying, public display, transfer/mirroring and reverse engineering software. This does not prove a user-produced export is forbidden or permitted for our service; its exact interoperability scope needs clarification. Homepage HTTP 200 explicitly advertises multiple characters and gear sets. Ignore unrelated consent/ad overlay text in the legal analysis.

### Pinned software candidate

GitHub API lookup of `ElliotWood/Forever` returned default branch `master`. Its current commit lookup returned `1e0474d94dc0ce7b38b4bf15f09416d20f797688`.

Read selected raw files at that commit, all HTTP 200:

- `LICENSE`: MIT, copyright 2022 wowsims team, software/documentation grant and notice condition.
- `docs/DATA_SOURCES.md`: documents client data, Wowhead tooltip/planner scraping, inherited Classic/SoD inputs and differences between client and server mechanics. Its claims are repository-author evidence, not independent client testing by this project.
- `assets/db_inputs/forever_client_build.txt`: `1.60.1.70178`.

No clone or binary/database download was performed. A code licence is not treated as clearance for third-party data inputs.

### Talent source and licence

Read `https://talentsforever.com/about`, `/terms`, `/contact` and `/data.json`. Terms explicitly permit copying/adapting/commercial use of the site's own reading/checking/arranging work under CC BY 4.0 with credit, while excluding Blizzard material. The homepage reports beta build `1.60.1.70170`.

The full JSON was parsed transiently, not committed:

- HTTP 200; 1,613,674 bytes.
- SHA-256 of observed source bytes: `fea556416a6f9ac4c8573b7aa98ca60d7aa576f947e906f5906edae50018636c`.
- `license`: `CC-BY-4.0`.
- `attribution`: `Data from talentsforever.com (https://talentsforever.com)`.
- `generated`: `2026-10-01`.
- `talents` class keys: Warrior, Paladin, Hunter, Rogue, Priest, Shaman, Mage, Warlock, Druid.

The hash is an observation, not a pinned fetch URL or a promise the live file will remain unchanged. Task 2 may retain a small attributed numeric/structural projection only; names, tooltip text, icons and art are excluded.

### Other provenance checks

- `wowforever.space/sources/` identifies a September 15 demo snapshot and explicitly says it is not independently footage-verified. Its coverage page reports 670 supplied / 662 missing rank descriptions for that snapshot; do not apply those counts to the current upstream JSON.
- `forever-codex.com/sources/` reports beta `1.60.1.70170`, October 2 generation, client/hotfix provenance, Classic-derived loot/vendor sources, and withheld level-50–60 content. These are provider claims, not our own client observations. No site-wide redistribution grant was established.
- Blizzard API terms and addon policy were fetched/read as public primary legal/policy documents. No API credential or Forever profile response was tested.
- The Classic namespace page's readable extraction was dynamic/unavailable; no Forever-specific namespace conclusion was drawn.

### Task 1 review

Source/rights findings are separated from gameplay accuracy. Positive community/code licences cite their actual scope; proprietary catalog and importer permissions remain blocked. No full catalog, private export, or artwork was retained. Report sections for Tasks 2–3 are explicitly not completed rather than filled with guessed data.

## Task 2 — catalog and rule coverage

### Structural audit and retained projection

Fetched `https://talentsforever.com/data.json` again. Source hash matched the Task 1 observation. Transiently inspected all nine classes / 27 trees / 467 class talent records. Observed positive integer rank caps, one-based row/column coordinates, 71 prerequisite names resolving within their own tree, zero duplicate grid positions and zero `complete`/description-count gaps. No game talent/spell identifier field was in the class-record key union. Provider verification flags and rank text were inspected, not independently validated against the client.

Retained `forever-fixtures/catalog-sample.json`: 17 numeric/structural records for the first Warrior tree, with attribution to Chris Baldwin / talentsforever.com, CC BY 4.0 link and a changes statement. Removed spell names, descriptions, artwork/icon references, Classic text, costs and unrelated data. Resolved same-tree prerequisite names to source indexes. IDs are snapshot locators (`2026-10-01:Warrior:0:<sourceIndex>`), explicitly not stable game IDs.

This is a small permitted community-work projection, not an application item catalog or clearance to republish mixed Blizzard text/art. For a future save format, tie allocations to an immutable data version and explicitly resolve migrations; do not load old positions against a new live tree.

Reproduce the source-consistency audit without retaining the mixed-content JSON:

```bash
python - <<'PY'
import hashlib, json, urllib.request
raw = urllib.request.urlopen('https://talentsforever.com/data.json', timeout=30).read(2500000)
assert hashlib.sha256(raw).hexdigest() == 'fea556416a6f9ac4c8573b7aa98ca60d7aa576f947e906f5906edae50018636c', 'Source changed; inspect a new version before comparing results'
data = json.loads(raw)
total = edges = 0
for player_class, block in data['talents'].items():
    assert len(block['trees']) == 3
    for tree in block['trees']:
        names = [t['name'] for t in tree['talents']]
        positions = set()
        for talent in tree['talents']:
            assert isinstance(talent['max'], int) and talent['max'] > 0
            assert talent['row'] >= 1 and 1 <= talent['col'] <= 4
            position = (talent['row'], talent['col'])
            assert position not in positions
            positions.add(position)
            assert talent['complete'] and len(talent['desc']) == talent['max']
            if talent.get('req'):
                assert names.count(talent['req']) == 1
                edges += 1
            total += 1
assert len(data['talents']) == 9 and total == 467 and edges == 71
print('Source structure: 9 classes, 27 trees, 467 talents, 71 resolved prerequisites')
PY
```

Ran the plan's evidence-wrapper check: `Evidence wrapper valid: 17 records`. Additional checks passed for unique snapshot locators, in-range prerequisites, attribution/licence and absence of excluded text/icon/art fields. Retained file: **8,839 bytes**, SHA-256 `0f9cc51c43ad7775368b2d6aa017ddc66afc632d20eaea714255fddc25e84341`.

### Provider UI cases

Primary rule descriptions: `https://talentsforever.com/warrior`. Selected level sets the point budget; rows require five points per tier and arrow prerequisites must be maxed. Observed cases:

| Input / setup | Expected | Observed |
| --- | --- | --- |
| Empty level 60 build | 51 points available | 51 |
| First-tree source index 3 at zero spent | Reject locked row-2 point | Rank stays 0 |
| First-tree source index 0, attempt rank 4 with cap 3 | Reject fourth point | Rank stays 3 |
| First row: index 0 rank 3, index 1 rank 2; add index 3 | Allow row-2 point after five spent | Rank becomes 1 |
| Row 3 open with ten points, prerequisite index 4 still 0; add index 6 | Reject unmet prerequisite | Rank stays 0 |
| Max prerequisite index 4 to rank 5; add index 6 | Allow point | Rank becomes 1 |
| Empty level 30 build | 21 points available | 21 |
| Empty level 10 build; attempt two points at index 0 | Allow one, reject second | Rank 1; zero points left |

**Result: 8/8 assertions passed.** These are observations of the provider UI only, not proof of game-server behavior or our future importer/validator.

The first combined browser probe failed after six checks because the provider restored the prior build within the same browser context. Evidence: revisiting after one point restored rank 1, `localStorage` key `tf_last_Warrior`, and a “Picked up your last Warrior build” notice; a fresh context showed rank 0. The test's empty-build assumption was wrong, not a demonstrated level-budget bug. Re-ran with one fresh context per scenario and all eight checks passed. No user browser state was accessed.

Reproduce using the already installed Playwright package from this Windows checkout (adjust only its absolute module path on another machine):

```bash
node - <<'JS'
const assert = require('node:assert/strict');
const {chromium} = require('C:/Users/merli/Documents/Projects/wow-upgrade-agent/ui-finder/node_modules/@playwright/test');
(async () => {
  const browser = await chromium.launch({headless: true});
  let context, page, checks = 0;
  const start = async level => {
    if (context) await context.close();
    context = await browser.newContext();
    page = await context.newPage();
    await page.goto('https://talentsforever.com/warrior', {
      waitUntil: 'domcontentloaded', timeout: 30000,
    });
    await page.locator('#level').selectOption(String(level));
  };
  const talent = name => page.getByRole('button', {
    name: new RegExp('^' + name + ', rank '),
  });
  const rank = async name => Number((await talent(name).getAttribute('aria-label')).match(/rank (\d+) of/)[1]);
  const add = async (name, count) => {
    for (let i = 0; i < count; i++) await talent(name).click();
  };
  const pass = label => { checks++; console.log('PASS', label); };
  try {
    await start(60);
    assert.match(await page.locator('body').innerText(), /Points left: 51/);
    pass('level 60 budget 51');
    await add('Improved Charge', 1);
    assert.equal(await rank('Improved Charge'), 0);
    pass('row 2 locked below five tree points');
    await add('Improved Heroic Strike', 4);
    assert.equal(await rank('Improved Heroic Strike'), 3);
    pass('rank cap 3 refuses fourth point');
    await add('Deflection', 2);
    await add('Improved Charge', 1);
    assert.equal(await rank('Improved Charge'), 1);
    pass('row 2 unlocks at five tree points');
    await add('Deflection', 3);
    await add('Improved Charge', 1);
    await add('Anger Management', 1);
    assert.equal(await rank('Anger Management'), 0);
    pass('unmet maxed prerequisite blocks allocation with row open');
    await add('Improved Tactical Mastery', 5);
    await add('Anger Management', 1);
    assert.equal(await rank('Anger Management'), 1);
    pass('maxed prerequisite permits allocation');
    await start(30);
    assert.match(await page.locator('body').innerText(), /Points left: 21/);
    pass('level 30 budget 21');
    await start(10);
    await add('Improved Heroic Strike', 2);
    assert.equal(await rank('Improved Heroic Strike'), 1);
    assert.match(await page.locator('body').innerText(), /Points left: 0/);
    pass('level 10 budget refuses second point');
    console.log('TOTAL', checks, 'provider UI observations passed');
  } finally {
    await browser.close();
  }
})().catch(error => { console.error(error.message); process.exitCode = 1; });
JS
```

### Equipment and statistics stop conditions

No equipment source passed Task 1's rights/acquisition gate, so no armor/weapon/ring/enchant records were copied and no equipment legality cases were claimed as tested. Expected future cases remain: too-high required level, wrong slot, and two-handed weapon/off-hand conflict. They need permitted Forever records and actual restrictions before becoming verified acceptance examples.

The public Codex source page reports unavailable level-50–60 data and Classic-derived loot. Coverage is provider-reported, not a full collection audit. No actual full-character stat engine/formulas were tested. The sample's single class-tree coverage does not imply a production all-class editor is ready.

## Task 3 — import routes

Not yet executed. No real addon or Sixty Upgrades export fixture retained or decoded.

## Retained fixtures

After Task 2, only `forever-fixtures/catalog-sample.json` is retained: an attributed numeric/structural talent projection. Its file hash is recorded with the artifact checks below. No complete live JSON, equipment catalog, addon export, or Sixty Upgrades export is retained. The live-source hash above is distinct from the committed fixture hash.

## Gate decisions

- Gate A: blocked for the full manual equipment/talent editor. Numeric talent evidence is available and eight provider-UI cases pass, but item reuse/acquisition, full presentation rights and versioned talent identity still need decisions.
- Gate B addon: blocked; client/version/export fixture absent.
- Gate B Sixty Upgrades: blocked; export access/format/reuse scope unverified.
- Whole-character stats: blocked; no complete authoritative formula coverage inspected.
- Gate C hosted ownership/sharing: outside this research phase, not tested.
