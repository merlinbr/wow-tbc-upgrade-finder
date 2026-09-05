<script>
  import { socketColors } from './labels.js';
  import ItemTooltip from './ItemTooltip.svelte';

  let { slot, side = 'left' } = $props();
  let itemIconFailed = $state(false);
  let failedGemIcons = $state({});

  function iconUrl(icon) {
    return `https://wow.zamimg.com/images/wow/icons/large/${icon}.jpg`;
  }

  function displayStats(stats) {
    return Object.entries(stats ?? {}).map(([key, value]) => `${key.replaceAll('_', ' ')} ${Number(value).toFixed(2).replace(/\.00$/, '')}`).join(', ');
  }

  function handleItemImageError() {
    itemIconFailed = true;
  }

  function handleGemImageError(index) {
    failedGemIcons[index] = true;
  }

  function slotInitial() {
    return slot.slotName?.trim()?.[0] || '?';
  }
</script>

<article class="gear-slot quality-{slot.quality ?? 0}" class:mirror={side === 'right'} data-slot={slot.slotName}>
  <div class="gear-icon-wrap">
    <button type="button" class="gear-trigger">
      {#if slot.ilvl}
        <span class="item-ilvl">{slot.ilvl}</span>
      {/if}
      {#if slot.icon && !itemIconFailed}
        <img class="gear-icon" src={iconUrl(slot.icon)} alt="{slot.itemName || 'Empty'} icon" onerror={handleItemImageError} />
      {:else}
        <span class="gear-icon-fallback" role="img" aria-label="{slot.slotName} icon unavailable">{slotInitial()}</span>
      {/if}
      {#if slot.sockets?.length}
        <div class="socket-strip" aria-label={`${slot.slotName} sockets`}>
          {#each slot.sockets as socket, index}
            <span class:empty-socket={!socket.gem} class="socket" title={`${socketColors[socket.color] || 'Unknown'} socket${socket.gem ? `: ${socket.gem.name}` : ': empty'}`}>
              {#if socket.gem?.icon && !failedGemIcons[index]}
                <img src={iconUrl(socket.gem.icon)} alt="{socket.gem.name} gem" onerror={() => handleGemImageError(index)} />
              {:else}
                <span aria-hidden="true">◇</span>
              {/if}
            </span>
          {/each}
        </div>
      {/if}
      {#if slot.itemName}
        <ItemTooltip item={slot} variant="full" />
      {/if}
    </button>
  </div>
  <div class="gear-copy">
    {#if slot.itemName}
      <span class="name-trigger">
        <h3 class="item-name quality-text-{slot.quality ?? 0}">{slot.itemName}</h3>
        <ItemTooltip item={slot} variant="full" />
      </span>
      {#if slot.enchant}
        <p class="enchant-effect">{slot.enchant.description || slot.enchant.name}</p>
      {/if}
    {:else}
      <p class="slot-caption">{slot.slotName}</p>
      <p class="item-name muted">Empty slot</p>
    {/if}
    <details class="gear-details">
      <summary>Details</summary>
      <dl class="gear-detail-list">
        <div><dt>Slot</dt><dd>{slot.slotName}</dd></div>
        {#if slot.itemId}
          <div><dt>Item</dt><dd>{slot.itemId} · Phase {slot.phase || '—'}</dd></div>
        {/if}
        {#if slot.setName}
          <div><dt>Set</dt><dd>{slot.setName}</dd></div>
        {/if}
        {#if slot.enchant}
          <div><dt>Enchant</dt><dd>{slot.enchant.name}</dd></div>
        {/if}
        {#if displayStats(slot.stats)}
          <div><dt>Stats</dt><dd>{displayStats(slot.stats)}</dd></div>
        {/if}
        {#if slot.socketBonus?.stats}
          <div><dt>Socket bonus</dt><dd class:bonus-active={slot.socketBonus.active} class:bonus-inactive={!slot.socketBonus.active}>{slot.socketBonus.active ? 'active' : 'inactive'}{displayStats(slot.socketBonus.stats) ? ` · ${displayStats(slot.socketBonus.stats)}` : ''}</dd></div>
        {/if}
      </dl>
    </details>
  </div>
</article>
