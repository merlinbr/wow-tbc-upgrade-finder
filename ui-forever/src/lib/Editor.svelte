<script>
  import { api } from './api.js';

  let { id } = $props();
  let character = $state(null);
  let error = $state('');

  $effect(() => {
    const requestedId = id;
    let cancelled = false;
    character = null;
    error = '';
    api
      .get(`/api/characters/${encodeURIComponent(requestedId)}`)
      .then((data) => {
        if (!cancelled) {
          character = data;
        }
      })
      .catch((requestError) => {
        if (!cancelled) {
          error = requestError.message ?? 'Could not load character.';
        }
      });
    return () => {
      cancelled = true;
    };
  });
</script>

{#if error}
  <p class="error-text" role="alert">{error}</p>
{:else if character}
  <h1>{character.name}</h1>
{:else}
  <p class="muted">Loading character…</p>
{/if}
