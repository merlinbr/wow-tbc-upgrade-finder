// Turns protobuf enum names ("RaceHuman", "ClassPaladin", "RetributionPaladin")
// into display text: strips the enum prefix, then splits camelCase words.
export function humanizeEnum(value, prefix) {
  if (!value) return '';
  const name = prefix && value.startsWith(prefix) ? value.slice(prefix.length) : value;
  return name.replace(/([a-z])([A-Z])/g, '$1 $2');
}

// Item stat keys arrive as snake_case of the engine's StatName() values
// (sim/core/stats/stats.go), e.g. "spell_hit_rating", "mp5".
const upperAcronyms = new Set(['mp5']);

export function statLabel(key) {
  if (!key) return '';
  return key
    .split('_')
    .map((word) => (upperAcronyms.has(word) ? word.toUpperCase() : word.charAt(0).toUpperCase() + word.slice(1)))
    .join(' ');
}

export function formatStatLine(key, value) {
  const formatted = Number(value).toFixed(2).replace(/\.00$/, '').replace(/(\.\d)0$/, '$1');
  return `+${formatted} ${statLabel(key)}`;
}

const qualityNames = {
  0: 'Junk', 1: 'Common', 2: 'Uncommon', 3: 'Rare',
  4: 'Epic', 5: 'Legendary', 6: 'Artifact', 7: 'Heirloom',
};

export function qualityLabel(value) {
  return qualityNames[value] ?? `Unknown quality (${value ?? '—'})`;
}

export const socketColors = {
  0: 'Unknown', 1: 'Meta', 2: 'Red', 3: 'Blue', 4: 'Yellow',
  5: 'Green', 6: 'Orange', 7: 'Purple', 8: 'Prismatic',
};

// proto.SourceFilterOption values, in enum order. Value 0 ("Unknown source")
// is not offered in filter checkboxes; it is governed by the include-unknown
// toggle.
export const sourceKinds = [
  { value: 0, label: 'Unknown source', proto: 'SourceUnknown' },
  { value: 1, label: 'Crafting', proto: 'SourceCrafting' },
  { value: 2, label: 'Quest', proto: 'SourceQuest' },
  { value: 3, label: 'Reputation', proto: 'SourceReputation' },
  { value: 4, label: 'PvP', proto: 'SourcePvP' },
  { value: 5, label: 'Dungeon', proto: 'SourceDungeon' },
  { value: 6, label: 'Heroic dungeon', proto: 'SourceHeroicDungeon' },
  { value: 7, label: 'Raid', proto: 'SourceRaid' },
  { value: 8, label: 'Heroic raid', proto: 'SourceHeroicRaid' },
  { value: 9, label: 'Raid finder', proto: 'SourceRaidFinder' },
  { value: 10, label: 'Flexible raid', proto: 'SourceFlexibleRaid' },
  { value: 11, label: 'Sold by vendor', proto: 'SourceSoldByVendor' },
];

const sourceKindsByValue = new Map(sourceKinds.map((k) => [k.value, k.label]));
const sourceKindsByProto = new Map(sourceKinds.map((k) => [k.proto, k.label]));

export function sourceKindLabel(value) {
  return sourceKindsByValue.get(value) ?? `Unknown source (${value ?? '—'})`;
}

export function humanizeSourceKind(name) {
  if (!name) return '';
  return sourceKindsByProto.get(name) ?? humanizeEnum(name, 'Source');
}
