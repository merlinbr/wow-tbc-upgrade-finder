<script>
  import { slotLabel } from './labels.js';

  let { violations = [] } = $props();
</script>

{#if violations.length > 0}
  <ul class="validation-list violation-list" role="alert" data-testid="violation-list">
    {#each violations as violation, index (index)}
      <li class:warning={violation.severity === 'warning'}>
        <code>{violation.code}</code>
        {#if violation.slot}
          <span class="violation-slot">({slotLabel(violation.slot)})</span>
        {/if}
        <span>{violation.message}</span>
      </li>
    {/each}
  </ul>
{/if}
