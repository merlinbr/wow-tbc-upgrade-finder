<script>
  import { onMount } from 'svelte';

  import CreateCharacterDialog from './CreateCharacterDialog.svelte';
  import { label } from './labels.js';
  import { href, navigate } from './router.svelte.js';
  import { loadCharacters, state as libraryState } from './stores.svelte.js';

  let dialogOpen = $state(false);

  onMount(() => {
    loadCharacters();
  });

  function openEditor(character) {
    navigate('editor', { id: character.id, tab: 'gear' });
  }
</script>

<svelte:head>
  <title>Characters · Forever Armory</title>
</svelte:head>

<section class="library">
  <header class="library-header">
    <div>
      <p class="kicker">Forever</p>
      <h1>Characters</h1>
      <p class="muted">Manual planning against the pinned Forever catalog.</p>
    </div>
    <button onclick={() => (dialogOpen = true)}>Create character</button>
  </header>

  {#if libraryState.loading}
    <p class="muted">Loading characters…</p>
  {:else if libraryState.error}
    <p class="error-text" role="alert">{libraryState.error}</p>
  {:else if libraryState.characters.length === 0}
    <p class="empty-state panel">No characters yet. Create one to start planning.</p>
  {:else}
    <ul class="character-grid">
      {#each libraryState.characters as character (character.id)}
        <li class="panel character-card">
          <div>
            <h2>{character.name}</h2>
            <p class="muted">
              Level {character.level} {label(character.race)} · {label(character.class)}
            </p>
            {#if character.professions?.length}
              <p class="muted">{character.professions.map(label).join(' · ')}</p>
            {/if}
          </div>
          <a class="reopen-link" href={href('editor', { id: character.id, tab: 'gear' })}>
            Reopen
          </a>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<CreateCharacterDialog
  open={dialogOpen}
  onclose={() => (dialogOpen = false)}
  oncreated={(character) => {
    dialogOpen = false;
    openEditor(character);
  }}
/>
