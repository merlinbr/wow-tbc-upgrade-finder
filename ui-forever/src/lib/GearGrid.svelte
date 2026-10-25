<script>
  import SlotCard from './SlotCard.svelte';
  import { SLOT_ORDER } from './labels.js';

  let { equipment = {}, items = {}, onopenslot = () => {} } = $props();

  const left = SLOT_ORDER.slice(0, 8);
  const right = SLOT_ORDER.slice(8);

  function instanceFor(slot) {
    return equipment[slot] ?? null;
  }

  function itemFor(slot) {
    const instance = instanceFor(slot);
    return instance ? (items[instance.itemId] ?? null) : null;
  }
</script>

<div class="gear-grid">
  {#each [left, right] as column, index (index)}
    <ul class="gear-column">
      {#each column as slot (slot)}
        <SlotCard
          {slot}
          item={itemFor(slot)}
          instance={instanceFor(slot)}
          onopen={onopenslot}
        />
      {/each}
    </ul>
  {/each}
</div>
