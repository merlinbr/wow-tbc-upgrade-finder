import { readFile, writeFile } from 'node:fs/promises';

const root = new URL('../../', import.meta.url);
const output = new URL('../src/lib/talentIcons.json', import.meta.url);
const classes = ['druid', 'hunter', 'mage', 'paladin', 'priest', 'rogue', 'shaman', 'warlock', 'warrior'];
const readJson = async (path) => JSON.parse(await readFile(new URL(path, root), 'utf8'));
const [db, ...classTrees] = await Promise.all([
  readJson('assets/database/db.json'),
  ...classes.map((name) => readJson(`ui/core/talents/trees/${name}.json`)),
]);
const source = new Map(db.spellIcons.map(({ id, icon }) => [id, icon]));
const icons = {};
for (const tree of classTrees.flat()) {
  for (const talent of tree.talents) {
    const id = talent.spellIds[0];
    const icon = source.get(id);
    if (!icon) throw new Error(`Missing talent icon: ${talent.fancyName} (spell ${id})`);
    icons[id] = icon;
  }
}
const text = `${JSON.stringify(icons, null, 2)}\n`;
if (process.argv.includes('--check')) {
  const current = await readFile(output, 'utf8').catch((error) => {
    if (error.code === 'ENOENT') return '';
    throw error;
  });
  if (current !== text) throw new Error('Stale talentIcons.json; run node scripts/generate-talent-icons.mjs');
} else {
  await writeFile(output, text);
}
