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

Not yet executed. Gate A remains blocked for a hosted complete item/talent editor; accessible community talent structure needs its own inspection before a narrower verdict.

## Task 3 — import routes

Not yet executed. No real addon or Sixty Upgrades export fixture retained or decoded.

## Retained fixtures

None after Task 1. The live-source hash above is not a committed fixture hash.

## Gate decisions

- Gate A: blocked for the full manual equipment/talent editor; talent-structure source is promising but not yet validation-tested.
- Gate B addon: blocked; client/version/export fixture absent.
- Gate B Sixty Upgrades: blocked; export access/format/reuse scope unverified.
- Whole-character stats: blocked; no complete authoritative formula coverage inspected.
- Gate C hosted ownership/sharing: outside this research phase, not tested.
