<script>
  import { importLink } from './stores.svelte.js';

  let { error = null, onImportPending = () => {}, onImportComplete = () => {} } = $props();
  let link = $state('');
  let submitting = $state(false);

  async function submitImport(event) {
    event.preventDefault();
    const value = link.trim();
    if (!value || submitting) return;
    submitting = true;
    onImportPending();
    const imported = await importLink(value);
    onImportComplete(value, Boolean(imported));
    submitting = false;
  }
</script>

<section class="panel import-panel" aria-labelledby="import-heading" data-region="import-panel">
  <div class="section-kicker">Armory review</div>
  <h2 id="import-heading">Import settings</h2>
  <p class="panel-intro">Paste an individual-sim export link to inspect the character, equipment, and server-calculated stats.</p>
  <form onsubmit={submitImport}>
    <label for="link-input">wowsims export link</label>
    <div class="input-row">
      <input bind:value={link} type="url" id="link-input" name="link" placeholder="https://wowsims.com/tbc/mage/#eJ..." autocomplete="url" required disabled={submitting} />
      <button type="submit" data-action="import" disabled={submitting}>{submitting ? 'Importing…' : 'Import settings'}</button>
    </div>
  </form>
</section>
