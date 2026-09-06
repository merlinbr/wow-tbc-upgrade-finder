// Prototype fixture data — real imported gear display IDs (Wowhead TBC XML).
// Paladin: retribution_no_settings_link.txt; Mage: fixed_individual_link.txt.
// [zamSlot, displayId] pairs; slot values verified against ZAM meta probes.
export const MH = Number(new URLSearchParams(location.search).get('mainhand') || 21);
export const CHEST = Number(new URLSearchParams(location.search).get('chest') || 20);

export const FIXTURES = {
  paladin: {
    label: 'Human Retribution Paladin (imported baseline)',
    race: 1,
    gender: 1,
    appearance: { skin: 7, face: 3, hairStyle: 9, hairColor: 5, facialStyle: 1 },
    items: [
      [1, 45779], // head — Furious Gizmatic Goggles
      [3, 46353], // shoulder
      [5, 42306], // chest (plate; armor/5 meta verified)
      [6, 42651], // waist
      [7, 37449], // legs
      [8, 45845], // feet
      [9, 46043], // wrist
      [10, 43549], // hands
      [16, 38976], // back cloak
      [MH, 39571], // main hand — Lionheart Executioner (2H axe), MH from URL
    ],
  },
  mage: {
    label: 'Troll Mage (imported baseline)',
    race: 8,
    gender: 1,
    appearance: { skin: 7, face: 3, hairStyle: 9, hairColor: 5, facialStyle: 1 },
    items: [
      [1, 38956], // head
      [3, 40645], // shoulder
      [20, 40468], // chest — Vestments of the Aldor (robe-cut; armor/20 meta verified)
      [6, 46112], // waist
      [7, 41384], // legs
      [8, 46116], // feet
      [9, 37721], // wrist
      [10, 41385], // hands
      [16, 38976], // back cloak
      [21, 43098], // main hand — Nathrezim Mindblade (1H)
      [22, 42626], // off hand — Orb of the Soul-Eater (held)
      [15, 43916], // ranged — Tirisfal Wand
    ],
  },
  robe: {
    label: 'Robes of Rhonin (robe-cut chest display)',
    race: 1,
    gender: 1,
    appearance: { skin: 7, face: 3, hairStyle: 9, hairColor: 5, facialStyle: 1 },
    items: [['CHEST', 45335]], // CHEST resolved from ?chest=5|20 (armor/20 verified)
  },
};
