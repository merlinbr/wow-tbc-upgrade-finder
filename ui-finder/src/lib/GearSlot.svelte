<script>
  import { socketColors } from './labels.js';
  import ItemTooltipTrigger from './ItemTooltipTrigger.svelte';

  let { slot, side = 'left' } = $props();
  let itemIconFailed = $state(false);
  let failedGemIcons = $state({});

  let preferredSide = $derived(side === 'left' ? 'right' : 'left');

  function iconUrl(icon) {
    return `https://wow.zamimg.com/images/wow/icons/large/${icon}.jpg`;
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
