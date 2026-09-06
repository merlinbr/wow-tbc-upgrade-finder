<script>
  import { cancelRanking, startRanking } from './stores.svelte.js';
  import { sourceKinds } from './labels.js';

  let { imported, link, job, onValidationError = () => {} } = $props();
  let maxPhase = $state(0);
  let includeUnknown = $state(false);
  let screeningIterations = $state(300);
  let confirmationIterations = $state(1000);
  let selectedKinds = $state([]);
  let initializedImport = $state(null);
  let submitting = $state(false);

  $effect(() => {
    if (!imported || imported === initializedImport) return;
    initializedImport = imported;
    const defaults = imported.defaults ?? {};
    maxPhase = Number(defaults.maxPhase ?? imported.character?.phase ?? 5);
    includeUnknown = Boolean(defaults.includeUnknown);
    screeningIterations = Number(defaults.screeningIterations ?? 300);
    confirmationIterations = Number(defaults.confirmationIterations ?? 1000);
    selectedKinds = [];
  });

  function progressText(job) {
    if (!job) return '';
    if (job.status === 'canceled') return 'Ranking canceled.';
    if (job.status !== 'queued' && job.status !== 'running') return '';
    const progress = job.progress;
    if (progress?.total > 0) {
      return `${job.status} — ${progress.stage || 'ranking'}: ${progress.completed} / ${progress.total} candidate runs`;
    }
    return `${job.status}…`;
  }

  let busy = $derived(submitting || job?.status === 'queued' || job?.status === 'running');
  let hasJob = $derived(job?.status === 'queued' || job?.status === 'running');

  async function submitRanking(event) {
    event.preventDefault();
    const screening = Number(screeningIterations);
    const confirmation = Number(confirmationIterations);
    if (!(screening > 0) || !(confirmation >= screening)) {
      onValidationError('Screening iterations must be > 0 and confirmation iterations must be >= screening.');
      return;
    }

    submitting = true;
    await startRanking(link, {
      filters: {
        maxPhase: Number(maxPhase),
        sourceKinds: [...selectedKinds],
        sourceNames: [],
        includeUnknown,
      },
      policy: {
        gemBySocket: {},
        maxGemQuality: 4,
        enchantByType: {},
      },
      options: {
        screeningIterations: screening,
        confirmationIterations: confirmation,
      },
    });
    submitting = false;
  }
</script>

<section class="panel ranking-panel" aria-labelledby="ranking-heading" data-region="ranking-panel">
  <div class="section-kicker">Server comparison</div>
  <h2 id="ranking-heading">Find upgrades</h2>
  <p class="panel-intro">The local server screens and confirms compatible single-item DPS upgrades using the imported simulation settings.</p>
  <form onsubmit={submitRanking}>
    <div class="control-grid">
      <div>
        <label for="max-phase">Maximum phase</label>
        <input id="max-phase" name="maxPhase" type="number" min="1" step="1" bind:value={maxPhase} disabled={busy} />
      </div>
      <div class="check-control">
        <label>
          <input id="include-unknown" name="includeUnknown" type="checkbox" bind:checked={includeUnknown} disabled={busy} />
          Include unknown-source items
        </label>
      </div>
      <div>
        <label for="screening-iterations">Screening iterations</label>
        <input id="screening-iterations" name="screeningIterations" type="number" min="0" step="1" bind:value={screeningIterations} disabled={busy} />
      </div>
      <div>
        <label for="confirmation-iterations">Confirmation iterations</label>
        <input id="confirmation-iterations" name="confirmationIterations" type="number" min="0" step="1" bind:value={confirmationIterations} disabled={busy} />
      </div>
    </div>
      <fieldset class="check-control source-kind-group" disabled={busy}>
        <legend>Include sources</legend>
        {#each sourceKinds.filter((kind) => kind.value !== 0) as kind}
          <label>
            <input
              type="checkbox"
              checked={selectedKinds.includes(kind.value)}
              onchange={(event) => {
                selectedKinds = event.currentTarget.checked
                  ? [...selectedKinds, kind.value]
                  : selectedKinds.filter((v) => v !== kind.value);
              }}
            />
            {kind.label}
          </label>
        {/each}
      </fieldset>
    <p class="progress-status" role="status" aria-live="polite">{progressText(job)}</p>
    <div class="button-row">
      <button id="rank-button" type="submit" disabled={busy} data-action="start-ranking">{busy ? 'Ranking…' : 'Start ranking'}</button>
      {#if hasJob}
        <button id="cancel-button" type="button" class="secondary-button" onclick={cancelRanking} data-action="cancel-ranking">Cancel ranking</button>
      {/if}
</div>
  </form>
</section>
