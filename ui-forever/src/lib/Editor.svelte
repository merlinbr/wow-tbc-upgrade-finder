<script>
  import { untrack } from 'svelte';

  import { api } from './api.js';
  import EditCharacterDialog from './EditCharacterDialog.svelte';
  import GearGrid from './GearGrid.svelte';
  import IdentityHeader from './IdentityHeader.svelte';
  import ItemPicker from './ItemPicker.svelte';
  import StatsPanel from './StatsPanel.svelte';
  import TalentEditor from './TalentEditor.svelte';
  import { href } from './router.svelte.js';

  let { id, tab = 'gear', buildId = '' } = $props();

  let character = $state(null);
  let builds = $state([]);
  let build = $state(null);
  let equipment = $state({});
  let talents = $state({});
  let items = $state({});
  let meta = $state(null);
  let warnings = $state([]);
  let loading = $state(true);
  let error = $state('');
  let actionError = $state('');
  let selectedSlot = $state('');
  let liveViolations = $state([]);
  let pickerSlot = $state('');

  let editDialogOpen = $state(false);
  let planDialogOpen = $state(false);
  let planName = $state('');
  let planSource = $state('');
  let renameDialogOpen = $state(false);
  let renameName = $state('');
  let deleteDialogOpen = $state(false);
  let validationTimer;

  $effect(() => {
    const requestedId = id;
    const requestedBuildId = untrack(() => buildId);
    load(requestedId, requestedBuildId);
  });

  async function load(requestedId, requestedBuildId) {
    loading = true;
    error = '';
    warnings = [];
    try {
      const [characterData, buildList, metaData] = await Promise.all([
        api.get(`/api/characters/${encodeURIComponent(requestedId)}`),
        api.get(`/api/characters/${encodeURIComponent(requestedId)}/builds`),
        api.get('/api/meta'),
      ]);
      character = characterData;
      builds = buildList.builds;
      meta = metaData;
      const target =
        requestedBuildId && builds.some((candidate) => candidate.id === requestedBuildId)
          ? requestedBuildId
          : (builds.find((candidate) => candidate.kind === 'current')?.id ?? builds[0]?.id ?? '');
      await selectBuild(target, { updateUrl: false });
    } catch (requestError) {
      error = requestError.message ?? 'Could not load the editor.';
    } finally {
      loading = false;
    }
  }

  async function selectBuild(targetId, { updateUrl = true } = {}) {
    const target =
      builds.find((candidate) => candidate.id === targetId) ??
      builds.find((candidate) => candidate.kind === 'current') ??
      builds[0] ??
      null;
    if (!target) {
      build = null;
      equipment = {};
      return;
    }
    build = target;
    equipment = JSON.parse(JSON.stringify(target.equipment ?? {}));
    talents = JSON.parse(JSON.stringify(target.talents ?? {}));
    selectedSlot = '';
    pickerSlot = '';
    liveViolations = [];
    await loadItems(equipment);
    if (updateUrl && character) {
      const currentTab = untrack(() => tab);
      window.history.replaceState(
        null,
        '',
        href('editor', { id: character.id, tab: currentTab, buildId: target.id }),
      );
    }
  }

  async function loadItems(currentEquipment) {
    const ids = [
      ...new Set(
        Object.values(currentEquipment)
          .map((instance) => instance.itemId)
          .filter((itemId) => itemId && !items[itemId]),
      ),
    ];
    if (ids.length === 0) {
      return;
    }
    const loaded = await Promise.all(
      ids.map(async (itemId) => {
        try {
          return await api.get(`/api/items/${itemId}`);
        } catch {
          return null;
        }
      }),
    );
    const next = { ...items };
    for (const item of loaded) {
      if (item?.id) {
        next[item.id] = item;
      }
    }
    items = next;
  }

  async function refreshBuilds(selectId) {
    const list = await api.get(`/api/characters/${character.id}/builds`);
    builds = list.builds;
    await selectBuild(selectId ?? build?.id ?? '');
  }

  function tabHref(nextTab) {
    return href('editor', { id: character.id, tab: nextTab, buildId: build?.id ?? '' });
  }

  function onBuildChange(event) {
    selectBuild(event.currentTarget.value);
  }

  function openPlanDialog() {
    planName = '';
    planSource = '';
    actionError = '';
    planDialogOpen = true;
  }

  async function createPlan(event) {
    event.preventDefault();
    actionError = '';
    try {
      const created = await api.post(`/api/characters/${character.id}/builds`, {
        name: planName,
        copyFrom: planSource || undefined,
      });
      planDialogOpen = false;
      await refreshBuilds(created.id);
    } catch (requestError) {
      actionError = requestError.message ?? 'Could not create the plan.';
    }
  }

  async function duplicateBuild() {
    if (!build) {
      return;
    }
    actionError = '';
    try {
      const copy = await api.post(`/api/builds/${build.id}/copy`);
      await refreshBuilds(copy.id);
    } catch (requestError) {
      actionError = requestError.message ?? 'Could not duplicate the build.';
    }
  }

  function openRename() {
    if (!build || build.kind === 'current') {
      return;
    }
    renameName = build.name;
    actionError = '';
    renameDialogOpen = true;
  }

  async function renameBuild(event) {
    event.preventDefault();
    actionError = '';
    try {
      await api.patch(`/api/builds/${build.id}`, {
        name: renameName,
        version: build.version,
      });
      renameDialogOpen = false;
      await refreshBuilds(build.id);
    } catch (requestError) {
      actionError = requestError.message ?? 'Could not rename the plan.';
    }
  }

  function openDelete() {
    if (!build || build.kind === 'current') {
      return;
    }
    actionError = '';
    deleteDialogOpen = true;
  }

  async function deleteBuild() {
    actionError = '';
    try {
      await api.del(`/api/builds/${build.id}`);
      deleteDialogOpen = false;
      await refreshBuilds();
    } catch (requestError) {
      actionError = requestError.message ?? 'Could not delete the plan.';
    }
  }

  function handleCharacterSaved(updated, newWarnings) {
    character = updated;
    warnings = newWarnings ?? [];
    editDialogOpen = false;
  }

  function applyItem(slot, instance) {
    equipment = { ...equipment, [slot]: instance };
    pickerSlot = '';
    loadItems(equipment);
    validateLocal();
  }

  async function validateLocal() {
    try {
      const result = await api.post(`/api/characters/${character.id}/validate`, {
        equipment,
        talents,
      });
      liveViolations = result.violations ?? [];
    } catch {
      liveViolations = [];
    }
  }

  function scheduleValidation() {
    clearTimeout(validationTimer);
    validationTimer = setTimeout(() => {
      validateLocal();
    }, 150);
  }

  function handleTalentsChange(next) {
    talents = next;
    scheduleValidation();
  }
</script>

{#if error}
  <p class="error-text" role="alert">{error}</p>
{:else if loading || !character}
  <p class="muted">Loading editor…</p>
{:else}
  <section class="editor">
    <IdentityHeader
      {character}
      {meta}
      onedit={() => {
        actionError = '';
        editDialogOpen = true;
      }}
    />

    {#if warnings.length > 0}
      <ul class="validation-list warning-list" role="status">
        {#each warnings as warning, index (index)}
          <li><code>{warning.code}</code> {warning.message}</li>
        {/each}
      </ul>
    {/if}

    <div class="build-bar panel">
      <div class="build-selector">
        <label for="build-select">Build</label>
        <select id="build-select" value={build?.id ?? ''} onchange={onBuildChange}>
          {#each builds as candidate (candidate.id)}
            <option value={candidate.id}>
              {candidate.kind === 'current' ? 'Current' : candidate.name}
            </option>
          {/each}
        </select>
      </div>
      <div class="button-row">
        <button type="button" onclick={openPlanDialog}>New plan</button>
        <button type="button" class="secondary-button" onclick={duplicateBuild}>Duplicate</button>
        <button
          type="button"
          class="secondary-button"
          onclick={openRename}
          disabled={build?.kind === 'current'}
        >
          Rename
        </button>
        <button
          type="button"
          class="danger-button"
          onclick={openDelete}
          disabled={build?.kind === 'current'}
        >
          Delete
        </button>
      </div>
    </div>

    {#if actionError}
      <p class="error-text" role="alert">{actionError}</p>
    {/if}

    <div class="tabs" role="tablist" aria-label="Editor views">
      {#each ['gear', 'stats', 'talents'] as value (value)}
        <a
          role="tab"
          href={tabHref(value)}
          aria-selected={tab === value}
          tabindex={tab === value ? 0 : -1}
        >
          {value === 'gear' ? 'Gear' : value === 'stats' ? 'Stats' : 'Talents'}
        </a>
      {/each}
    </div>

    {#if liveViolations.length > 0}
      <ul class="validation-list" role="status">
        {#each liveViolations as violation, index (index)}
          <li><code>{violation.code}</code> {violation.message}</li>
        {/each}
      </ul>
    {/if}

    {#if tab === 'gear'}
      <GearGrid
        {equipment}
        {items}
        onopenslot={(slot) => {
          selectedSlot = slot;
          pickerSlot = slot;
        }}
      />
    {:else if tab === 'stats'}
      <StatsPanel />
    {:else}
      <TalentEditor
        {character}
        {build}
        {talents}
        violations={liveViolations}
        onchange={handleTalentsChange}
      />
    {/if}
  </section>

  {#if pickerSlot}
    <ItemPicker
      open={true}
      slot={pickerSlot}
      {character}
      current={equipment[pickerSlot] ?? null}
      onclose={() => (pickerSlot = '')}
      onapply={(instance) => applyItem(pickerSlot, instance)}
    />
  {/if}

  <EditCharacterDialog
    open={editDialogOpen}
    {character}
    {meta}
    onclose={() => (editDialogOpen = false)}
    onsave={handleCharacterSaved}
  />

  {#if planDialogOpen}
    <div class="dialog-backdrop">
      <dialog open class="panel dialog" aria-modal="true" aria-labelledby="new-plan-title">
        <h2 id="new-plan-title">New plan</h2>
        <form onsubmit={createPlan}>
          <label for="plan-name">Plan name</label>
          <input id="plan-name" bind:value={planName} maxlength="40" />
          <label for="plan-source">Start from</label>
          <select id="plan-source" bind:value={planSource}>
            <option value="">Empty</option>
            {#each builds as candidate (candidate.id)}
              <option value={candidate.id}>
                {candidate.kind === 'current' ? 'Current' : candidate.name}
              </option>
            {/each}
          </select>
          {#if actionError}
            <p class="error-text" role="alert">{actionError}</p>
          {/if}
          <div class="button-row dialog-actions">
            <button type="submit">Create plan</button>
            <button type="button" class="secondary-button" onclick={() => (planDialogOpen = false)}>
              Cancel
            </button>
          </div>
        </form>
      </dialog>
    </div>
  {/if}

  {#if renameDialogOpen}
    <div class="dialog-backdrop">
      <dialog open class="panel dialog" aria-modal="true" aria-labelledby="rename-plan-title">
        <h2 id="rename-plan-title">Rename plan</h2>
        <form onsubmit={renameBuild}>
          <label for="rename-plan-name">Plan name</label>
          <input id="rename-plan-name" bind:value={renameName} maxlength="40" />
          {#if actionError}
            <p class="error-text" role="alert">{actionError}</p>
          {/if}
          <div class="button-row dialog-actions">
            <button type="submit">Rename plan</button>
            <button
              type="button"
              class="secondary-button"
              onclick={() => (renameDialogOpen = false)}
            >
              Cancel
            </button>
          </div>
        </form>
      </dialog>
    </div>
  {/if}

  {#if deleteDialogOpen}
    <div class="dialog-backdrop">
      <dialog open class="panel dialog" aria-modal="true" aria-labelledby="delete-plan-title">
        <h2 id="delete-plan-title">Delete plan?</h2>
        <p>“{build?.name}” will be permanently deleted.</p>
        {#if actionError}
          <p class="error-text" role="alert">{actionError}</p>
        {/if}
        <div class="button-row dialog-actions">
          <button type="button" class="danger-button" onclick={deleteBuild}>Delete plan</button>
          <button type="button" class="secondary-button" onclick={() => (deleteDialogOpen = false)}>
            Cancel
          </button>
        </div>
      </dialog>
    </div>
  {/if}
{/if}
