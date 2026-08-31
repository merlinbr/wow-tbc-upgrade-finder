import { deleteJSON, getJSON, postJSON } from './api.js';
let actionGeneration = 0;

export const state = $state({
  imported: null,
  job: null,
  report: null,
  error: null,
  pollTimer: null,
  copyStatus: '',
});

export async function importLink(link) {
  const generation = ++actionGeneration;
  const previousJob = state.job;
  const previousReport = state.report;
  state.error = null;
  state.copyStatus = '';
  state.report = null;

  const oldJobID = previousJob?.id;
  if (state.pollTimer !== null) {
    clearTimeout(state.pollTimer);
    state.pollTimer = null;
  }
  if (oldJobID) {
    state.job = { ...previousJob, status: 'canceled', report: null };
    try {
      await deleteJSON(`/api/jobs/${encodeURIComponent(oldJobID)}`);
    } catch (error) {
      if (generation === actionGeneration && state.job?.id === oldJobID) {
        state.job = previousJob;
        state.report = previousReport;
        state.error = error;
        pollJob();
      }
      return null;
    }
  }

  if (generation !== actionGeneration) return null;
  try {
    const imported = await postJSON('/api/import', { link });
    if (generation !== actionGeneration) return null;
    state.imported = imported;
    state.error = null;
    if (!oldJobID || state.job?.id === oldJobID) state.job = null;
    state.report = null;
    return imported;
  } catch (error) {
    if (generation === actionGeneration) state.error = error;
    return null;
  }
}

export async function startRanking(link, input) {
  const generation = ++actionGeneration;
  state.error = null;
  try {
    const job = await postJSON('/api/jobs', { ...input, link });
    if (generation !== actionGeneration) {
      if (job?.id) {
        try {
          await deleteJSON(`/api/jobs/${encodeURIComponent(job.id)}`);
        } catch {
          // Best effort cleanup for a job that can no longer be displayed.
        }
      }
      return null;
    }
    state.report = null;
    state.copyStatus = '';
    state.job = job;
    pollJob();
    return job;
  } catch (error) {
    if (generation === actionGeneration) state.error = error;
    return null;
  }
}

function isActiveJob(job) {
  return job?.status === 'queued' || job?.status === 'running';
}

function pollJob() {
  const jobID = state.job?.id;
  if (!jobID) return;

  state.pollTimer = setTimeout(async () => {
    state.pollTimer = null;
    try {
      const job = await getJSON(`/api/jobs/${encodeURIComponent(jobID)}`);
      if (state.job?.id !== jobID || state.job?.status === 'canceled') return;
      state.job = job;
      if (isActiveJob(job)) {
        pollJob();
        return;
      }
      if (job.status === 'completed') {
        state.report = job.report;
      } else {
        state.report = null;
        if (job.status === 'failed') state.error = job.error;
      }
    } catch (error) {
      if (state.job?.id === jobID) {
        state.pollTimer = null;
        state.report = null;
        state.error = error;
      }
    }
  }, 500);
}

export async function cancelRanking() {
  const generation = ++actionGeneration;
  const previousJob = state.job;
  const previousReport = state.report;
  state.report = null;
  if (state.pollTimer !== null) {
    clearTimeout(state.pollTimer);
    state.pollTimer = null;
  }

  const jobID = previousJob?.id;
  if (!jobID) return;

  state.error = null;
  state.job = { ...previousJob, status: 'canceled', report: null };
  try {
    await deleteJSON(`/api/jobs/${encodeURIComponent(jobID)}`);
    if (generation === actionGeneration && state.job?.id === jobID) {
      state.job = { ...state.job, status: 'canceled', report: null };
    }
  } catch (error) {
    if (generation === actionGeneration && state.job?.id === jobID) {
      state.job = previousJob;
      state.report = previousReport;
      state.error = error;
      pollJob();
    }
  }
}

export async function copyReport() {
  if (!state.report) return;
  try {
    await navigator.clipboard.writeText(JSON.stringify(state.report, null, 2));
    state.copyStatus = 'Report copied to clipboard.';
  } catch (error) {
    state.copyStatus = `Copy failed: ${error}`;
  }
}

