import test from 'node:test';
import assert from 'node:assert/strict';
import { postJSON } from './api.js';

test('successful malformed JSON throws an invalid_response error', async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => ({
    ok: true,
    status: 200,
    json: async () => {
      throw new SyntaxError('unexpected end of JSON input');
    },
  });

  try {
    await assert.rejects(postJSON('/api/import', { link: 'test' }), {
      code: 'invalid_response',
      message: 'Invalid JSON response',
    });
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test('structured non-2xx errors retain their API code and message', async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => ({
    ok: false,
    status: 400,
    json: async () => ({ error: { code: 'invalid_link', message: 'Bad link' } }),
  });

  try {
    await assert.rejects(postJSON('/api/import', { link: 'test' }), {
      code: 'invalid_link',
      message: 'Bad link',
    });
  } finally {
    globalThis.fetch = originalFetch;
  }
});
