<script>
  import { onMount } from 'svelte';

  import { api } from './api.js';
  import { label } from './labels.js';
  import { createCharacter } from './stores.svelte.js';

  let { open = false, onclose = () => {}, oncreated = () => {} } = $props();

  let meta = $state(null);
  let metaError = $state('');
  let name = $state('');
  let className = $state('');
  let specName = $state('');
  let raceName = $state('');
  let level = $state(60);
  let professionOne = $state('');
  let professionTwo = $state('');
  let submitting = $state(false);
  let violations = $state([]);
  let errorMessage = $state('');

  let wasOpen = false;

  onMount(async () => {
    try {
      meta = await api.get('/api/meta');
      resetForm();
    } catch (error) {
      metaError = error.message ?? 'Could not load character options.';
    }
  });

  $effect(() => {
    if (open && !wasOpen && meta) {
      resetForm();
    }
    wasOpen = open;
  });

  function resetForm() {
    className = meta.classes[0] ?? '';
    specName = meta.specs[className]?.[0]?.key ?? '';
    raceName = meta.classRaces[className]?.[0] ?? '';
    name = '';
    level = 60;
    professionOne = '';
    professionTwo = '';
    violations = [];
    errorMessage = '';
  }

  function onClassChange() {
    specName = meta.specs[className]?.[0]?.key ?? '';
    const allowed = meta.classRaces[className] ?? [];
    if (!allowed.includes(raceName)) {
      raceName = allowed[0] ?? '';
    }
  }

  function specOptions() {
    return meta?.specs[className] ?? [];
  }

  function raceAllowed(race) {
    return (meta?.classRaces[className] ?? []).includes(race);
  }

  async function submit(event) {
    event.preventDefault();
    if (submitting || !meta) {
      return;
    }
    submitting = true;
    violations = [];
    errorMessage = '';
    try {
      const professions = [professionOne, professionTwo].filter(Boolean);
      const character = await createCharacter({
        name,
        class: className,
        spec: specName,
        race: raceName,
        level: Number(level),
        professions,
      });
      resetForm();
      oncreated(character);
    } catch (error) {
      violations = error.violations ?? [];
      errorMessage = error.message ?? 'Could not create character.';
    } finally {
      submitting = false;
    }
  }
</script>

{#if open}
  <div class="dialog-backdrop">
    <dialog
      open class="panel dialog"
      aria-modal="true"
      aria-labelledby="create-character-title"
    >
      <h2 id="create-character-title">Create character</h2>
      {#if metaError}
        <p class="error-text" role="alert">{metaError}</p>
      {/if}
      <form onsubmit={submit}>
        <label for="character-name">Name</label>
        <input id="character-name" bind:value={name} maxlength="40" autocomplete="off" />

        <div class="field-row">
          <div>
            <label for="character-class">Class</label>
            <select id="character-class" bind:value={className} onchange={onClassChange}>
              {#each meta?.classes ?? [] as classKey (classKey)}
                <option value={classKey}>{label(classKey)}</option>
              {/each}
            </select>
          </div>
          <div>
            <label for="character-spec">Specialization</label>
            <select id="character-spec" bind:value={specName}>
              {#each specOptions() as spec (spec.key)}
                <option value={spec.key}>{spec.name}</option>
              {/each}
            </select>
          </div>
        </div>

        <div class="field-row">
          <div>
            <label for="character-race">Race</label>
            <select id="character-race" bind:value={raceName}>
              {#each meta?.races ?? [] as race (race)}
                <option value={race} disabled={!raceAllowed(race)}>{label(race)}</option>
              {/each}
            </select>
          </div>
          <div>
            <label for="character-level">Level</label>
            <input id="character-level" type="number" min="1" max="60" bind:value={level} />
          </div>
        </div>

        <div class="field-row">
          <div>
            <label for="profession-1">Profession 1</label>
            <select id="profession-1" bind:value={professionOne}>
              <option value="">None</option>
              {#each meta?.professions ?? [] as profession (profession)}
                <option value={profession}>{label(profession)}</option>
              {/each}
            </select>
          </div>
          <div>
            <label for="profession-2">Profession 2</label>
            <select id="profession-2" bind:value={professionTwo}>
              <option value="">None</option>
              {#each meta?.professions ?? [] as profession (profession)}
                <option value={profession}>{label(profession)}</option>
              {/each}
            </select>
          </div>
        </div>

        {#if violations.length > 0}
          <ul class="validation-list" role="alert">
            {#each violations as violation, index (index)}
              <li><code>{violation.code}</code> {violation.message}</li>
            {/each}
          </ul>
        {:else if errorMessage}
          <p class="error-text" role="alert">{errorMessage}</p>
        {/if}

        <div class="button-row dialog-actions">
          <button type="submit" disabled={submitting || !meta}>
            {submitting ? 'Creating…' : 'Create'}
          </button>
          <button type="button" class="secondary-button" onclick={onclose}>Cancel</button>
        </div>
      </form>
    </dialog>
  </div>
{/if}
