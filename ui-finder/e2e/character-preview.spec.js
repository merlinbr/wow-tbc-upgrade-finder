import { expect, test } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const fixtureLink = readFileSync(
  fileURLToPath(new URL('../../cmd/wowsimcli/cmd/upgrades/testdata/fixed_individual_link.txt', import.meta.url)),
  'utf8',
).trim();

async function importFixture(page) {
  await page.goto('/');
  await page.getByLabel('wowsims export link').fill(fixtureLink);
  await page.getByRole('button', { name: 'Import settings', exact: true }).click();
  await expect(page.locator('[data-slot]')).toHaveCount(17);
}

// Runs against the --enable-3d server with the fake adapter: deterministic
// activation, controls, pause/resume and unmount cleanup.
test('activated preview with fake adapter: controls, pause, unmount', async ({ page }) => {
  await page.addInitScript(() => {
    window.__VISUAL_PROVIDER__ = 'fake';
  });
  await importFixture(page);

  await page.getByRole('button', { name: 'Activate 3D', exact: true }).click();
  const fakeViewer = page.getByTestId('fake-viewer');
  await expect(fakeViewer).toBeVisible();
  await expect(fakeViewer).toHaveText(/race 8, female/); // Troll, default female preset

  await page.getByLabel('Body preset').selectOption('male');
  await expect(fakeViewer).toHaveText(/race 8, male/);

  await page.getByRole('button', { name: 'Rotate', exact: true }).click();
  await page.getByRole('button', { name: 'Reset', exact: true }).click();
  await expect(fakeViewer).toBeVisible();

  await page.getByRole('button', { name: 'Pause', exact: true }).click();
  await expect(fakeViewer).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Resume 3D', exact: true })).toBeVisible();

  await page.getByRole('button', { name: 'Resume 3D', exact: true }).click();
  await expect(fakeViewer).toBeVisible();

  // Gear-tab unmount must destroy the viewer.
  await page.getByRole('tab', { name: 'Stats', exact: true }).click();
  await expect(fakeViewer).toHaveCount(0);
  await page.getByRole('tab', { name: 'Gear', exact: true }).click();
  await expect(fakeViewer).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Activate 3D', exact: true })).toBeVisible();
});
