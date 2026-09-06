// Class presentation values mirror the simulator UI:
// ui/core/player_classes/*.ts (hexColor) and ui/core/player_specs/*.ts
// (getIcon('medium') icon names).
const classColors = {
  ClassDruid: '#ff7d0a',
  ClassHunter: '#abd473',
  ClassMage: '#69ccf0',
  ClassPaladin: '#f58cba',
  ClassPriest: '#ffffff',
  ClassRogue: '#fff569',
  ClassShaman: '#2459ff',
  ClassWarlock: '#9482c9',
  ClassWarrior: '#c79c6e',
};

const classIcons = {
  ClassDruid: 'class_druid.jpg',
  ClassHunter: 'class_hunter.jpg',
  ClassMage: 'class_mage.jpg',
  ClassPaladin: 'class_paladin.jpg',
  // The simulator UI uses this priest icon as the class icon.
  ClassPriest: 'spell_shadow_shadowwordpain.jpg',
  ClassRogue: 'class_rogue.jpg',
  ClassShaman: 'class_shaman.jpg',
  ClassWarlock: 'class_warlock.jpg',
  ClassWarrior: 'class_warrior.jpg',
};

export function classColor(value) {
  return classColors[value] ?? '';
}

export function classIcon(value) {
  const name = classIcons[value];
  return name ? `https://wow.zamimg.com/images/wow/icons/medium/${name}` : '';
}

export function avgItemLevel(gear = []) {
  const levels = gear
    .filter((slot) => (slot.itemId ?? 0) !== 0 && (slot.ilvl ?? 0) > 0)
    .map((slot) => slot.ilvl);
  if (levels.length === 0) return 0;
  return Math.round(levels.reduce((sum, value) => sum + value, 0) / levels.length);
}
