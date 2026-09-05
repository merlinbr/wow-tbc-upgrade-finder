<script>
  import { decodeTalentsString, rankAt, treePoints } from './talents.js';
  import druidTrees from '../../../ui/core/talents/trees/druid.json';
  import hunterTrees from '../../../ui/core/talents/trees/hunter.json';
  import mageTrees from '../../../ui/core/talents/trees/mage.json';
  import paladinTrees from '../../../ui/core/talents/trees/paladin.json';
  import priestTrees from '../../../ui/core/talents/trees/priest.json';
  import rogueTrees from '../../../ui/core/talents/trees/rogue.json';
  import shamanTrees from '../../../ui/core/talents/trees/shaman.json';
  import warlockTrees from '../../../ui/core/talents/trees/warlock.json';
  import warriorTrees from '../../../ui/core/talents/trees/warrior.json';

  const treesByClass = {
    ClassDruid: druidTrees,
    ClassHunter: hunterTrees,
    ClassMage: mageTrees,
    ClassPaladin: paladinTrees,
    ClassPriest: priestTrees,
    ClassRogue: rogueTrees,
    ClassShaman: shamanTrees,
    ClassWarlock: warlockTrees,
    ClassWarrior: warriorTrees,
  };

  let { class: playerClass = '', talentsString = '' } = $props();

  let ranksByTree = $derived(decodeTalentsString(talentsString, 3));

  let treeData = $derived((treesByClass[playerClass] ?? []).map((tree, treeIndex) => {
    const ranks = ranksByTree[treeIndex] ?? '';
    const maxRow = Math.max(...tree.talents.map((talent) => talent.location.rowIdx), 0);
    const maxCol = Math.max(...tree.talents.map((talent) => talent.location.colIdx), 0);
    const byLocation = new Map(
      tree.talents.map((talent, index) => [
        `${talent.location.rowIdx}:${talent.location.colIdx}`,
        { ...talent, index, points: rankAt(ranks, index) },
      ]),
    );
    const cells = [];
    for (let row = 0; row <= maxRow; row++) {
      for (let col = 0; col <= maxCol; col++) {
        cells.push(byLocation.get(`${row}:${col}`) ?? null);
      }
    }
    return {
      name: tree.name,
      backgroundUrl: tree.backgroundUrl,
      points: treePoints(ranks),
      cols: maxCol + 1,
      cells,
    };
  }));
</script>

{#if treeData.length}
  <div class="talent-trees" data-region="talent-trees">
    {#each treeData as tree}
      <section class="talent-tree" style="--talent-cols: {tree.cols}" aria-label="{tree.name} talents">
        <div class="talent-tree-header">
          <h3>{tree.name}</h3>
          <span class="talent-tree-points">{tree.points} points</span>
        </div>
        <div class="talent-tree-body" style:background-image={tree.backgroundUrl ? `url('${tree.backgroundUrl}')` : 'none'}>
          {#each tree.cells as cell, cellIndex (cell?.index ?? `empty-${cellIndex}`)}
            {#if cell}
              <div class="talent-cell" class:empty-rank={cell.points === 0} title="{cell.fancyName} {cell.points}/{cell.maxPoints}">
                <span class="talent-name">{cell.fancyName}</span>
                <span class="rank-pips" aria-label="{cell.points} of {cell.maxPoints} points">
                  {#each Array(cell.maxPoints) as _, pipIndex}
                    <span class:filled={pipIndex < cell.points} class="rank-pip" aria-hidden="true"></span>
                  {/each}
                </span>
              </div>
            {:else}
              <div class="talent-cell talent-cell-empty" aria-hidden="true"></div>
            {/if}
          {/each}
        </div>
      </section>
    {/each}
  </div>
{:else}
  <p class="muted">No talent data for this class.</p>
{/if}
