import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  humanizeEnum, statLabel, formatStatLine,
  qualityLabel, sourceKindLabel, humanizeSourceKind, sourceKinds, socketColors,
} from './labels.js';

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

test('statLabel humanizes snake_case stat keys with acronyms', () => {
  assert.equal(statLabel('strength'), 'Strength');
  assert.equal(statLabel('spell_hit_rating'), 'Spell Hit Rating');
  assert.equal(statLabel('mp5'), 'MP5');
  assert.equal(statLabel('armor_penetration_rating'), 'Armor Penetration Rating');
  assert.equal(statLabel('some_future_stat'), 'Some Future Stat');
});

test('formatStatLine signs and formats values', () => {
  assert.equal(formatStatLine('strength', 32), '+32 Strength');
  assert.equal(formatStatLine('armor', 1825), '+1825 Armor');
  assert.equal(formatStatLine('mp5', 7.5), '+7.5 MP5');
});

test('qualityLabel maps known qualities and flags unknown', () => {
  assert.equal(qualityLabel(4), 'Epic');
  assert.equal(qualityLabel(99), 'Unknown quality (99)');
});

test('sourceKindLabel maps known kinds and flags unknown', () => {
  assert.equal(sourceKindLabel(6), 'Heroic dungeon');
  assert.equal(sourceKindLabel(42), 'Unknown source (42)');
});

test('humanizeSourceKind maps proto names to labels', () => {
  assert.equal(humanizeSourceKind('SourceCrafting'), 'Crafting');
  assert.equal(humanizeSourceKind('SourceQuest'), 'Quest');
  assert.equal(humanizeSourceKind('SourceReputation'), 'Reputation');
  assert.equal(humanizeSourceKind('SourcePvp'), 'PvP');
  assert.equal(humanizeSourceKind('SourceDungeon'), 'Dungeon');
  assert.equal(humanizeSourceKind('SourceDungeonH'), 'Heroic dungeon');
  assert.equal(humanizeSourceKind('SourceRaid'), 'Raid');
  assert.equal(humanizeSourceKind('SourceRaidH'), 'Heroic raid');
  assert.equal(humanizeSourceKind('SourceRaidRF'), 'Raid finder');
  assert.equal(humanizeSourceKind('SourceRaidFlex'), 'Flexible raid');
  assert.equal(humanizeSourceKind('SourceSoldBy'), 'Sold by vendor');
  assert.equal(humanizeSourceKind('SourceSomethingNew'), 'Something New');
});

test('sourceKinds lists 12 kinds with Unknown first', () => {
  assert.equal(sourceKinds.length, 12);
  assert.deepEqual(sourceKinds[0], { value: 0, label: 'Unknown source', proto: 'SourceUnknown' });
  assert.equal(sourceKinds.find((k) => k.value === 7).label, 'Raid');
});

test('socketColors covers all gem color enum values', () => {
  assert.equal(socketColors[1], 'Meta');
  assert.equal(socketColors[2], 'Red');
  assert.equal(socketColors[8], 'Prismatic');
  assert.equal(Object.keys(socketColors).length, 9);
});
