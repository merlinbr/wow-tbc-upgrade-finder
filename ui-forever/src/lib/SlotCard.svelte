<script>
  import { slotLabel } from './labels.js';

  let { slot, item, instance, onopen } = $props();

  function qualityClass() {
    return item?.quality ? `quality-${item.quality}` : '';
  }
</script>

<li class="gear-slot-item" data-testid="gear-slot">
  <button
    type="button"
    class="slot-card"
    data-testid={`gear-slot-${slot}`}
    onclick={() => onopen(slot)}
  >
    <span class="slot-label">{slotLabel(slot)}</span>
    {#if instance}
      <span class="slot-item-name {qualityClass()}">
        {item?.name ?? `Unknown item ${instance.itemId}`}
      </span>
      <span class="slot-meta">
        {#if item?.ilvl}
          <span class="ilvl-badge">i{item.ilvl}</span>
        {/if}
      </span>
      {#if instance.enchant}
        <span class="slot-enchant">Enchant effect {instance.enchant.effectId}</span>
      {/if}
    {:else}
      <span class="slot-empty">Empty</span>
      <span></span>
    {/if}
  </button>
</li>
