import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  appearanceOptions,
  buildPreviewItems,
  positionSlot,
  zamModelId,
  zamRaceId,
} from './characterPreview.js';

test('zamRaceId maps simulator race strings to ZAM race ids', () => {
  assert.equal(zamRaceId('RaceHuman'), 1);
  assert.equal(zamRaceId('Human'), 1);
  assert.equal(zamRaceId('RaceTroll'), 8);
  assert.equal(zamRaceId('RaceBloodElf'), 10);
  assert.equal(zamRaceId('RaceDraenei'), 11);
  assert.equal(zamRaceId('RaceUnknown'), null);
  assert.equal(zamRaceId(''), null);
  assert.equal(zamRaceId(null), null);
});

test('zamModelId follows race*2-1+gender', () => {
  assert.equal(zamModelId(1, 0), 1); // human male
  assert.equal(zamModelId(1, 1), 2); // human female
  assert.equal(zamModelId(8, 1), 16); // troll female
  assert.equal(zamModelId(10, 1), 20); // blood elf female
});

test('positionSlot overrides only weapon positions', () => {
  assert.equal(positionSlot('Main Hand'), 21);
  assert.equal(positionSlot('Off Hand'), 22);
  assert.equal(positionSlot('Ranged'), 15);
  assert.equal(positionSlot('Head'), null);
  assert.equal(positionSlot('Chest'), null);
});

const gear = [
  { slotName: 'Head', itemId: 32461 },
  { slotName: 'Neck', itemId: 30022 },
  { slotName: 'Shoulder', itemId: 30055 },
  { slotName: 'Back', itemId: 24259 },
  { slotName: 'Chest', itemId: 30129 },
  { slotName: 'Main Hand', itemId: 28430 },
  { slotName: 'Off Hand', itemId: 0 }, // empty
  { slotName: 'Hands', itemId: 29947 },
  { slotName: 'Finger 1', itemId: 30061 },
  { slotName: 'Trinket 2', itemId: 29383 },
  { slotName: 'Ranged', itemId: 27484 }, // relic, no display
];

test('buildPreviewItems maps slots, skips hidden/empty, overrides weapon slots', () => {
  const resolved = {
    '32461': { displayId: 45779, zamSlot: 1 },
    '30055': { displayId: 46353, zamSlot: 3 },
    '24259': { displayId: 38976, zamSlot: 16 },
    '30129': { displayId: 42306, zamSlot: 5 },
    '28430': { displayId: 39571, zamSlot: 17 }, // 2H → main hand position 21
    '29947': { displayId: 43549, zamSlot: 10 },
  };
  const { items, unresolved } = buildPreviewItems(gear, resolved);
  assert.deepEqual(items, [
    [1, 45779],
    [3, 46353],
    [16, 38976],
    [5, 42306],
    [21, 39571],
    [10, 43549],
  ]);
  assert.deepEqual(unresolved, ['Ranged']);
});

test('buildPreviewItems normalizes off-hand held items to slot 22', () => {
  const mageGear = [
    { slotName: 'Main Hand', itemId: 28770 },
    { slotName: 'Off Hand', itemId: 29272 },
    { slotName: 'Ranged', itemId: 28673 },
  ];
  const resolved = {
    '28770': { displayId: 43098, zamSlot: 21 },
    '29272': { displayId: 42626, zamSlot: 23 }, // held-in-off-hand
    '28673': { displayId: 43916, zamSlot: 15 },
  };
  const { items, unresolved } = buildPreviewItems(mageGear, resolved);
  assert.deepEqual(items, [
    [21, 43098],
    [22, 42626],
    [15, 43916],
  ]);
  assert.deepEqual(unresolved, []);
});

test('buildPreviewItems keeps resolver chest slot (robe probe 5 vs 20)', () => {
  const robeGear = [{ slotName: 'Chest', itemId: 29077 }];
  const { items, unresolved } = buildPreviewItems(robeGear, {
    '29077': { displayId: 40468, zamSlot: 20 },
  });
  assert.deepEqual(items, [[20, 40468]]);
  assert.deepEqual(unresolved, []);
});

const customization = {
  Options: [
    { Id: 14, Name: 'Skin Color', Choices: [{ Id: 17215 }, { Id: 17216 }, { Id: 17217 }] },
    { Id: 15, Name: 'Face', Choices: [{ Id: 5 }, { Id: 6 }] },
    { Id: 99, Name: 'Fur Color', Choices: [{ Id: 7 }] }, // no mapping entry
  ],
};

test('appearanceOptions maps indexes to choices and defaults to first', () => {
  const options = appearanceOptions({ skin: 2, face: 1 }, customization);
  assert.deepEqual(options, [
    { optionId: 15, choiceId: 6 },
    { optionId: 14, choiceId: 17217 },
  ]);
});

test('appearanceOptions uses first choice when appearance omitted', () => {
  const options = appearanceOptions({}, customization);
  assert.deepEqual(options, [
    { optionId: 15, choiceId: 5 },
    { optionId: 14, choiceId: 17215 },
  ]);
});
