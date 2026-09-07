<script module>
  let triggerCounter = 0;
</script>

<script>
  import { getContext, onDestroy } from 'svelte';
  import { TOOLTIP_CONTEXT } from './tooltipController.js';

  let { content, preferredSide = 'right', children } = $props();

  const controller = getContext(TOOLTIP_CONTEXT);
  const owner = Symbol('tooltip-trigger');
  const id = `tooltip-${++triggerCounter}`;

  function buildRecord(target) {
    return { owner, id, content, preferredSide, anchor: target };
  }

  function onpointerenter(event) {
    if (event.pointerType === 'touch') return;
    controller.enter(buildRecord(event.currentTarget), 'pointer');
  }

  function onpointerleave() {
    controller.leave(owner, 'pointer');
  }

  function onfocus(event) {
    controller.enter(buildRecord(event.currentTarget), 'focus');
  }

  function onblur() {
    controller.leave(owner, 'focus');
  }

  function onactivate(event) {
    const target = event.currentTarget;
    target.focus();
    // Explicit activation reopens a popup dismissed while focus stayed here.
    controller.release(owner);
    controller.enter(buildRecord(target), 'focus');
  }

  // Observe content only; release in teardown does not track popup state.
  // Re-import and unmount close the panel and cancel delayed ownership.
  $effect(() => {
    const current = content;
    return () => controller.release(owner);
  });
  onDestroy(() => controller.release(owner));
</script>

{@render children?.({ id, onpointerenter, onpointerleave, onfocus, onblur, onactivate })}
