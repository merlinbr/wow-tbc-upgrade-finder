import test from 'node:test';
import assert from 'node:assert/strict';
import { classColor, classIcon, avgItemLevel } from './identity.js';

test('classColor covers all nine classes', () => {
  const classes = [
    'ClassDruid', 'ClassHunter', 'ClassMage', 'ClassPaladin', 'ClassPriest',
    'ClassRogue', 'ClassShaman', 'ClassWarlock', 'ClassWarrior',
  ];
  for (const klass of classes) assert.match(classColor(klass), /^#[0-9a-f]{6}$/i);
  assert.equal(classColor('UnknownClass'), '');
});

test('classIcon builds a ZAM CDN medium icon URL', () => {
  assert.equal(classIcon('ClassMage'), 'https://wow.zamimg.com/images/wow/icons/medium/class_mage.jpg');
  assert.equal(classIcon(''), '');
});

test('avgItemLevel skips empty slots and rounds', () => {
  const gear = [
    { itemId: 1, ilvl: 100 },
    { itemId: 0, ilvl: 0 },
    { itemId: 2, ilvl: 101 },
  ];
  assert.equal(avgItemLevel(gear), 101);
  assert.equal(avgItemLevel([]), 0);
  assert.equal(avgItemLevel([{ itemId: 0, ilvl: 0 }]), 0);
});
