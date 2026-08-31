<script>
  let { stats = {}, derivedStats = {} } = $props();

  const labels = {
    strength: 'Strength',
    agility: 'Agility',
    stamina: 'Stamina',
    intellect: 'Intellect',
    spirit: 'Spirit',
    attack_power: 'Attack Power',
    ranged_attack_power: 'Ranged Attack Power',
    spell_damage: 'Spell Damage',
    healing_power: 'Healing Power',
    armor: 'Armor',
    mp5: 'Mana per 5 Seconds',
    melee_hit_rating: 'Melee Hit Rating',
    melee_crit_rating: 'Melee Crit Rating',
    spell_hit_rating: 'Spell Hit Rating',
    spell_crit_rating: 'Spell Crit Rating',
    expertise_rating: 'Expertise Rating',
    haste_rating: 'Haste Rating',
    armor_penetration_rating: 'Armor Penetration Rating',
    melee_hit_percent: 'Melee Hit',
    spell_hit_percent: 'Spell Hit',
    melee_crit_percent: 'Melee Crit',
    spell_crit_percent: 'Spell Crit',
    ranged_hit_percent: 'Ranged Hit',
    ranged_crit_percent: 'Ranged Crit',
    block_percent: 'Block',
  };

  function entries(values) {
    return Object.entries(values ?? {})
      .filter(([, value]) => Number(value) !== 0)
      .map(([key, value]) => ({ key, label: labels[key] ?? key, value: Number(value) }))
      .sort((a, b) => a.label.localeCompare(b.label));
  }

  function rawValue(value) {
    return value.toFixed(2).replace(/\.00$/, '').replace(/(\.\d)0$/, '$1');
  }

  let rawEntries = $derived(entries(stats));
  let derivedEntries = $derived(entries(derivedStats));
</script>

<div class="stat-panels" data-region="stat-panels">
  <section class="stat-panel" aria-labelledby="raw-stats-heading">
    <div class="section-kicker">unbuffed (base + gear)</div>
    <h3 id="raw-stats-heading">Raw stats</h3>
    {#if rawEntries.length}
      <dl class="stat-list">
        {#each rawEntries as stat (stat.key)}
          <div><dt>{stat.label}</dt><dd>{rawValue(stat.value)}</dd></div>
        {/each}
      </dl>
    {:else}
      <p class="muted">No non-zero raw stats reported.</p>
    {/if}
  </section>
  <section class="stat-panel" aria-labelledby="derived-stats-heading">
    <div class="section-kicker">Server snapshot</div>
    <h3 id="derived-stats-heading">Derived percentages</h3>
    {#if derivedEntries.length}
      <dl class="stat-list">
        {#each derivedEntries as stat (stat.key)}
          <div><dt>{stat.label}</dt><dd>{stat.value.toFixed(2)}%</dd></div>
        {/each}
      </dl>
    {:else}
      <p class="muted">No non-zero derived stats reported.</p>
    {/if}
  </section>
</div>
