<script>
  import { copyReport } from './stores.svelte.js';
  import { humanizeEnum } from './labels.js';

  let { report, copyStatus = '' } = $props();


  const slotNames = {
    0: 'Head', 1: 'Neck', 2: 'Shoulder', 3: 'Back', 4: 'Chest', 5: 'Wrist',
    6: 'Hands', 7: 'Waist', 8: 'Legs', 9: 'Feet', 10: 'Finger 1', 11: 'Finger 2',
    12: 'Trinket 1', 13: 'Trinket 2', 14: 'Main Hand', 15: 'Off Hand', 16: 'Ranged',
  };
  const qualityNames = {
    0: 'Junk', 1: 'Common', 2: 'Uncommon', 3: 'Rare',
    4: 'Epic', 5: 'Legendary', 6: 'Artifact', 7: 'Heirloom',
  };
  const sourceKinds = {
    0: 'Unknown source', 1: 'Crafting', 2: 'Quest', 3: 'Reputation', 4: 'PvP',
    5: 'Dungeon', 6: 'Heroic dungeon', 7: 'Raid', 8: 'Heroic raid',
    9: 'Raid finder', 10: 'Flexible raid', 11: 'Sold by vendor',
  };

  function slotLabel(value) {
    return slotNames[value] ?? `Unknown slot (${value ?? '—'})`;
  }

  function sourceKindLabel(value) {
    return sourceKinds[value] ?? `Unknown source (${value ?? '—'})`;
  }
  function qualityLabel(value) {
    return qualityNames[value] ?? `Unknown quality (${value ?? '—'})`;
  }

  function signed(value, digits) {
    const number = Number(value);
    return `${number >= 0 ? '+' : ''}${number.toFixed(digits)}`;
  }

  function appliedText(applied) {
    if (!applied) return '—';
    const gems = applied.gemIds?.length ? `gems ${applied.gemIds.join(', ')}` : 'no gems';
    const enchant = applied.enchantId ? `enchant ${applied.enchantId}` : 'no enchant';
    const sockets = applied.socketChoices?.length ? `sockets ${applied.socketChoices.join(', ')}` : '';
    return [gems, enchant, sockets].filter(Boolean).join(' / ');
  }

  function displacedText(items) {
    return items?.length ? items.map((item) => `${item.name || item.id} (${item.id})`).join(', ') : '—';
  }

  function sourceText(source) {
    if (!source) return sourceKindLabel(0);
    return [sourceKindLabel(source.kind), source.name, source.zone, source.category].filter(Boolean).join(' · ');
  }

  function excludedText(excluded) {
    const parts = [];
    if (excluded?.unknownSource) parts.push(`${excluded.unknownSource} without source metadata`);
    if (excluded?.source) parts.push(`${excluded.source} by source filter`);
    if (excluded?.compatibility) parts.push(`${excluded.compatibility} by compatibility`);
    if (excluded?.policy) parts.push(`${excluded.policy} by gem/enchant policy`);
    return parts.join(', ');
  }
</script>

<section class="panel report-panel" aria-labelledby="report-heading" data-region="report-view">
  <div class="report-header">
    <div>
      <div class="section-kicker">Completed comparison</div>
      <h2 id="report-heading">Upgrade report</h2>
    </div>
    <button type="button" class="secondary-button" onclick={copyReport} data-action="copy-report">Copy JSON</button>
  </div>
  {#if report.character}
    <p class="report-character">Character: {report.character.name || 'Unnamed'} · Level 70 {humanizeEnum(report.character.race, 'Race') || 'Unknown race'} {humanizeEnum(report.character.class, 'Class') || 'Unknown class'}{report.character.spec ? ` · ${humanizeEnum(report.character.spec)}` : ''} · Phase {report.character.phase ?? '—'}</p>
  {/if}
  <p class="copy-status" aria-live="polite">{copyStatus}</p>

  <dl class="report-summary">
    <div><dt>Baseline DPS</dt><dd>{report.baseline?.dps?.toFixed(1) ?? '—'} ±{report.baseline?.dpsStdev?.toFixed(1) ?? '—'} ({report.baseline?.iterations ?? '—'} iterations)</dd></div>
    <div><dt>Simulator revision</dt><dd>{report.simulatorRevision}</dd></div>
    <div><dt>Database revision</dt><dd>{report.databaseRevision}</dd></div>
    <div><dt>Assumptions fingerprint</dt><dd>{report.assumptionsFingerprint}</dd></div>
    <div><dt>Screening / confirmation</dt><dd>{report.assumptions?.screeningIterations} / {report.assumptions?.confirmationIterations} iterations</dd></div>
    <div><dt>Maximum phase</dt><dd>{report.assumptions?.maxPhase}</dd></div>
    <div><dt>Unknown sources</dt><dd>{report.assumptions?.includeUnknown ? 'Included' : 'Excluded'}</dd></div>
    <div><dt>Source filters</dt><dd>{report.assumptions?.sourceKinds?.join(', ') || 'All sources'}{report.assumptions?.sourceNames?.length ? ` · ${report.assumptions.sourceNames.join(', ')}` : ''}</dd></div>
  </dl>

  {#if report.capTruncatedTieRegion}
    <p class="report-note">The finalist tie region was truncated at 50 candidates.</p>
  {/if}
  {#if excludedText(report.excluded)}
    <p class="report-muted">Excluded (not evaluated): {excludedText(report.excluded)}.</p>
  {/if}
  {#if report.excluded?.reasons && Object.keys(report.excluded.reasons).length}
    <details class="report-details"><summary>Exclusion reasons</summary><pre>{JSON.stringify(report.excluded.reasons, null, 2)}</pre></details>
  {/if}
  {#if report.failed?.length}
    <div class="report-failures">
      <h3>Failed simulations</h3>
      <ul>
        {#each report.failed as failure}
          <li>Item {failure.item?.id} ({failure.item?.name || 'unknown'}) in {slotLabel(failure.targetSlot)}: {failure.reason}</li>
        {/each}
      </ul>
    </div>
  {/if}

  <div class="report-table-wrap">
    <table class="report-table">
      <caption>Confirmed single-item upgrades</caption>
      <thead>
        <tr>
          <th scope="col">Rank</th>
          <th scope="col">Item</th>
          <th scope="col">Assumptions</th>
          <th scope="col">Target slot</th>
          <th scope="col">Applied</th>
          <th scope="col">Displaced</th>
          <th scope="col">Source</th>
          <th scope="col">DPS delta</th>
          <th scope="col">Gain</th>
          <th scope="col">Iterations</th>
          <th scope="col">Std error</th>
          <th scope="col">95% CI</th>
        </tr>
      </thead>
      <tbody>
        {#each report.confirmed ?? [] as upgrade}
          <tr>
            <td class:too-close={upgrade.tooCloseToCall}>{upgrade.tooCloseToCall ? 'Too close to call' : upgrade.rank}</td>
            <td>{upgrade.item?.name || 'Unknown'} ({upgrade.item?.id}) · Phase {upgrade.item?.phase ?? '—'} · {qualityLabel(upgrade.item?.quality)}</td>
            <td>
              {#if upgrade.assumptions}
                <details class="report-details"><summary>View</summary><pre>{JSON.stringify(upgrade.assumptions, null, 2)}</pre></details>
              {:else}
                —
              {/if}
            </td>
            <td>{slotLabel(upgrade.targetSlot)}</td>
            <td>{appliedText(upgrade.applied)}</td>
            <td>{displacedText(upgrade.displaced)}</td>
            <td>{sourceText(upgrade.source)}</td>
            <td>{signed(upgrade.dpsDelta, 1)}</td>
            <td>{signed(upgrade.relativeGainPercent, 2)}%</td>
            <td>{upgrade.iterations}</td>
            <td>{Number(upgrade.standardError ?? 0).toFixed(2)}</td>
            <td>[{upgrade.confidenceInterval95?.[0]?.toFixed(1)}, {upgrade.confidenceInterval95?.[1]?.toFixed(1)}]</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  {#if report.assumptions}
    <details class="report-details"><summary>Full assumptions payload</summary><pre>{JSON.stringify(report.assumptions, null, 2)}</pre></details>
  {/if}
</section>
