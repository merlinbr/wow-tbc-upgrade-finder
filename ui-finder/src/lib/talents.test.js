import assert from 'node:assert/strict';
import test from 'node:test';
import { decodeTalentsString, rankAt, treePoints } from './talents.js';

test('decodes hyphen-separated trees and pads missing trees', () => {
  assert.deepEqual(decodeTalentsString('321-5-0'), ['321', '5', '0']);
  assert.deepEqual(decodeTalentsString(''), ['', '', '']);
  assert.deepEqual(decodeTalentsString('321'), ['321', '', '']);
});

test('reads a rank at an index and treats missing digits as zero', () => {
  assert.equal(rankAt('321', 0), 3);
  assert.equal(rankAt('321', 2), 1);
  assert.equal(rankAt('12', 4), 0);
  assert.equal(rankAt('', 0), 0);
});

test('sums points across a tree', () => {
  assert.equal(treePoints('321'), 6);
  assert.equal(treePoints(''), 0);
});
