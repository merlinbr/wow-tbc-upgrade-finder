<script>
  import { untrack } from 'svelte';

  // Owns the viewer lifetime: creates on mount/config change, resizes with the
  // container, destroys on unmount and on stale asynchronous completion.
  let { config, adapter, onReady, onError } = $props();

  let mountEl = $state(null);
  let handle = $state(null);

  $effect(() => {
    const current = config;
    const currentAdapter = adapter;
    if (!mountEl || !current) return;
    let cancelled = false;
    let created = null;

    currentAdapter
      .create({ container: mountEl, ...current })
      .then((viewer) => {
        if (cancelled) {
          viewer.destroy();
          return;
        }
        created = viewer;
        handle = viewer;
        untrack(() => onReady?.(viewer));
      })
      .catch((error) => {
        if (!cancelled) untrack(() => onError?.(error));
      });

    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        created?.resize(entry.contentRect.width, entry.contentRect.height);
      }
    });
    observer.observe(mountEl);
    // The mount can start at an interim layout size (activation changes the
    // stage's aspect/height); re-check once after it settles in case the
    // observer's last entry preceded it.
    const settleTimer = setTimeout(() => {
      created?.resize(mountEl.clientWidth, mountEl.clientHeight);
    }, 600);

    return () => {
      cancelled = true;
      handle = null;
      clearTimeout(settleTimer);
      observer.disconnect();
      created?.destroy();
      created = null;
    };
  });
</script>

<div class="stage-viewer" bind:this={mountEl} aria-label="Equipped 3D character preview"></div>
