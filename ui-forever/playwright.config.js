import { defineConfig, devices } from '@playwright/test';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const port = 4173;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  workers: 1,
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    trace: 'on-first-retry',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    command: `go run ./cmd/forever-armory --no-browser --addr 127.0.0.1:${port} --db :memory:`,
    cwd: repoRoot,
    url: `http://127.0.0.1:${port}/api/meta`,
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },
});
