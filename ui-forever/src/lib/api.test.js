import assert from 'node:assert/strict';
import test from 'node:test';

import { ApiError, api } from './api.js';

test('parses typed API errors with violations', async () => {
  const original = globalThis.fetch;
  globalThis.fetch = async () =>
    new Response(
      JSON.stringify({
        error: {
          code: 'invalid_build',
          message: 'The build did not pass validation.',
          violations: [{ code: 'incompatible_armor', message: 'no plate' }],
        },
      }),
      { status: 422, headers: { 'Content-Type': 'application/json' } },
    );
  try {
    await assert.rejects(
      () => api.patch('/api/builds/1', { version: 1 }),
      (error) => {
        assert.ok(error instanceof ApiError);
        assert.equal(error.code, 'invalid_build');
        assert.equal(error.violations[0].code, 'incompatible_armor');
        return true;
      },
    );
  } finally {
    globalThis.fetch = original;
  }
});

test('returns null for 204 responses', async () => {
  const original = globalThis.fetch;
  globalThis.fetch = async () => new Response(null, { status: 204 });
  try {
    assert.equal(await api.del('/api/builds/1'), null);
  } finally {
    globalThis.fetch = original;
  }
});
