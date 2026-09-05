<script>
  import GearSlot from './GearSlot.svelte';
  import StatPanels from './StatPanels.svelte';
  import { humanizeEnum } from './labels.js';

  let { imported } = $props();
  let character = $derived(imported?.character ?? {});
  let gear = $derived(Array.isArray(imported?.gear) ? imported.gear : []);
  let professions = $derived((character.professions ?? []).map((profession) => {
    const names = {
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
    return names[profession] ?? String(profession);
  }));

  function gearKey(slot) {
    const gems = (slot.sockets ?? []).map((socket) => socket.gem?.id ?? '').join(',');
    return `${slot.slotName}:${slot.itemId ?? ''}:${slot.icon ?? ''}:${gems}:${slot.enchant?.id ?? ''}`;
  }

  function digest(value) {
    if (!value) return '—';
    return `${value.slice(0, 16)}…`;
  }

</script>

<section class="panel armory-panel" aria-labelledby="armory-heading" data-region="armory-view">
  <div class="character-header">
    <div>
      <div class="section-kicker">Imported character</div>
      <h2 id="armory-heading">{character.name || 'Unnamed character'}</h2>
      <p class="character-subtitle">Level 70 {humanizeEnum(character.race, 'Race') || 'Unknown race'} · {character.spec ? humanizeEnum(character.spec) : humanizeEnum(character.class, 'Class') || 'Unknown class'}</p>
    </div>
    <dl class="character-facts">
      <div><dt>Professions</dt><dd>{professions.length ? professions.join(', ') : 'None'}</dd></div>
      <div><dt>Phase</dt><dd>{character.phase ?? '—'}</dd></div>
      <div><dt>Settings digest</dt><dd title={imported?.settingsDigest || ''}>{digest(imported?.settingsDigest)}</dd></div>
    </dl>
  </div>

  <div class="gear-grid" aria-label="Equipped gear">
    <div class="gear-column gear-column-left">
      {#each gear.slice(0, 8) as slot (gearKey(slot))}
        <GearSlot {slot} />
      {/each}
    </div>
    <div class="gear-center" aria-hidden="true"></div>
    <div class="gear-column gear-column-right">
      {#each gear.slice(8, 17) as slot (gearKey(slot))}
        <GearSlot {slot} />
      {/each}
    </div>
  </div>

  <StatPanels stats={imported?.stats} derivedStats={imported?.derivedStats} />
</section>
