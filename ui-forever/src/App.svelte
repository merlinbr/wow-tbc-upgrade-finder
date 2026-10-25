<script>
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
    <h1>Library</h1>
  {:else}
    <h1>Editor</h1>
  {/if}
</main>
