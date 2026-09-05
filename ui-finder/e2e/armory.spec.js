import { expect, test } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const fixtureLink = readFileSync(
  fileURLToPath(new URL('../../cmd/wowsimcli/cmd/upgrades/testdata/fixed_individual_link.txt', import.meta.url)),
  'utf8',
).trim();

const enchantedFixtureLink = readFileSync(
  fileURLToPath(new URL('../../cmd/wowsimcli/cmd/upgrades/testdata/retribution_no_settings_link.txt', import.meta.url)),
  'utf8',
).trim();

test('imports the armory, ranks upgrades, copies the report, and cancels a job', async ({ page }) => {
  await page.goto('/');

  // The fixed mage fixture has no enchants; the Retribution fixture does.
  // Import it first to exercise the enchant effect line, then re-import the
  // mage fixture for the remaining armory and ranking assertions.
  await page.getByLabel('wowsims export link').fill(enchantedFixtureLink);
  await page.getByRole('button', { name: 'Import settings', exact: true }).click();
  await expect(page.locator('.enchant-effect').first()).toBeVisible();

  await page.getByLabel('wowsims export link').fill(fixtureLink);
  await page.getByRole('button', { name: 'Import settings', exact: true }).click();

  await expect(page.getByRole('heading', { name: 'TestMage', exact: true })).toBeVisible();
  await expect(page.locator('[data-slot]')).toHaveCount(17);
  await expect(page.getByLabel(/sockets$/i).first()).toBeVisible();
  await expect(page.locator('.gear-details summary').first()).toBeVisible();
  await expect(page.locator('.item-ilvl').first()).toBeVisible();
  await expect(page.getByRole('link', { name: 'Find upgrades', exact: true })).toBeVisible();
  await page.getByRole('tab', { name: 'Stats', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Raw stats', exact: true })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Derived percentages', exact: true })).toBeVisible();
  await expect(page.getByText('raid buffed (link settings)', { exact: true })).toBeVisible();
  await page.getByRole('tab', { name: 'Talents', exact: true }).click();
  await expect(page.locator('.talent-tree')).toHaveCount(3);
  await page.getByRole('tab', { name: 'Gear', exact: true }).click();
  await expect(page.getByRole('link', { name: 'wowsims/tbc-new', exact: true })).toHaveAttribute(
    'href',
    'https://github.com/wowsims/tbc-new',
  );

  // Full tooltip on the armory (hover the first gear icon).
  const firstTrigger = page.locator('.gear-trigger').first();
  await firstTrigger.hover();
  const tooltip = page.locator('.item-tooltip').first();
  await expect(tooltip).toBeVisible();
  await expect(tooltip).toContainText(/Item Level/i);
  await expect(tooltip.locator('.tooltip-enchant, .tooltip-socket').first()).toBeAttached();

  // Summary tooltip in the report appears after ranking; asserted below.

  const maxPhase = page.getByLabel('Maximum phase');
  const screening = page.getByLabel('Screening iterations');
  const confirmation = page.getByLabel('Confirmation iterations');
  await maxPhase.fill('1');
  await screening.fill('1');
  await confirmation.fill('1');
  await expect(maxPhase).toHaveValue('1');
  await expect(screening).toHaveValue('1');
  await expect(confirmation).toHaveValue('1');
  const kindBoxes = page.locator('.source-kind-group input[type="checkbox"]');
  const kindCount = await kindBoxes.count();
  for (let index = 0; index < kindCount; index += 1) {
    await kindBoxes.nth(index).check();
  }

  const progress = page.getByRole('status');
  await page.getByRole('button', { name: 'Start ranking', exact: true }).click();
  await expect(progress).toContainText(/queued|running/i, { timeout: 15_000 });
  await expect(page.getByRole('heading', { name: 'Upgrade report', exact: true })).toBeVisible({ timeout: 15_000 });
  await expect(page.getByRole('table', { name: 'Confirmed single-item upgrades' })).toBeVisible({ timeout: 15_000 });
  await expect(page.getByText('Source filters')).toBeVisible();
  await expect(page.locator('.report-summary dd').filter({ hasText: 'Crafting' })).toBeVisible();
  const reportTrigger = page.locator('.report-item-trigger').first();
  await reportTrigger.hover();
  await expect(reportTrigger.locator('.item-tooltip')).toBeVisible();

  await page.getByRole('button', { name: 'Copy JSON', exact: true }).click();
  await expect(page.getByText('Report copied to clipboard.', { exact: true })).toBeVisible();

  await screening.fill('300');
  await confirmation.fill('1000');
  await expect(screening).toHaveValue('300');
  await expect(confirmation).toHaveValue('1000');
  await page.getByRole('button', { name: 'Start ranking', exact: true }).click();
  await expect(progress).toContainText(/queued|running/i, { timeout: 15_000 });
  const cancel = page.getByRole('button', { name: 'Cancel ranking', exact: true });
  await expect(cancel).toBeVisible();
  await cancel.click();
  await expect(progress).toHaveText('Ranking canceled.');
  await expect(page.getByRole('heading', { name: 'Upgrade report', exact: true })).toHaveCount(0);
});
