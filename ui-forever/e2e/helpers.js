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

export async function getCurrentBuild(request, characterId) {
  const response = await request.get(`/api/characters/${characterId}/builds`);
  const builds = (await response.json()).builds;
  return builds.find((build) => build.kind === 'current');
}

export async function equipItem(request, characterId, slot, itemId) {
  const buildsResponse = await request.get(`/api/characters/${characterId}/builds`);
  const builds = (await buildsResponse.json()).builds;
  const current = builds.find((build) => build.kind === 'current');
  const response = await request.patch(`/api/builds/${current.id}`, {
    data: { version: current.version, equipment: { [slot]: { itemId } } },
  });
  if (!response.ok()) {
    throw new Error(`equip item failed: ${response.status()} ${await response.text()}`);
  }
  return response.json();
}

export async function seedTalents(request, characterId, talents) {
  const buildsResponse = await request.get(`/api/characters/${characterId}/builds`);
  const builds = (await buildsResponse.json()).builds;
  const current = builds.find((build) => build.kind === 'current');
  const response = await request.patch(`/api/builds/${current.id}`, {
    data: { version: current.version, talents },
  });
  if (!response.ok()) {
    throw new Error(`seed talents failed: ${response.status()} ${await response.text()}`);
  }
  return response.json();
}

export function buildIdFromUrl(url) {
  const hash = new URL(url).hash;
  return hash.split('?build=')[1] ?? '';
}
