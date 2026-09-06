<script>
  let { icon, name, size = 16 } = $props();

  let failed = $state(false);
  let url = $derived(icon ? `https://wow.zamimg.com/images/wow/icons/large/${icon}.jpg` : '');

  // Reset the failure state whenever the icon URL changes (re-import or item swap).
  $effect(() => {
    void url;
    failed = false;
  });
</script>

{#if url && !failed}
  <img class="tooltip-icon" src={url} alt={name} style:width="{size}px" style:height="{size}px" onerror={() => (failed = true)} />
{:else}
  <span class="tooltip-icon-fallback" role="img" aria-label={name ? `${name} icon unavailable` : 'icon unavailable'} style:width="{size}px" style:height="{size}px">?</span>
{/if}

<style>
  .tooltip-icon {
    background: #10182a;
    border: 1px solid #85848b;
    border-radius: 3px;
    object-fit: cover;
    flex: none;
  }
  .tooltip-icon-fallback {
    align-items: center;
    background: #10182a;
    border: 1px solid #85848b;
    border-radius: 3px;
    box-sizing: border-box;
    color: #9d9da8;
    display: inline-flex;
    flex: none;
    font-size: 0.8em;
    justify-content: center;
  }
</style>
