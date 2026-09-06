<script>
  import { socketColors } from './labels.js';
  import { buildItemTooltip } from './itemTooltip.js';
  import ItemTooltipTrigger from './ItemTooltipTrigger.svelte';

  let { slot, side = 'left' } = $props();
  let itemIconFailed = $state(false);
  let failedGemIcons = $state({});

  let vm = $derived.by(() => buildItemTooltip(slot ?? {}));
  let preferredSide = $derived(side === 'left' ? 'right' : 'left');

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

{#if slot.itemName}
  <ItemTooltipTrigger item={slot} preferredSide={preferredSide}>
    {#snippet children({ id, onpointerenter, onpointerleave, onfocus, onblur })}
      <article class="gear-slot quality-{slot.quality ?? 0}" class:mirror={side === 'right'} data-slot={slot.slotName}>
        <div class="gear-icon-wrap">
          <button
            type="button"
            class="gear-trigger"
            aria-describedby={id}
            onpointerenter={onpointerenter}
            onpointerleave={onpointerleave}
            onfocus={onfocus}
            onblur={onblur}
          >
            {#if slot.ilvl}
              <span class="item-ilvl">{slot.ilvl}</span>
            {/if}
            {#if slot.icon && !itemIconFailed}
              <img
                class="gear-icon"
                src={iconUrl(slot.icon)}
                alt="{slot.itemName || 'Empty'} icon"
                onerror={handleItemImageError}
              />
            {:else}
              <span class="gear-icon-fallback" role="img" aria-label="{slot.slotName} icon unavailable">{slotInitial()}</span>
            {/if}
            {#if slot.sockets?.length}
              <div class="socket-strip" aria-label={`${slot.slotName} sockets`}>
                {#each slot.sockets as socket, index}
                  <span
                    class:empty-socket={!socket.gem}
                    class="socket"
                    title={`${socketColors[socket.color] || 'Unknown'} socket${socket.gem ? `: ${socket.gem.name}` : ': empty'}`}
                  >
                    {#if socket.gem?.icon && !failedGemIcons[index]}
                      <img src={iconUrl(socket.gem.icon)} alt="{socket.gem.name} gem" onerror={() => handleGemImageError(index)} />
                    {:else}
                      <span aria-hidden="true">◇</span>
                    {/if}
                  </span>
                {/each}
              </div>
            {/if}
          </button>
        </div>
        <div class="gear-copy">
          <h3 class="item-name quality-text-{slot.quality ?? 0}">
            <button
              type="button"
              class="name-trigger"
              aria-describedby={id}
              onpointerenter={onpointerenter}
              onpointerleave={onpointerleave}
              onfocus={onfocus}
              onblur={onblur}
            >{slot.itemName}</button>
          </h3>
          {#if slot.enchant}
            <p class="enchant-effect">{slot.enchant.description || slot.enchant.name}</p>
          {/if}
          <details class="gear-details">
            <summary>Details</summary>
            <dl class="gear-detail-list">
              <div><dt>Slot</dt><dd>{slot.slotName}</dd></div>
              {#if vm.typeLabel}
                <div><dt>Type</dt><dd>{vm.typeLabel}</dd></div>
              {/if}
              {#if slot.itemId}
                <div><dt>Item</dt><dd>{slot.itemId} · Phase {slot.phase || '—'}</dd></div>
              {/if}
              {#if vm.suffixLabel}
                <div><dt>Suffix</dt><dd>{vm.suffixLabel}</dd></div>
              {/if}
              {#if vm.weaponLines.length}
                <div><dt>Weapon</dt><dd>{vm.weaponLines.join(' · ')}</dd></div>
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
              {#each vm.sockets as socket, index}
                <div>
                  <dt>{socket.gem ? `Gem ${index + 1}: ${socket.gem.name}` : `Socket ${socketColors[socket.color] ?? 'Unknown'}`}</dt>
                  <dd>{socket.text}</dd>
                </div>
              {/each}
              {#if slot.socketBonus?.stats}
                <div><dt>Socket bonus</dt><dd class:bonus-active={slot.socketBonus.active} class:bonus-inactive={!slot.socketBonus.active}>{slot.socketBonus.active ? 'active' : 'inactive'}{displayStats(slot.socketBonus.stats) ? ` · ${displayStats(slot.socketBonus.stats)}` : ''}</dd></div>
              {/if}
              {#if vm.restrictionLines.length}
                <div><dt>Restrictions</dt><dd>{vm.restrictionLines.join('; ')}</dd></div>
              {/if}
            </dl>
          </details>
        </div>
      </article>
    {/snippet}
  </ItemTooltipTrigger>
{:else}
  <article class="gear-slot quality-0" data-slot={slot.slotName}>
    <div class="gear-copy">
      <p class="slot-caption">{slot.slotName}</p>
      <p class="item-name muted">Empty slot</p>
    </div>
  </article>
{/if}
