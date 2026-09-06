// Picks the viewer adapter: the real ZAM adapter by default, or the network-
// free fake (test/CI) chosen via window.__VISUAL_PROVIDER__.
export async function createViewerAdapter() {
  const kind = window.__VISUAL_PROVIDER__ || 'zam';
  if (kind === 'fake') {
    return (await import('./characterViewerAdapterFake.js')).fakeAdapter;
  }
  return (await import('./characterViewerAdapter.js')).zamAdapter;
}
