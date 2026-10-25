<script>
  import { api } from './api.js';
  import TalentTree from './TalentTree.svelte';

  let {
    character,
    build = null,
    talents = {},
    violations = [],
    onchange = () => {},
  } = $props();

  let trees = $state(null);
  let error = $state('');
  let selectedId = $state(null);
  let activeTree = $state(0);
  let resetOpen = $state(false);

  $effect(() => {
    const className = character.class;
    let cancelled = false;
    trees = null;
    selectedId = null;
    activeTree = 0;
    error = '';
    api
      .get(`/api/talents?class=${encodeURIComponent(className)}`)
      .then((data) => {
        if (!cancelled) {
          trees = data;
        }
      })
      .catch((requestError) => {
        if (!cancelled) {
          error = requestError.message ?? 'Could not load talents.';
        }
      });
    return () => {
      cancelled = true;
    };
  });

  let index = $derived.by(() => {
    const map = {};
    if (!trees?.trees) {
      return map;
    }
    trees.trees.forEach((tree, treeIndex) => {
      tree.talents.forEach((talent) => {
        map[talent.id] = { talent, treeIndex, treeName: tree.name };
      });
    });
    return map;
  });

  function pointsFor(id) {
    return Number(talents[id] ?? 0);
  }

  function capped(id) {
    const entry = index[id];
    if (!entry) {
      return 0;
    }
    return Math.max(0, Math.min(pointsFor(id), entry.talent.maxPoints));
  }

  let brokenIDs = $derived.by(() => {
    const broken = new Set();
    if (!trees) {
      return broken;
    }
    const memo = new Map();
    const visiting = new Set();
    function isBroken(id) {
      if (memo.has(id)) {
        return memo.get(id);
      }
      if (visiting.has(id)) {
        return true;
      }
      const entry = index[id];
      if (!entry) {
        return false;
      }
      const points = capped(id);
      if (points <= 0) {
        memo.set(id, false);
        return false;
      }
      visiting.add(id);
      let result = false;
      if (entry.talent.row > 0) {
        let above = 0;
        for (const [otherId, other] of Object.entries(index)) {
          if (other.treeIndex !== entry.treeIndex || other.talent.row >= entry.talent.row) {
            continue;
          }
          if (capped(otherId) > 0 && !isBroken(otherId)) {
            above += capped(otherId);
          }
        }
        if (above < 5 * entry.talent.row) {
          result = true;
        }
      }
      if (entry.talent.prereqId) {
        const prereq = index[entry.talent.prereqId];
        if (
          !prereq ||
          capped(prereq.talent.id) < prereq.talent.maxPoints ||
          isBroken(prereq.talent.id)
        ) {
          result = true;
        }
      }
      visiting.delete(id);
      memo.set(id, result);
      return result;
    }
    for (const [id, points] of Object.entries(talents)) {
      if (Number(points) > 0 && index[id] && isBroken(id)) {
        broken.add(id);
      }
    }
    for (const violation of violations) {
      if (
        violation.talentId &&
        (violation.code === 'talent_tier_locked' ||
          violation.code === 'talent_prereq_unmet' ||
          violation.code === 'talent_rank_exceeded')
      ) {
        broken.add(violation.talentId);
      }
    }
    return broken;
  });

  let budget = $derived(Math.max(0, Number(character.level) - 9));
  let totalPoints = $derived(
    Object.values(talents).reduce((sum, points) => sum + (Number(points) || 0), 0),
  );
  let unknownIDs = $derived(
    Object.keys(talents).filter((id) => !index[id] && Number(talents[id]) !== 0),
  );

  function pointsAbove(entry) {
    let above = 0;
    for (const [otherId, other] of Object.entries(index)) {
      if (other.treeIndex !== entry.treeIndex || other.talent.row >= entry.talent.row) {
        continue;
      }
      const points = pointsFor(otherId);
      if (points > 0 && !brokenIDs.has(otherId)) {
        above += Math.min(points, other.talent.maxPoints);
      }
    }
    return above;
  }

  function locked(id) {
    const entry = index[id];
    if (!entry) {
      return true;
    }
    if (entry.talent.row > 0 && pointsAbove(entry) < 5 * entry.talent.row) {
      return true;
    }
    if (entry.talent.prereqId) {
      const prereq = index[entry.talent.prereqId];
      if (!prereq) {
        return true;
      }
      if (
        pointsFor(prereq.talent.id) < prereq.talent.maxPoints ||
        brokenIDs.has(prereq.talent.id)
      ) {
        return true;
      }
    }
    return false;
  }

  function nodeState(id) {
    const entry = index[id];
    if (!entry) {
      return 'locked';
    }
    const points = pointsFor(id);
    if (brokenIDs.has(id)) {
      return 'broken';
    }
    if (points >= entry.talent.maxPoints) {
      return 'maxed';
    }
    if (points > 0) {
      return 'allocated';
    }
    return locked(id) ? 'locked' : 'available';
  }

  let treeData = $derived.by(() => {
    if (!trees?.trees) {
      return [];
    }
    return trees.trees.map((tree) => ({
      name: tree.name,
      background: tree.background,
      total: tree.talents.reduce((sum, talent) => sum + Math.max(0, pointsFor(talent.id)), 0),
      talents: tree.talents.map((talent) => ({
        talent,
        points: pointsFor(talent.id),
        state: nodeState(talent.id),
      })),
    }));
  });

  let selectedEntry = $derived(selectedId ? index[selectedId] : null);
  let selectedPoints = $derived(selectedId ? pointsFor(selectedId) : 0);
  let selectedState = $derived(selectedId ? nodeState(selectedId) : '');
  let selectedViolations = $derived(
    selectedId ? violations.filter((violation) => violation.talentId === selectedId) : [],
  );

  function lockReason(id) {
    const entry = index[id];
    if (!entry) {
      return '';
    }
    if (entry.talent.row > 0 && pointsAbove(entry) < 5 * entry.talent.row) {
      return `Requires ${5 * entry.talent.row} points in higher rows of this tree.`;
    }
    if (entry.talent.prereqId) {
      const prereq = index[entry.talent.prereqId];
      if (prereq) {
        return `Requires ${prereq.talent.name} at maximum rank.`;
      }
    }
    return 'Locked.';
  }

  let addHint = $derived.by(() => {
    if (!selectedId || !selectedEntry) {
      return '';
    }
    if (selectedPoints >= selectedEntry.talent.maxPoints) {
      return 'Maximum rank reached.';
    }
    if (totalPoints >= budget) {
      return 'Talent budget exhausted.';
    }
    if (selectedState === 'broken') {
      return 'A broken talent cannot gain points.';
    }
    if (selectedState === 'locked') {
      return lockReason(selectedId);
    }
    return '';
  });
  let canAdd = $derived(Boolean(selectedId) && addHint === '');

  function setPoints(id, next) {
    const updated = { ...talents };
    if (next <= 0) {
      delete updated[id];
    } else {
      updated[id] = next;
    }
    onchange(updated);
  }

  function addRank() {
    if (canAdd) {
      setPoints(selectedId, selectedPoints + 1);
    }
  }

  function removeRank() {
    if (selectedPoints > 0) {
      setPoints(selectedId, selectedPoints - 1);
    }
  }

  function resetTalents() {
    onchange({});
    selectedId = null;
    resetOpen = false;
  }
</script>

<section class="talent-editor">
  <header class="talent-header">
    <div>
      <h2>Talent editor</h2>
      <p class="muted">Broken talents keep their points but unlock nothing.</p>
    </div>
    <div class="talent-summary">
      <strong data-testid="point-total">{totalPoints} / {budget}</strong>
      <button
        type="button"
        class="secondary-button"
        onclick={() => (resetOpen = true)}
        disabled={unknownIDs.length > 0}
      >
        Reset talents…
      </button>
    </div>
  </header>

  {#if error}
    <p class="error-text" role="alert">{error}</p>
  {:else if !trees}
    <p class="muted">Loading talents…</p>
  {:else}
    <div class="talent-tabs" role="tablist" aria-label="Talent trees">
      {#each treeData as tree, treeIndex (tree.name)}
        <button
          type="button"
          role="tab"
          aria-selected={activeTree === treeIndex}
          onclick={() => (activeTree = treeIndex)}
        >
          {tree.name}
        </button>
      {/each}
    </div>

    <div class="talent-trees">
      {#each treeData as tree, treeIndex (tree.name)}
        <TalentTree
          name={tree.name}
          background={tree.background}
          talents={tree.talents}
          total={tree.total}
          active={activeTree === treeIndex}
          onactivate={(id) => (selectedId = id)}
        />
      {/each}
    </div>

    <div class="talent-detail panel">
      {#if selectedEntry}
        <h3>{selectedEntry.talent.name}</h3>
        <p class="muted">
          {selectedPoints}/{selectedEntry.talent.maxPoints} points · {selectedState}
        </p>
        {#if selectedViolations.length > 0}
          <ul class="validation-list" role="status">
            {#each selectedViolations as violation, index (index)}
              <li><code>{violation.code}</code> {violation.message}</li>
            {/each}
          </ul>
        {:else if selectedState === 'locked'}
          <p class="muted">{lockReason(selectedId)}</p>
        {/if}
        {#if addHint}
          <p class="muted">{addHint}</p>
        {/if}
        <div class="button-row">
          <button
            type="button"
            class="secondary-button"
            onclick={removeRank}
            disabled={selectedPoints <= 0}
          >
            − Rank
          </button>
          <button type="button" onclick={addRank} disabled={!canAdd}>+ Rank</button>
        </div>
      {:else}
        <p class="muted">Select a talent to spend points.</p>
      {/if}
    </div>

    {#if unknownIDs.length > 0}
      <div class="panel unknown-talents">
        <h3>Unknown talents</h3>
        <p class="muted">
          These saved talents are not in the pinned trees. They are retained and block reset.
        </p>
        <ul>
          {#each unknownIDs as id (id)}
            <li><code>{id}</code> · {talents[id]} points</li>
          {/each}
        </ul>
      </div>
    {/if}
  {/if}

  {#if resetOpen}
    <div class="dialog-backdrop">
      <dialog open class="panel dialog" aria-modal="true" aria-labelledby="reset-talents-title">
        <h2 id="reset-talents-title">Reset talents?</h2>
        <p>All allocated points in this build will be cleared.</p>
        <div class="button-row dialog-actions">
          <button type="button" class="danger-button" onclick={resetTalents}>Reset</button>
          <button type="button" class="secondary-button" onclick={() => (resetOpen = false)}>
            Cancel
          </button>
        </div>
      </dialog>
    </div>
  {/if}
</section>
