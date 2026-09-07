<script>
  let { talent, id } = $props();
  let allocation = $derived(talent.points === 0 ? 'Unallocated'
    : talent.points === talent.maxPoints ? 'Maximum rank' : 'Allocated');
</script>

<div class="talent-tooltip" role="tooltip" {id}>
  <strong>{talent.name}</strong>
  <div>{talent.treeName}</div>
  <div>Rank {talent.points} / {talent.maxPoints}</div>
  <div>{allocation} · Imported build · Read-only</div>
</div>

<style>
  .talent-tooltip {
    background: #080d16;
    border: 1px solid #85848b;
    border-radius: 4px;
    color: #e8edf5;
    font-size: 0.85rem;
    max-width: min(20rem, calc(100vw - 1rem));
    overflow-wrap: anywhere;
    padding: 0.65rem 0.8rem;
  }
  strong { color: #c7ef31; }
</style>
