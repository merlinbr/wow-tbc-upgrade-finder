<script>
  import ArmoryView from './lib/ArmoryView.svelte';
  import ImportPanel from './lib/ImportPanel.svelte';
  import RankingPanel from './lib/RankingPanel.svelte';
  import ReportView from './lib/ReportView.svelte';
  import { state as uiState } from './lib/stores.svelte.js';

  let originalLink = $state('');
  let importing = $state(false);

  function importPending() {
    importing = true;
  }

  function importComplete(link, succeeded) {
    if (succeeded) originalLink = link;
    importing = false;
  }

  function validationError(message) {
    uiState.error = { code: 'invalid_options', message };
  }

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
</script>

<svelte:head>
  <title>TBC Upgrade Finder</title>
</svelte:head>

<main>
  <header class="page-header">
    <div class="section-kicker">TBC · Upgrade intelligence</div>
    <h1>TBC Upgrade Finder</h1>
    <p>Review your imported armory, then rank practical single-item DPS upgrades under the exact simulation configuration.</p>
  </header>

  <p class="alert" role="alert" aria-live="assertive">{uiState.error?.message ?? ''}</p>
  <p class="progress-status" role="status" aria-live="polite">{progressText(uiState.job)}</p>

  <ImportPanel error={uiState.error} onImportPending={importPending} onImportComplete={importComplete} />

  {#if uiState.imported && !importing}
    <ArmoryView imported={uiState.imported} />
    <RankingPanel imported={uiState.imported} link={originalLink} job={uiState.job} onValidationError={validationError} />
  {/if}

  {#if uiState.report}
    <ReportView report={uiState.report} copyStatus={uiState.copyStatus} />
  {/if}
</main>

<footer>
  <p>
    Simulation and item data provided by
    <a href="https://github.com/wowsims/tbc-new">wowsims/tbc-new</a>.
    Rankings are local single-item DPS comparisons, not acquisition pricing or multi-item optimization.
  </p>
</footer>
