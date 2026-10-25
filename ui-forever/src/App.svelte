<script>
  import CharacterLibrary from './lib/CharacterLibrary.svelte';
  import Editor from './lib/Editor.svelte';
  import { parse } from './lib/router.svelte.js';

  let route = $state(parse(window.location.hash));

  $effect(() => {
    const onHash = () => {
      route = parse(window.location.hash);
    };
    window.addEventListener('hashchange', onHash);
    return () => window.removeEventListener('hashchange', onHash);
  });

  $effect(() => {
    const hash = window.location.hash;
    if (hash !== '' && hash !== '#/' && route.name === 'library') {
      window.location.hash = '#/';
    }
  });
</script>

<svelte:head>
  <title>Forever Armory</title>
</svelte:head>

<main>
  {#if route.name === 'library'}
    <CharacterLibrary />
  {:else}
    <Editor id={route.id} tab={route.tab} buildId={route.buildId} />
  {/if}
</main>
