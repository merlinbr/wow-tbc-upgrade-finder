import { test } from 'node:test';
import assert from 'node:assert/strict';
import { buildItemTooltip } from './itemTooltip.js';

test('orders item stats without folding in gems or enchants', () => {
  const input = {
    itemName: 'Chest fixture', slotName: 'Chest', armorType: 4,
    stats: { melee_crit_rating: 31, intellect: 31, strength: 56, armor: 1825, stamina: 48 },
    randomSuffix: { name: 'of Strength', stats: { strength: 2 } },
    enchant: { description: '+6 All Stats', stats: { strength: 6 } },
    sockets: [{ color: 2, gem: { id: 1, name: 'Gem fixture', icon: 'gem_fixture', stats: { strength: 10 } } }],
    socketBonus: { stats: { strength: 4 }, active: false },
  };
  const before = structuredClone(input);
  const result = buildItemTooltip(input);
  assert.deepEqual(result.baseLines.map((line) => line.text),
    ['1825 Armor', '+58 Strength', '+48 Stamina', '+31 Intellect']);
  assert.equal(result.typeLabel, 'Plate');
  assert.equal(result.enchantLine, '+6 All Stats');
  assert.equal(result.sockets[0].gem.icon, 'gem_fixture');
  assert.equal(result.socketBonus.active, false);
  assert.deepEqual(result.equipLines.map((line) => line.text),
    ['Equip: Improves melee critical strike rating by 31.']);
  assert.deepEqual(input, before);
});

test('summary never invents full-item data', () => {
  const vm = buildItemTooltip({ name: 'Candidate', quality: 4, phase: 3 }, 'summary');
  assert.equal(vm.name, 'Candidate');
  assert.equal(vm.qualityLabel, 'Epic');
  assert.deepEqual(vm.baseLines, []);
  assert.deepEqual(vm.weaponLines, []);
  assert.deepEqual(vm.sockets, []);
  assert.equal(vm.socketBonus, null);
  assert.equal(vm.enchantLine, '');
  assert.deepEqual(vm.restrictionLines, []);
  assert.deepEqual(vm.equipLines, []);
});

test('empty sockets keep their color label', () => {
  const vm = buildItemTooltip({
    itemName: 'Bracers', slots: undefined,
    sockets: [{ color: 2, gem: null }, { color: 4, gem: null }],
  });
  assert.deepEqual(vm.sockets.map((socket) => socket.text),
    ['Red socket (empty)', 'Yellow socket (empty)']);
  assert.equal(vm.socketBonus, null);
});

test('gem with no numeric stats falls back to its name', () => {
  const vm = buildItemTooltip({
    itemName: 'Chest', sockets: [{ color: 3, gem: { id: 7, name: 'Gem of Silence', stats: {} } }],
  });
  assert.equal(vm.sockets[0].text, 'Gem of Silence');
  assert.equal(vm.sockets[0].gem.id, 7);
});

test('unknown stat keys stay visible as white lines in lexical order', () => {
  const vm = buildItemTooltip({ itemName: 'Odd', stats: { future_stat: 4, strength: 1, another_future: 2 } });
  assert.deepEqual(vm.baseLines.map((line) => line.text), ['+1 Strength', '+2 Another Future', '+4 Future Stat']);
  assert.equal(vm.typeLabel, '');
});

test('unknown nonzero types and classes get explicit labels', () => {
  const vm = buildItemTooltip({ itemName: 'Odd', armorType: 99, unique: true, classAllowlist: [2, 99], requiredProfession: 4 });
  assert.equal(vm.typeLabel, 'Unknown type (99)');
  assert.deepEqual(vm.restrictionLines, ['Unique', 'Classes: Paladin, Unknown class (99)', 'Requires Engineering']);
});

test('negative stats stay signed, zero and non-finite values are dropped', () => {
  const vm = buildItemTooltip({ itemName: 'Odd', stats: { strength: -5, stamina: 0, intellect: Number.POSITIVE_INFINITY } });
  assert.deepEqual(vm.baseLines.map((line) => line.text), ['-5 Strength']);
});

test('weapon items expose hand type, type, damage and speed', () => {
  const vm = buildItemTooltip({
    itemName: 'Sword', slotName: 'Main Hand', handType: 4, weaponType: 9,
    weaponDamageMin: 100, weaponDamageMax: 200, weaponSpeed: 3.5,
  });
  assert.equal(vm.slotLabel, 'Two-Hand');
  assert.equal(vm.typeLabel, 'Sword');
  assert.deepEqual(vm.weaponLines, ['100 - 200 Damage', 'Speed 3.5']);
});

test('invalid weapon damage range shows no weapon lines', () => {
  const vm = buildItemTooltip({
    itemName: 'Sword', weaponDamageMin: 0, weaponDamageMax: 0, weaponSpeed: 3.5,
  });
  assert.deepEqual(vm.weaponLines, []);
});

test('ranged type takes precedence over weapon type', () => {
  const vm = buildItemTooltip({
    itemName: 'Bow', slotName: 'Ranged', rangedWeaponType: 1, weaponType: 9, handType: 3,
  });
  assert.equal(vm.slotLabel, 'Off Hand');
  assert.equal(vm.typeLabel, 'Bow');
});

test('suffix label is retained and suffix stats counted once', () => {
  const vm = buildItemTooltip({
    itemName: 'Cloak', randomSuffix: { name: 'of Strength', stats: { strength: 2, armor: 1 } },
    stats: { strength: 56, armor: 1825 },
  });
  assert.equal(vm.suffixLabel, 'of Strength');
  assert.deepEqual(vm.baseLines.map((line) => line.text), ['1826 Armor', '+58 Strength']);
});

test('socket bonus uses only its own stats', () => {
  const vm = buildItemTooltip({
    itemName: 'Chest', stats: { strength: 56 },
    socketBonus: { stats: { spell_damage: 5, healing_power: 5 }, active: true },
  });
  assert.equal(vm.socketBonus.text, '+5 Healing Power, +5 Spell Damage');
  assert.equal(vm.socketBonus.active, true);
  assert.deepEqual(vm.baseLines.map((line) => line.text), ['+56 Strength']);
});

test('empty socketBonus stats produce no bonus line', () => {
  const vm = buildItemTooltip({ itemName: 'Chest', socketBonus: { stats: {}, active: false } });
  assert.equal(vm.socketBonus, null);
});

test('signed equip lines handle negative combat ratings', () => {
  const vm = buildItemTooltip({ itemName: 'Odd', stats: { melee_crit_rating: -5 } });
  assert.deepEqual(vm.equipLines.map((line) => line.text), ['Equip: -5 Melee Crit Rating']);
});
