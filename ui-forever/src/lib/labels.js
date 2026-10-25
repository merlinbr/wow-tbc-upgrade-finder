export function label(value) {
  if (!value) {
    return '';
  }
  return String(value)
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}

export const SLOT_ORDER = [
  'head',
  'neck',
  'shoulder',
  'back',
  'chest',
  'wrist',
  'main_hand',
  'off_hand',
  'hands',
  'waist',
  'legs',
  'feet',
  'finger1',
  'finger2',
  'trinket1',
  'trinket2',
  'ranged',
];

export const SLOT_LABELS = {
  head: 'Head',
  neck: 'Neck',
  shoulder: 'Shoulder',
  back: 'Back',
  chest: 'Chest',
  wrist: 'Wrist',
  main_hand: 'Main Hand',
  off_hand: 'Off Hand',
  hands: 'Hands',
  waist: 'Waist',
  legs: 'Legs',
  feet: 'Feet',
  finger1: 'Finger 1',
  finger2: 'Finger 2',
  trinket1: 'Trinket 1',
  trinket2: 'Trinket 2',
  ranged: 'Ranged',
};

export function slotLabel(slot) {
  return SLOT_LABELS[slot] ?? slot;
}
