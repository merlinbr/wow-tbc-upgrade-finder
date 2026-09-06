<script>
  import { avgItemLevel, classColor, classIcon } from './identity.js';
  import { humanizeEnum } from './labels.js';

  let { character = {}, phase = 0, gear = [], settingsDigest = '', simulatorRevision = '', databaseRevision = '' } = $props();

  const professionNames = {
    1: 'Alchemy',
    2: 'Blacksmithing',
    3: 'Enchanting',
    4: 'Engineering',
    5: 'Herbalism',
    6: 'Inscription',
    7: 'Jewelcrafting',
    8: 'Leatherworking',
    9: 'Mining',
    10: 'Skinning',
    11: 'Tailoring',
  };

  let professions = $derived((character.professions ?? []).map((profession) => professionNames[profession] ?? String(profession)));
  let color = $derived(classColor(character.class));
  let avatarUrl = $derived(classIcon(character.class));
  let avatarFailed = $state(false);
  let itemLevel = $derived(avgItemLevel(gear));

  function digest(value) {
    if (!value) return '—';
    return `${value.slice(0, 16)}…`;
  }
</script>

<div class="character-header">
  <div class="character-identity">
    <div class="identity-avatar" aria-hidden="true">
      {#if avatarUrl && !avatarFailed}
        <img src={avatarUrl} alt="" onerror={() => (avatarFailed = true)} />
      {:else}
        <span class="avatar-fallback" style:background={color || '#536b8a'}>{(character.name || '?').trim()[0]?.toUpperCase() || '?'}</span>
      {/if}
    </div>
    <div class="identity-copy">
      <div class="section-kicker">Imported character</div>
      <h2 id="armory-heading" class="character-name" style:color={color || 'var(--text)'}>{character.name || 'Unnamed character'}</h2>
      <p class="character-subtitle">Level 70 {humanizeEnum(character.race, 'Race') || 'Unknown race'} · {character.spec ? humanizeEnum(character.spec) : humanizeEnum(character.class, 'Class') || 'Unknown class'}</p>
      <div class="identity-chips" aria-label="Character facts">
        <span class="chip"><span class="chip-label">Avg ilvl</span><strong>{itemLevel > 0 ? itemLevel : '—'}</strong></span>
        <span class="chip"><span class="chip-label">Phase</span><strong>{phase || '—'}</strong></span>
        <span class="chip"><span class="chip-label">Professions</span><strong>{professions.length ? professions.join(', ') : 'None'}</strong></span>
      </div>
      <p class="identity-note">No ratings — local simulation import</p>
    </div>
  </div>
  <div class="character-actions">
    <a class="find-upgrades" href="#ranking-heading">Find upgrades</a>
    <details class="import-details">
      <summary>Import details</summary>
      <dl>
        <div><dt>Settings digest</dt><dd title={settingsDigest}>{digest(settingsDigest)}</dd></div>
        <div><dt>Simulator revision</dt><dd>{simulatorRevision || '—'}</dd></div>
        <div><dt>Database revision</dt><dd>{databaseRevision || '—'}</dd></div>
      </dl>
    </details>
  </div>
</div>
