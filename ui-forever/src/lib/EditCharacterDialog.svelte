<script>
  import { api } from './api.js';
  import { label } from './labels.js';

  let { open = false, character, meta = null, onclose = () => {}, onsave = () => {} } = $props();

  let name = $state('');
  let specName = $state('');
  let raceName = $state('');
  let professionOne = $state('');
  let professionTwo = $state('');
  let submitting = $state(false);
  let violations = $state([]);
  let errorMessage = $state('');
  let wasOpen = false;

  $effect(() => {
    if (open && !wasOpen && character) {
      reset();
    }
    wasOpen = open;
  });

  function reset() {
    name = character.name;
    specName = character.spec;
    raceName = character.race;
    professionOne = character.professions?.[0] ?? '';
    professionTwo = character.professions?.[1] ?? '';
    violations = [];
    errorMessage = '';
  }

  function specOptions() {
    return meta?.specs?.[character.class] ?? [];
  }

  function raceAllowed(race) {
    return (meta?.classRaces?.[character.class] ?? []).includes(race);
  }

  async function submit(event) {
    event.preventDefault();
    if (submitting) {
      return;
    }
    submitting = true;
    violations = [];
    errorMessage = '';
    try {
      const professions = [professionOne, professionTwo].filter(Boolean);
      const result = await api.patch(`/api/characters/${character.id}`, {
        name,
        spec: specName,
        race: raceName,
        professions,
      });
      onsave(result.character, result.warnings ?? []);
    } catch (error) {
      violations = error.violations ?? [];
      errorMessage = error.message ?? 'Could not save character.';
    } finally {
      submitting = false;
    }
  }
</script>

{#if open}
  <div class="dialog-backdrop">
    <dialog open class="panel dialog" aria-modal="true" aria-labelledby="edit-character-title">
      <h2 id="edit-character-title">Edit character</h2>
      <form onsubmit={submit}>
        <label for="edit-character-name">Name</label>
        <input id="edit-character-name" bind:value={name} maxlength="40" />

        <div class="field-row">
          <div>
            <label for="edit-character-spec">Specialization</label>
            <select id="edit-character-spec" bind:value={specName}>
              {#each specOptions() as spec (spec.key)}
                <option value={spec.key}>{spec.name}</option>
              {/each}
            </select>
          </div>
          <div>
            <label for="edit-character-race">Race</label>
            <select id="edit-character-race" bind:value={raceName}>
              {#each meta?.races ?? [] as race (race)}
                <option value={race} disabled={!raceAllowed(race)}>{label(race)}</option>
              {/each}
            </select>
          </div>
        </div>

        <div class="field-row">
          <div>
            <label for="edit-profession-1">Profession 1</label>
            <select id="edit-profession-1" bind:value={professionOne}>
              <option value="">None</option>
              {#each meta?.professions ?? [] as profession (profession)}
                <option value={profession}>{label(profession)}</option>
              {/each}
            </select>
          </div>
          <div>
            <label for="edit-profession-2">Profession 2</label>
            <select id="edit-profession-2" bind:value={professionTwo}>
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
          <button type="submit" disabled={submitting}>Save character</button>
          <button type="button" class="secondary-button" onclick={onclose}>Cancel</button>
        </div>
      </form>
    </dialog>
  </div>
{/if}
