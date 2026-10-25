const TABS = new Set(['gear', 'stats', 'talents']);

/** Parse a location hash into a route object. Unknown hashes become the library. */
export function parse(hash) {
  const raw = String(hash ?? '').replace(/^#/, '');
  const [path, query = ''] = raw.split('?');
  const segments = path.split('/').filter(Boolean);

  if (segments.length === 0) {
    return { name: 'library' };
  }
  if (segments[0] === 'characters' && segments.length >= 2) {
    const tab = TABS.has(segments[2]) ? segments[2] : 'gear';
    const buildId = new URLSearchParams(query).get('build') ?? '';
    return {
      name: 'editor',
      id: decodeURIComponent(segments[1]),
      tab,
      buildId,
    };
  }
  return { name: 'library' };
}

/** Build a location hash for a route. */
export function href(name, params = {}) {
  if (name !== 'editor') {
    return '#/';
  }
  const tab = TABS.has(params.tab) ? params.tab : 'gear';
  const base = `#/characters/${encodeURIComponent(params.id ?? '')}/${tab}`;
  return params.buildId ? `${base}?build=${encodeURIComponent(params.buildId)}` : base;
}

/** Navigate to a route. */
export function navigate(name, params = {}) {
  window.location.hash = href(name, params).slice(1);
}
