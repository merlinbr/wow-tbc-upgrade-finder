import { defineConfig, devices } from '@playwright/test';
import { fileURLToPath } from 'node:url';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  reporter: 'line',
  use: {
    baseURL: 'http://127.0.0.1:43123',
    permissions: ['clipboard-write'],
    trace: 'retain-on-failure',
  },
  projects: [{
    name: 'chromium',
    use: { ...devices['Desktop Chrome'] },
  }],
  webServer: {
    command: 'rtk go run ./cmd/wowsimcli rank-upgrades --addr 127.0.0.1:43123 --no-browser',
    url: 'http://127.0.0.1:43123/',
    reuseExistingServer: false,
    cwd: fileURLToPath(new URL('..', import.meta.url)),
  },
});
