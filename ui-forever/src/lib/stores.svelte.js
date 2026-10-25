import { api } from './api.js';

export const state = $state({
  characters: [],
  loading: false,
  error: '',
});

export async function loadCharacters() {
  state.loading = true;
  state.error = '';
  try {
    const data = await api.get('/api/characters');
    state.characters = data.characters ?? [];
    return state.characters;
  } catch (error) {
    state.error = error.message ?? 'Could not load characters.';
    return state.characters;
  } finally {
    state.loading = false;
  }
}

export async function createCharacter(input) {
  const character = await api.post('/api/characters', input);
  state.characters = [...state.characters, character];
  return character;
}
