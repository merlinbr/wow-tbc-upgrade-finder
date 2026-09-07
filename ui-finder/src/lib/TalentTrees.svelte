<script>
  import { decodeTalentsString, rankAt, treePoints } from './talents.js';
  import talentIcons from './talentIcons.json';
  import TooltipIcon from './TooltipIcon.svelte';
  import TooltipTrigger from './TooltipTrigger.svelte';
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
    return {
      ...tree,
      ranks,
      points: treePoints(ranks),
      rows: Math.max(...tree.talents.map((talent) => talent.location.rowIdx), 0) + 1,
      cols: Math.max(...tree.talents.map((talent) => talent.location.colIdx), 0) + 1,
      headerIcon: talentIcons[tree.talents.at(-1)?.spellIds[0]] ?? '',
    };
  }));
</script>

{#if treeData.length}
  <div class="talent-view">
    <div class="talent-build-summary">
      <strong>Imported build</strong>
      <span class="talent-readonly">Read-only</span>
      <span role="group" aria-label={treeData.map((tree) => `${tree.name}: ${tree.points} points`).join(', ')}>
        {treeData.map((tree) => tree.points).join(' / ')}
      </span>
    </div>
    <div class="talent-trees" data-region="talent-trees">
      {#each treeData as tree}
        <section
          class="talent-tree"
          style="--talent-cols: {tree.cols}; --talent-rows: {tree.rows}"
          aria-label="{tree.name} talents"
        >
          <header class="talent-tree-header">
            <span aria-hidden="true"><TooltipIcon icon={tree.headerIcon} name={tree.name} size={32} /></span>
            <h3>{tree.name}</h3>
            <span class="talent-tree-points">{tree.points} points</span>
          </header>
          <div
            class="talent-tree-body"
            style:background-image={tree.backgroundUrl
              ? `linear-gradient(rgba(5, 8, 12, .18), rgba(5, 8, 12, .40)), url('${tree.backgroundUrl}'), linear-gradient(#182130, #080d16)`
              : 'linear-gradient(#182130, #080d16)'}
          >
            {#each tree.talents as talent, index (talent.spellIds[0])}
              {@const points = rankAt(tree.ranks, index)}
              <TooltipTrigger
                content={{ kind: 'talent', talent: { name: talent.fancyName, treeName: tree.name, points, maxPoints: talent.maxPoints } }}
              >
                {#snippet children({ id, onpointerenter, onpointerleave, onfocus, onblur, onactivate })}
                  <button
                    type="button"
                    class="talent-node"
                    class:unallocated={points === 0}
                    class:allocated={points > 0}
                    class:maxed={points === talent.maxPoints}
                    style:grid-row={talent.location.rowIdx + 1}
                    style:grid-column={talent.location.colIdx + 1}
                    aria-label={`Inspect ${talent.fancyName}, rank ${points} of ${talent.maxPoints}`}
                    aria-describedby={id}
                    {onpointerenter}
                    {onpointerleave}
                    {onfocus}
                    {onblur}
                    onclick={onactivate}
                  >
                    <span class="talent-node-art" aria-hidden="true">
                      <TooltipIcon icon={talentIcons[talent.spellIds[0]] ?? ''} name={talent.fancyName} size={40} />
                    </span>
                    <span class="talent-rank" aria-hidden="true">{points}/{talent.maxPoints}</span>
                  </button>
                {/snippet}
              </TooltipTrigger>
            {/each}
          </div>
        </section>
      {/each}
    </div>
  </div>
{:else}
  <p class="muted">No talent data for this class.</p>
{/if}
