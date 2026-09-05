import { test } from 'node:test';
import assert from 'node:assert/strict';
import { humanizeEnum } from './labels.js';

test('humanizeEnum strips proto enum prefixes and splits camelCase', () => {
  assert.equal(humanizeEnum('RaceHuman', 'Race'), 'Human');
  assert.equal(humanizeEnum('ClassPaladin', 'Class'), 'Paladin');
  assert.equal(humanizeEnum('RaceNightElf', 'Race'), 'Night Elf');
  assert.equal(humanizeEnum('RetributionPaladin'), 'Retribution Paladin');
  assert.equal(humanizeEnum('ShadowPriest'), 'Shadow Priest');
});

test('humanizeEnum leaves missing and unprefixed values intact', () => {
  assert.equal(humanizeEnum('', 'Race'), '');
  assert.equal(humanizeEnum('Human', 'Race'), 'Human');
  assert.equal(humanizeEnum(null, 'Class'), '');
});
