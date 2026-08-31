<script>
  let { slot } = $props();
  let itemIconFailed = $state(false);
  let failedGemIcons = $state({});

  const socketColors = {
    0: 'Unknown',
    1: 'Meta',
    2: 'Red',
    3: 'Blue',
    4: 'Yellow',
    5: 'Green',
    6: 'Orange',
    7: 'Purple',
    8: 'Prismatic',
  };

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

<article class="gear-slot quality-{slot.quality ?? 0}" data-slot={slot.slotName}>
  <div class="gear-icon-wrap">
    {#if slot.icon && !itemIconFailed}
      <img class="gear-icon" src={iconUrl(slot.icon)} alt="{slot.itemName || 'Empty'} icon" onerror={handleItemImageError} />
    {:else}
      <span class="gear-icon-fallback" role="img" aria-label="{slot.slotName} icon unavailable">{slotInitial()}</span>
    {/if}
  </div>
  <div class="gear-copy">
    <h3>{slot.slotName}</h3>
    {#if slot.itemName}
      <p class="item-name">{slot.itemName}</p>
      <p class="item-meta">Item {slot.itemId} · Phase {slot.phase || '—'}{slot.setName ? ` · ${slot.setName}` : ''}</p>
    {:else}
      <p class="item-name muted">Empty slot</p>
    {/if}
    <p class="enchant">Enchant: {slot.enchant?.name || 'No Enchant'}</p>
    {#if slot.sockets?.length}
      <div class="socket-row" aria-label={`${slot.slotName} sockets`}>
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
    {#if slot.socketBonus}
      <p class:bonus-active={slot.socketBonus.active} class:bonus-inactive={!slot.socketBonus.active} class="socket-bonus">
        Socket bonus {slot.socketBonus.active ? 'active' : 'inactive'}{displayStats(slot.socketBonus.stats) ? ` · ${displayStats(slot.socketBonus.stats)}` : ''}
      </p>
    {/if}
  </div>
</article>
