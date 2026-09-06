// Positions the full icon-plus-panel footprint in CSS pixels, coordinates in
// the same space as getBoundingClientRect. Keeps an 8px viewport margin,
// prefers the requested side, flips to the alternate side when the preferred
// side would clip, then clamps both axes.
export function positionTooltip(anchor, size, viewport, preferredSide = 'right') {
  const margin = 8;
  const gap = 10;
  const minX = viewport.left + margin;
  const minY = viewport.top + margin;
  const maxX = Math.max(minX, viewport.left + viewport.width - margin - size.width);
  const maxY = Math.max(minY, viewport.top + viewport.height - margin - size.height);
  const candidates = { right: anchor.right + gap, left: anchor.left - size.width - gap };
  const alternate = preferredSide === 'right' ? 'left' : 'right';
  const fits = (x) => x >= minX && x <= maxX;
  const x = fits(candidates[preferredSide]) ? candidates[preferredSide]
    : fits(candidates[alternate]) ? candidates[alternate]
    : candidates[preferredSide];
  return {
    left: Math.min(maxX, Math.max(minX, x)),
    top: Math.min(maxY, Math.max(minY, anchor.top)),
  };
}
