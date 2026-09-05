<script>
  import CharacterHeader from './CharacterHeader.svelte';
  import CharacterStage from './CharacterStage.svelte';
  import GearSlot from './GearSlot.svelte';
  import StatPanels from './StatPanels.svelte';
  import TalentTrees from './TalentTrees.svelte';

  let { imported } = $props();
  let character = $derived(imported?.character ?? {});
  let gear = $derived(Array.isArray(imported?.gear) ? imported.gear : []);
  // Show the effective content phase (exported settings phase, else highest
  // equipped-item phase) rather than the raw value, which is 0 for exports
  // that omit simulation settings.
  let phase = $derived(imported?.defaults?.maxPhase ?? 0);
  let activeTab = $state('gear');

  const tabs = [
    { id: 'gear', label: 'Gear' },
    { id: 'stats', label: 'Stats' },
    { id: 'talents', label: 'Talents' },
  ];

  const leftColumn = ['Head', 'Neck', 'Shoulder', 'Back', 'Chest', 'Wrist', 'Hands'];
  const rightColumn = ['Waist', 'Legs', 'Feet', 'Finger 1', 'Finger 2', 'Trinket 1', 'Trinket 2'];
  const weaponSlots = ['Main Hand', 'Off Hand', 'Ranged'];

  let gearBySlot = $derived(new Map(gear.map((slot) => [slot.slotName, slot])));

  function slotsFor(names) {
    return names.map((name) => gearBySlot.get(name)).filter(Boolean);
  }

  function gearKey(slot) {
    const gems = (slot.sockets ?? []).map((socket) => socket.gem?.id ?? '').join(',');
    return `${slot.slotName}:${slot.itemId ?? ''}:${slot.icon ?? ''}:${gems}:${slot.enchant?.id ?? ''}`;
  }
</script>

<section class="panel armory-panel" aria-labelledby="armory-heading" data-region="armory-view">
  <CharacterHeader {character} {phase} settingsDigest={imported?.settingsDigest} simulatorRevision={imported?.simulatorRevision} databaseRevision={imported?.databaseRevision} />

  <div class="armory-tabs" role="tablist" aria-label="Armory views">
    {#each tabs as tab}
      <button id="tab-{tab.id}" class="armory-tab" class:active={activeTab === tab.id} role="tab" aria-controls="panel-{tab.id}" aria-selected={activeTab === tab.id} onclick={() => (activeTab = tab.id)}>{tab.label}</button>
    {/each}
  </div>

  {#if activeTab === 'gear'}
    <div id="panel-gear" role="tabpanel" aria-labelledby="tab-gear">
      <div class="gear-grid" aria-label="Equipped gear">
        <div class="gear-column gear-column-left">
          {#each slotsFor(leftColumn) as slot (gearKey(slot))}
            <GearSlot {slot} side="left" />
          {/each}
        </div>
        <CharacterStage race={character.race} class={character.class} spec={character.spec} />
        <div class="gear-column gear-column-right">
          {#each slotsFor(rightColumn) as slot (gearKey(slot))}
            <GearSlot {slot} side="right" />
          {/each}
        </div>
      </div>
      <div class="weapon-strip" aria-label="Weapons">
        {#each slotsFor(weaponSlots) as slot (gearKey(slot))}
          <GearSlot {slot} />
        {/each}
      </div>
    </div>
  {:else if activeTab === 'stats'}
    <div id="panel-stats" role="tabpanel" aria-labelledby="tab-stats">
      <StatPanels stats={imported?.stats} derivedStats={imported?.derivedStats} />
    </div>
  {:else}
    <div id="panel-talents" role="tabpanel" aria-labelledby="tab-talents">
      <TalentTrees class={character.class} talentsString={imported?.talentsString} />
    </div>
  {/if}
</section>
