// Pure mapping of imported race/gear into ZAM viewer inputs.
// Keeps simulator enums distinct from ZAM race IDs, inventory slots and
// display IDs. No DOM, no network.

// ZAM race IDs for TBC. The simulator's race strings are a different domain
// (e.g. RaceHuman) and must not be forwarded numerically.
export const ZAM_RACES = {
  Human: 1,
  Orc: 2,
  Dwarf: 3,
  NightElf: 4,
  Undead: 5,
  Tauren: 6,
  Gnome: 7,
  Troll: 8,
  BloodElf: 10,
  Draenei: 11,
};

// ZAM viewer display slots by equipped position. Weapons are positional:
// the viewer ignores the raw inventory type of a weapon (1H 13, 2H 17,
// shield 14, held 23) and wants the equip slot (21/22/15). Keys are the
// canonical armory slot names.
export const POSITION_SLOTS = {
  'Main Hand': 21,
  'Off Hand': 22,
  Ranged: 15,
};

// The viewer never displays these slots; do not create phantom parts.
export const HIDDEN_SLOTS = new Set(['Neck', 'Finger 1', 'Finger 2', 'Trinket 1', 'Trinket 2']);

// ZAM character customization part names → imported appearance keys.
const CHARACTER_PART = {
  Face: 'face',
  'Skin Color': 'skin',
  'Hair Style': 'hairStyle',
  'Hair Color': 'hairColor',
  'Facial Hair': 'facialStyle',
  Mustache: 'facialStyle',
  Beard: 'facialStyle',
  Sideburns: 'facialStyle',
  'Face Shape': 'facialStyle',
  Eyebrow: 'facialStyle',
};

export function zamRaceId(race) {
  if (typeof race !== 'string') return null;
  return ZAM_RACES[race.replace(/^Race/, '')] ?? null;
}

// model id: race*2-1+gender (gender 0 male, 1 female)
export function zamModelId(raceId, gender) {
  return raceId * 2 - 1 + (gender === 1 ? 1 : 0);
}

// Returns the ZAM slot for an equipped position, or null for armor slots
// that must use the resolver-provided meta slot (chest 5/20 probe etc.).
export function positionSlot(slotName) {
  return POSITION_SLOTS[slotName] ?? null;
}

/**
 * Builds the viewer item list from imported gear and resolved visuals.
 * @returns {{items: Array<[number, number]>, unresolved: string[]}}
 *   items: [zamSlot, displayId] pairs; unresolved: slot names without a
 *   visual mapping (never silently substituted).
 */
export function buildPreviewItems(gear, resolved) {
  const items = [];
  const unresolved = [];
  for (const slot of gear) {
    if (HIDDEN_SLOTS.has(slot.slotName)) continue;
    if (!slot.itemId) continue; // genuinely empty slot (e.g. empty off hand)
    const visual = resolved[String(slot.itemId)];
    if (!visual) {
      unresolved.push(slot.slotName);
      continue;
    }
    const slotOverride = positionSlot(slot.slotName);
    const zamSlot = slotOverride ?? visual.zamSlot;
    items.push([zamSlot, visual.displayId]);
  }
  return { items, unresolved };
}

/**
 * Translates appearance indexes into ZAM option/choice pairs using the
 * versioned customization metadata fetched from the provider.
 * Unknown parts fall back to their first choice; missing appearance values
 * use the default (first) choice.
 */
export function appearanceOptions(appearance, customization) {
  const parts = customization?.Options ?? [];
  const options = [];
  for (const [label, key] of Object.entries(CHARACTER_PART)) {
    const part = parts.find((p) => p.Name === label);
    const choices = part?.Choices;
    if (!choices?.length) continue;
    const index = key && appearance?.[key] !== undefined ? appearance[key] : 0;
    const choice = choices[index] ?? choices[0];
    if (choice?.Id) {
      options.push({ optionId: part.Id, choiceId: choice.Id });
    }
  }
  return options;
}
