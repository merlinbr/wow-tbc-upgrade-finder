import { defineConfig, devices } from '@playwright/test';
import { fileURLToPath } from 'node:url';

const cwd = fileURLToPath(new URL('..', import.meta.url));
const run = (port, flags) => `go run ./cmd/wowsimcli rank-upgrades --addr 127.0.0.1:${port} --no-browser ${flags}`;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  reporter: 'line',
  use: {
    permissions: ['clipboard-write'],
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      testIgnore: /character-preview\.spec\.js/,
      use: { ...devices['Desktop Chrome'], baseURL: 'http://127.0.0.1:43123' },
    },
    {
      // 3D preview flows run against a server with the authorized-integration
      // flag on; the adapter is the deterministic fake (no ZAM access).
      name: 'visuals',
      testMatch: /character-preview\.spec\.js/,
      use: { ...devices['Desktop Chrome'], baseURL: 'http://127.0.0.1:43124' },
    },
  ],
  webServer: [
    {
      command: run(43123, ''),
      url: 'http://127.0.0.1:43123/',
      reuseExistingServer: false,
      cwd,
    },
    {
      command: run(43124, '--enable-3d'),
      url: 'http://127.0.0.1:43124/',
      reuseExistingServer: false,
      cwd,
    },
  ],
});
