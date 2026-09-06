<script>
  import { postJSON } from './api.js';
  import { untrack } from 'svelte';
  import CharacterViewer from './CharacterViewer.svelte';
  import { createViewerAdapter } from './characterViewerFactory.js';
  import { buildPreviewItems, HIDDEN_SLOTS, zamRaceId } from './characterPreview.js';

  const backdrops = import.meta.glob('../../../assets/img/*.jpg', { eager: true, import: 'default' });

  const backdropFor = {
    BalanceDruid: 'balance_druid_background.jpg',
    FeralCatDruid: 'feral_druid_background.jpg',
    FeralBearDruid: 'feral_druid_tank_background.jpg',
    RestorationDruid: 'resto_druid_background.jpg',
    Hunter: 'hunter_background.jpg',
    Mage: 'mage_background.jpg',
    HolyPaladin: 'holy_paladin_background.jpg',
    ProtectionPaladin: 'prot_paladin.jpg',
    RetributionPaladin: 'retribution_paladin.jpg',
    Priest: 'healing_priest_background.jpg',
    Rogue: 'rogue_background.jpg',
    ElementalShaman: 'elemental_shaman_background.jpg',
    EnhancementShaman: 'enhancement_shaman_background.jpg',
    RestorationShaman: 'resto_shaman_background.jpg',
    Warlock: 'warlock_background.jpg',
    DpsWarrior: 'warrior_background.jpg',
    ProtectionWarrior: 'warrior_background.jpg',
  };

  let { race = '', class: playerClass = '', spec = '', gear = [], visualsEnabled = false } = $props();

  let backdropUrl = $derived.by(() => {
    const fileName = backdropFor[spec] ?? backdropFor[playerClass];
    if (!fileName) return '';
    const entry = Object.entries(backdrops).find(([path]) => path.endsWith(`/${fileName}`));
    return entry?.[1] ?? '';
  });

  const raceId = $derived(zamRaceId(race));
  // The viewer accepts import race/class but not the imported appearance;
  // the preview uses the provider defaults. Gender is an explicit body preset.
  let bodyPreset = $state('female');
  const gender = $derived(bodyPreset === 'female' ? 1 : 0);

  let stage = $state('idle'); // unavailable | idle | loading | ready | partial | error | paused
  let errorMessage = $state('');
  let unresolved = $state([]);
  let adapter = $state(null);
  let active = $state(false);
  let viewerConfig = $state(null);
  let viewerHandle = null;
  let generation = 0;

  const itemsSignature = $derived.by(() => {
    const gearPart = gear.map((g) => `${g.slotName}:${g.itemId ?? ''}`).join('|');
    return `${gearPart}|${raceId}|${gender}`;
  });

  async function loadViewer() {
    const token = ++generation;
    stage = 'loading';
    errorMessage = '';
    try {
      const view = await createViewerAdapter();
      adapter = view;
      let items = [];
      let unresolvedNow = [];
      if (view.kind === 'zam') {
        const ids = gear
          .filter((slot) => slot.itemId && !HIDDEN_SLOTS.has(slot.slotName))
          .map((slot) => slot.itemId);
        const result = await postJSON('/api/visuals/resolve', { items: ids });
        if (token !== generation) return;
        const built = buildPreviewItems(gear, result.items);
        items = built.items;
        unresolvedNow = built.unresolved;
      }
      if (token !== generation) return;
      if (items.length === 0 && view.kind === 'zam') {
        stage = 'error';
        errorMessage = 'No equipment has a visible model.';
        return;
      }
      unresolved = unresolvedNow;
      active = true;
      viewerConfig = { raceId, gender, items, appearance: {} };
      stage = unresolvedNow.length > 0 ? 'partial' : 'ready';
    } catch (error) {
      if (token !== generation) return;
      stage = 'error';
      errorMessage = error?.message ?? 'Failed to load the character model.';
    }
  }

  function activate() {
    if (stage === 'loading' || stage === 'error') return;
    loadViewer();
  }

  function pause() {
    if (!active) return;
    active = false;
    viewerConfig = null;
    stage = 'paused';
  }

  function resume() {
    loadViewer();
  }

  function retry() {
    loadViewer();
  }

  function viewError(error) {
    stage = 'error';
    errorMessage = error?.message ?? 'The character model failed to load.';
  }

  // Re-import, race change or body-preset change while the viewer is active
  // must recreate it from the new baseline; stale results are rejected via
  // generation tokens inside loadViewer. State reads are untracked: this
  // effect runs only on the signature, never on stage transitions.
  $effect(() => {
    const key = itemsSignature;
    untrack(() => {
      if (stage === 'ready' || stage === 'partial') {
        loadViewer();
      }
    });
    void key;
  });
</script>

<div
  class="character-stage"
  class:active={active}
  data-region="character-stage"
>
  <div class="stage-backdrop" aria-hidden="true" style:background-image={backdropUrl ? `url('${backdropUrl}')` : 'none'}></div>

  {#if !visualsEnabled}
    <div class="stage-placeholder">
      <span class="stage-kicker">Character preview</span>
      <span class="stage-note">3D preview unavailable — provider integration pending</span>
    </div>
  {:else if !active}
    <div class="stage-placeholder">
      <span class="stage-kicker">Character preview</span>
      <span class="stage-note">Default appearance · imported gear</span>
      {#if stage === 'paused'}
        <button type="button" class="secondary-button" onclick={resume}>Resume 3D</button>
      {:else}
        <button
          type="button"
          class="secondary-button"
          onclick={activate}
          disabled={stage === 'loading' || !raceId}
        >Activate 3D</button>
      {/if}
      {#if stage === 'loading'}
        <span class="stage-status" role="status">Loading character…</span>
      {:else if stage === 'error'}
        <span class="stage-status stage-status-error" role="alert">{errorMessage}</span>
        <button type="button" class="secondary-button" onclick={retry}>Retry</button>
      {/if}
      {#if !raceId}
        <span class="stage-status">This race has no 3D model.</span>
      {/if}
    </div>
  {:else}
    <div class="stage-viewer-wrap">
      {#if viewerConfig}
        <CharacterViewer config={viewerConfig} adapter={adapter} onReady={(handle) => { viewerHandle = handle; }} onError={viewError} />
      {/if}
      {#if stage === 'loading'}
        <div class="stage-overlay" role="status">Loading character…</div>
      {/if}
    </div>
    <div class="stage-footer">
      <div class="stage-label">Default appearance · imported gear</div>
      <div class="stage-controls">
        <label class="stage-preset">
          Body
          <select bind:value={bodyPreset} aria-label="Body preset">
            <option value="female">Female</option>
            <option value="male">Male</option>
          </select>
        </label>
        <button type="button" class="secondary-button" onclick={() => viewerHandle?.rotate(Math.PI / 4)}>Rotate</button>
        <button type="button" class="secondary-button" onclick={() => viewerHandle?.resetView()}>Reset</button>
        <button type="button" class="secondary-button" onclick={pause}>Pause</button>
      </div>
      {#if stage === 'partial'}
        <div class="stage-notes">Partial preview: {unresolved.join(', ')} {unresolved.length === 1 ? 'has' : 'have'} no visible model version.</div>
      {/if}
      {#if stage === 'error'}
        <div class="stage-notes stage-notes-error" role="alert">{errorMessage}</div>
      {/if}
    </div>
  {/if}
</div>
