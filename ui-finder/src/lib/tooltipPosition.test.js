import { test } from 'node:test';
import assert from 'node:assert/strict';
import { positionTooltip } from './tooltipPosition.js';

test('flips at the right edge and clamps vertically', () => {
  const result = positionTooltip(
    { left: 900, right: 954, top: 650 },
    { width: 363, height: 300 },
    { left: 0, top: 0, width: 1000, height: 800 }, 'right');
  assert.deepEqual(result, { left: 527, top: 492 });
});

test('prefers the requested side when it fits', () => {
  const result = positionTooltip(
    { left: 100, right: 154, top: 100 },
    { width: 363, height: 300 },
    { left: 0, top: 0, width: 1000, height: 800 }, 'right');
  assert.deepEqual(result, { left: 164, top: 100 });
});

test('flips to the right when the left side would clip', () => {
  const result = positionTooltip(
    { left: 20, right: 74, top: 100 },
    { width: 363, height: 300 },
    { left: 0, top: 0, width: 1000, height: 800 }, 'left');
  assert.deepEqual(result, { left: 84, top: 100 });
});

test('clamps top to the viewport margin', () => {
  const result = positionTooltip(
    { left: 100, right: 154, top: 2 },
    { width: 363, height: 300 },
    { left: 0, top: 0, width: 1000, height: 800 }, 'right');
  assert.equal(result.top, 8);
});

test('respects nonzero visual-viewport offsets', () => {
  const result = positionTooltip(
    { left: 400, right: 454, top: 250 },
    { width: 300, height: 200 },
    { left: 10, top: 20, width: 500, height: 400 }, 'right');
  assert.deepEqual(result, { left: 90, top: 212 });
});

test('oversized content is clamped to the minimum corner rather than freed', () => {
  const result = positionTooltip(
    { left: 500, right: 554, top: 100 },
    { width: 1200, height: 500 },
    { left: 0, top: 0, width: 1000, height: 800 }, 'right');
  assert.deepEqual(result, { left: 8, top: 100 });
});
