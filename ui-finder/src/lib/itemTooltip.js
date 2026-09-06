import { qualityLabel, socketColors, statLabel } from './labels.js';
import { armorTypes, weaponTypes, handTypes, rangedTypes, classes, professions } from './tooltipLabels.js';

// White base-stat section order (WoW tooltip convention for TBC).
const baseOrder = ['armor', 'bonus_armor', 'strength', 'agility', 'stamina',
  'intellect', 'spirit', 'health', 'mana', 'arcane_resistance',
  'fire_resistance', 'frost_resistance', 'nature_resistance', 'shadow_resistance'];
// Green "Equip:" secondary-stat section order.
const equipOrder = ['attack_power', 'ranged_attack_power', 'feral_attack_power',
  'healing_power', 'spell_damage', 'arcane_damage', 'fire_damage', 'frost_damage',
  'holy_damage', 'nature_damage', 'shadow_damage', 'physical_damage',
  'melee_hit_rating', 'spell_hit_rating', 'melee_crit_rating', 'spell_crit_rating',
  'melee_haste_rating', 'spell_haste_rating', 'expertise_rating',
  'armor_penetration', 'spell_penetration', 'defense_rating', 'block_rating',
  'block_value', 'dodge_rating', 'parry_rating', 'resilience_rating', 'mp5'];
const baseOrderIndex = new Map(baseOrder.map((key, index) => [key, index]));
const equipOrderIndex = new Map(equipOrder.map((key, index) => [key, index]));
const ratingNames = {
  melee_hit_rating: 'melee hit rating', spell_hit_rating: 'spell hit rating',
  melee_crit_rating: 'melee critical strike rating', spell_crit_rating: 'spell critical strike rating',
  melee_haste_rating: 'melee haste rating', spell_haste_rating: 'spell haste rating',
};

const numberText = (value) => Number(value).toFixed(2).replace(/\.?0+$/, '');
const signedText = (value) => `${value >= 0 ? '+' : ''}${numberText(value)}`;
const statText = (key, value) => `${signedText(value)} ${statLabel(key)}`;

function labelOr(map, value, kind) {
  return map[value] ?? `Unknown ${kind} (${value})`;
}

// Ordered stat entries: base order first, then equip order, then unknown keys
// lexically. Zero and non-finite values are dropped; negatives are preserved.
function orderedStatEntries(stats) {
  return Object.entries(stats)
    .filter(([, value]) => Number.isFinite(Number(value)) && Number(value) !== 0)
    .sort((a, b) => {
      const ra = baseOrderIndex.has(a[0]) ? [0, baseOrderIndex.get(a[0])]
        : equipOrderIndex.has(a[0]) ? [1, equipOrderIndex.get(a[0])]
        : [2, a[0]];
      const rb = baseOrderIndex.has(b[0]) ? [0, baseOrderIndex.get(b[0])]
        : equipOrderIndex.has(b[0]) ? [1, equipOrderIndex.get(b[0])]
        : [2, b[0]];
      if (ra[0] !== rb[0]) return ra[0] - rb[0];
      return typeof ra[1] === 'string' ? ra[1].localeCompare(rb[1]) : ra[1] - rb[1];
    });
}

function gemEffectText(gem) {
  const parts = orderedStatEntries(gem.stats ?? {}).map(([key, value]) => statText(key, value));
  return parts.length ? parts.join(', ') : (gem.name ?? '');
}

function buildFull(item) {
  const combined = {};
  for (const [key, value] of Object.entries(item.stats ?? {})) {
    combined[key] = Number(value);
  }
  for (const [key, value] of Object.entries(item.randomSuffix?.stats ?? {})) {
    combined[key] = (combined[key] ?? 0) + Number(value);
  }

  const baseLines = [];
  const equipLines = [];
  for (const [key, value] of orderedStatEntries(combined)) {
    if (baseOrderIndex.has(key)) {
      const text = key === 'armor' ? `${numberText(value)} ${statLabel(key)}` : statText(key, value);
      baseLines.push({ key, text });
    } else if (equipOrderIndex.has(key)) {
      const ratingLabel = ratingNames[key];
      const text = ratingLabel && value > 0
        ? `Equip: Improves ${ratingLabel} by ${numberText(value)}.`
        : `Equip: ${statText(key, value)}`;
      equipLines.push({ key, text });
    } else {
      // Unknown engine keys stay visible as plain white lines, not discarded.
      baseLines.push({ key, text: statText(key, value) });
    }
  }

  const min = Number(item.weaponDamageMin ?? 0);
  const max = Number(item.weaponDamageMax ?? 0);
  const speed = Number(item.weaponSpeed ?? 0);
  const weaponLines = speed > 0 && min > 0 && max >= min
    ? [`${numberText(min)} - ${numberText(max)} Damage`, `Speed ${numberText(speed)}`]
    : [];

  const sockets = (item.sockets ?? []).map((socket) => {
    const color = socket?.color ?? 0;
    const gem = socket?.gem ?? null;
    return {
      color,
      gem,
      text: gem ? gemEffectText(gem) : `${socketColors[color] ?? 'Unknown'} socket (empty)`,
    };
  });

  const bonus = item.socketBonus;
  const socketBonus = bonus && bonus.stats && Object.keys(bonus.stats).length
    ? {
      text: orderedStatEntries(bonus.stats).map(([key, value]) => statText(key, value)).join(', '),
      active: Boolean(bonus.active),
    }
    : null;

  const restrictions = [];
  if (item.unique) restrictions.push('Unique');
  if (item.classAllowlist?.length) {
    restrictions.push(`Classes: ${item.classAllowlist.map((value) => labelOr(classes, value, 'class')).join(', ')}`);
  }
  if (item.requiredProfession) {
    restrictions.push(`Requires ${labelOr(professions, item.requiredProfession, 'profession')}`);
  }

  return {
    slotLabel: item.handType ? labelOr(handTypes, item.handType, 'type') : (item.slotName ?? ''),
    typeLabel: item.rangedWeaponType ? labelOr(rangedTypes, item.rangedWeaponType, 'type')
      : item.weaponType ? labelOr(weaponTypes, item.weaponType, 'type')
      : item.armorType ? labelOr(armorTypes, item.armorType, 'type')
      : '',
    suffixLabel: item.randomSuffix?.name ?? '',
    weaponLines,
    baseLines,
    enchantLine: item.enchant ? (item.enchant.description || item.enchant.name || '') : '',
    sockets,
    socketBonus,
    restrictionLines: restrictions,
    equipLines,
    setName: item.setName ?? '',
  };
}

export function buildItemTooltip(item = {}, variant = 'full') {
  const name = item.itemName ?? item.name ?? '';
  const model = {
    name,
    icon: item.icon ?? '',
    quality: item.quality ?? 0,
    phase: item.phase ?? 0,
    ilvl: item.ilvl ?? 0,
    qualityLabel: qualityLabel(item.quality),
    slotLabel: '',
    typeLabel: '',
    suffixLabel: '',
    weaponLines: [],
    baseLines: [],
    enchantLine: '',
    sockets: [],
    socketBonus: null,
    restrictionLines: [],
    equipLines: [],
    setName: '',
  };
  if (variant === 'summary') return model;
  return { ...model, ...buildFull(item) };
}
