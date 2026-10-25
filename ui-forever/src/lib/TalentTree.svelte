<script>
  import TalentNode from './TalentNode.svelte';

  let {
    name,
    background = '',
    talents = [],
    total = 0,
    active = false,
    onactivate = () => {},
  } = $props();
</script>

<section class="talent-tree" class:active>
  <header class="talent-tree-header">
    <h3>{name}</h3>
    <span class="muted">{total} points</span>
  </header>
  {#if background}
    <div class="talent-tree-background" style={`background-image: url(${background})`}></div>
  {/if}
  <div class="talent-grid">
    {#each talents as entry (entry.talent.id)}
      <div
        class="talent-cell"
        style={`grid-row: ${entry.talent.row + 1}; grid-column: ${entry.talent.col + 1}`}
      >
        <TalentNode
          talent={entry.talent}
          points={entry.points}
          state={entry.state}
          {onactivate}
        />
      </div>
    {/each}
  </div>
</section>
