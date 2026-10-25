import assert from 'node:assert/strict';
import test from 'node:test';

import { href, parse } from './router.svelte.js';

test('parses editor routes', () => {
  assert.deepEqual(parse('#/characters/abc/talents?build=p1'), {
    name: 'editor',
    id: 'abc',
    tab: 'talents',
    buildId: 'p1',
  });
  assert.deepEqual(parse('#/'), { name: 'library' });
  assert.equal(href('editor', { id: 'abc', tab: 'gear', buildId: 'p1' }), '#/characters/abc/gear?build=p1');
});

test('defaults unknown tabs and hashes', () => {
  assert.deepEqual(parse('#/characters/abc/unknown'), {
    name: 'editor',
    id: 'abc',
    tab: 'gear',
    buildId: '',
  });
  assert.deepEqual(parse('#/not-a-route'), { name: 'library' });
  assert.equal(href('library'), '#/');
});
