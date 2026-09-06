<script>
  import { positionTooltip } from './tooltipPosition.js';
  import ItemTooltip from './ItemTooltip.svelte';

  let { active, controller } = $props();

  let layerEl = $state(null);
  let rect = $state({ left: 0, top: 0, width: 0, height: 0 });
  let ready = $state(false);
  let pendingFrame = 0;

  function cancelFrame() {
    if (pendingFrame) {
      cancelAnimationFrame(pendingFrame);
      pendingFrame = 0;
    }
  }

  // Measure while hidden, then apply the position and reveal. All
  // reposition triggers coalesce into a single animation frame.
  function measure() {
    cancelFrame();
    pendingFrame = requestAnimationFrame(() => {
      pendingFrame = 0;
      const record = active;
      if (!record || !layerEl || !record.anchor?.isConnected) return;
      const anchorRect = record.anchor.getBoundingClientRect();
      const layerRect = layerEl.getBoundingClientRect();
      const vv = window.visualViewport;
      const viewport = {
        left: vv?.offsetLeft ?? 0,
        top: vv?.offsetTop ?? 0,
        width: vv?.width ?? window.innerWidth,
        height: vv?.height ?? window.innerHeight,
      };
      const { left, top } = positionTooltip(
        anchorRect,
        { width: layerRect.width, height: layerRect.height },
        viewport,
        record.preferredSide,
      );
      rect = { left, top };
      ready = true;
    });
  }

  function onKeydown(event) {
    if (event.key === 'Escape') controller.dismiss();
  }

  $effect(() => {
    const record = active;
    if (!record) {
      ready = false;
      cancelFrame();
      return;
    }
    if (!record.anchor?.isConnected) {
      ready = false;
      controller.release(record.owner);
      return;
    }
    ready = false;
    measure();
    const anchorObserver = new ResizeObserver(measure);
    anchorObserver.observe(record.anchor);
    const layerObserver = layerEl ? new ResizeObserver(measure) : null;
    layerObserver?.observe(layerEl);
    window.addEventListener('resize', measure);
    window.addEventListener('scroll', measure, { capture: true, passive: true });
    window.addEventListener('keydown', onKeydown);
    window.visualViewport?.addEventListener('scroll', measure);
    window.visualViewport?.addEventListener('resize', measure);
    return () => {
      cancelFrame();
      anchorObserver.disconnect();
      layerObserver?.disconnect();
      window.removeEventListener('resize', measure);
      window.removeEventListener('scroll', measure, { capture: true });
      window.removeEventListener('keydown', onKeydown);
      window.visualViewport?.removeEventListener('scroll', measure);
      window.visualViewport?.removeEventListener('resize', measure);
    };
  });
</script>

{#if active && active.anchor?.isConnected}
  <div
    class="tooltip-layer"
    class:visible={ready}
    style:left="{rect.left}px"
    style:top="{rect.top}px"
    bind:this={layerEl}
    role="presentation"
    onpointerenter={() => controller.enterPanel(active.owner)}
    onpointerleave={() => controller.leavePanel(active.owner)}
  >
    {#key active.owner}
      <ItemTooltip item={active.item} variant={active.variant} id={active.id} />
    {/key}
  </div>
{/if}

<style>
  .tooltip-layer {
    left: 0;
    position: fixed;
    top: 0;
    visibility: hidden;
    z-index: 1000;
  }
  .tooltip-layer.visible {
    visibility: visible;
  }
</style>
