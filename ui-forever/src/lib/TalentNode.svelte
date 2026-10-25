<script>
  let { talent, points = 0, state = 'available', onactivate = () => {} } = $props();

  function iconSource() {
    return talent?.icon
      ? `https://wow.zamimg.com/images/wow/icons/medium/${talent.icon}.jpg`
      : '';
  }
</script>

<button
  type="button"
  class="talent-node state-{state}"
  data-testid="talent-node"
  data-state={state}
  data-talent-id={talent.id}
  aria-pressed={points > 0}
  onclick={() => onactivate(talent.id)}
>
  {#if iconSource()}
    <img
      class="talent-icon"
      src={iconSource()}
      alt=""
      loading="lazy"
      onerror={(event) => event.currentTarget.remove()}
    />
  {/if}
  <span class="talent-node-name">{talent.name}</span>
  <span class="talent-node-points">{points}/{talent.maxPoints}</span>
</button>
