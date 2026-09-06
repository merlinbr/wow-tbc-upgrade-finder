<script module>
  let triggerCounter = 0;
</script>

<script>
  import { getContext, onDestroy } from 'svelte';
  import { ITEM_TOOLTIP_CONTEXT } from './tooltipController.js';

  let { item, variant = 'full', preferredSide = 'right', children } = $props();

  const controller = getContext(ITEM_TOOLTIP_CONTEXT);
  const owner = Symbol('item-tooltip-trigger');
  const id = `item-tooltip-${++triggerCounter}`;

  function buildRecord(target) {
    return { owner, id, item, variant, preferredSide, anchor: target };
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

  // Close the associated panel and cancel its timers when the item object
  // changes (re-import) or the trigger unmounts.
  $effect(() => {
    const current = item;
    return () => controller.release(owner);
  });
  onDestroy(() => controller.release(owner));
</script>

{@render children?.({ id, onpointerenter, onpointerleave, onfocus, onblur })}
