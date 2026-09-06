import { expect, test } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const fixtureLink = readFileSync(
  fileURLToPath(new URL('../../cmd/wowsimcli/cmd/upgrades/testdata/fixed_individual_link.txt', import.meta.url)),
  'utf8',
).trim();

// Runs against the default (flag-off) server: the stage must stay an honest
// placeholder, never a half-enabled 3D preview.
test('default build keeps the 3D stage unavailable', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('wowsims export link').fill(fixtureLink);
  await page.getByRole('button', { name: 'Import settings', exact: true }).click();
  await expect(page.locator('[data-slot]')).toHaveCount(17);
  await expect(page.getByText('3D preview unavailable — provider integration pending', { exact: true })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Activate 3D', exact: true })).toHaveCount(0);
});
