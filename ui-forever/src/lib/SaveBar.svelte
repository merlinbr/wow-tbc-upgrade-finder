<script>
  let { dirty = false, saving = false, version = 1, onsave = () => {} } = $props();
</script>

<div class="panel save-bar" data-testid="save-bar">
  <div class="save-state">
    {#if saving}
      <span class="muted">Saving…</span>
    {:else if dirty}
      <span class="unsaved">Unsaved changes</span>
    {:else}
      <span class="saved">All changes saved</span>
    {/if}
    <span class="muted">Version {version}</span>
  </div>
  <button type="button" onclick={onsave} disabled={!dirty || saving}>Save changes</button>
</div>
