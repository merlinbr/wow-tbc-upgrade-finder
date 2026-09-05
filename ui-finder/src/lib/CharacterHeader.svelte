<script>
  import { humanizeEnum } from './labels.js';

  let { character = {}, phase = 0, settingsDigest = '', simulatorRevision = '', databaseRevision = '' } = $props();

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

  function digest(value) {
    if (!value) return '—';
    return `${value.slice(0, 16)}…`;
  }
</script>

<div class="character-header">
  <div class="character-identity">
    <div class="section-kicker">Imported character</div>
    <h2 id="armory-heading">{character.name || 'Unnamed character'}</h2>
    <p class="character-subtitle">Level 70 {humanizeEnum(character.race, 'Race') || 'Unknown race'} · {character.spec ? humanizeEnum(character.spec) : humanizeEnum(character.class, 'Class') || 'Unknown class'}</p>
    <dl class="character-facts">
      <div><dt>Professions</dt><dd>{professions.length ? professions.join(', ') : 'None'}</dd></div>
      <div><dt>Phase</dt><dd>{phase || '—'}</dd></div>
    </dl>
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
