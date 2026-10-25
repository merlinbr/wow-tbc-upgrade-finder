export async function seedCharacter(request, overrides = {}) {
  const response = await request.post('/api/characters', {
    data: {
      name: 'Aria',
      class: 'paladin',
      spec: 'paladin_retribution',
      race: 'human',
      level: 60,
      ...overrides,
    },
  });
  if (!response.ok()) {
    throw new Error(`seed character failed: ${response.status()} ${await response.text()}`);
  }
  return response.json();
}

export function buildIdFromUrl(url) {
  const hash = new URL(url).hash;
  return hash.split('?build=')[1] ?? '';
}
