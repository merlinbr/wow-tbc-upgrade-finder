import { expect, test } from '@playwright/test';
import { readFileSync } from 'node:fs';

const paladinLink = readFileSync(
  new URL('../../cmd/wowsimcli/cmd/upgrades/testdata/retribution_no_settings_link.txt', import.meta.url),
  'utf8',
).trim();
const mageLink = readFileSync(
  new URL('../../cmd/wowsimcli/cmd/upgrades/testdata/fixed_individual_link.txt', import.meta.url),
  'utf8',
).trim();

test('talent inspection preserves imported ranks and clears on tab change or re-import', async ({ page }) => {
  // Artwork is not the behavioral contract; exercise real imported ranks offline.
  await page.route('**/images/wow/**', (route) => route.abort());
  await page.goto('/');
  await page.getByLabel('wowsims export link').fill(paladinLink);
  await page.getByRole('button', { name: 'Import settings', exact: true }).click();
  await page.getByRole('tab', { name: 'Talents', exact: true }).click();

  const trees = page.locator('[data-region="talent-trees"]');
  const ranks = () => trees.getByRole('button').evaluateAll(nodes => nodes.map(node => node.getAttribute('aria-label')));
  const before = await ranks();
  await expect(trees.getByRole('region', { name: 'Holy talents', exact: true })).toContainText('5 points');
  await expect(trees.getByRole('region', { name: 'Protection talents', exact: true })).toContainText('11 points');
  await expect(trees.getByRole('region', { name: 'Retribution talents', exact: true })).toContainText('45 points');

  const node = page.getByRole('button', { name: 'Inspect Divine Strength, rank 5 of 5', exact: true });
  await node.focus();
  const tip = page.getByRole('tooltip');
  await expect(tip).toContainText('Divine Strength');
  await expect(tip).toHaveAttribute('id', await node.getAttribute('aria-describedby'));
  await page.keyboard.press('Escape');
  await expect(tip).toHaveCount(0);
  await page.keyboard.press('Enter');
  await expect(tip).toContainText('Rank 5 / 5');

  await page.getByRole('button', { name: 'Inspect Divine Intellect, rank 0 of 5', exact: true }).click();
  await expect(tip).toContainText('Divine Intellect');
  await expect(tip).toContainText('Rank 0 / 5');
  await expect(tip).toContainText('Unallocated');
  expect(await ranks()).toEqual(before);

  await page.getByRole('tab', { name: 'Gear', exact: true }).click();
  await expect(tip).toHaveCount(0);
  await page.getByRole('tab', { name: 'Talents', exact: true }).click();
  await node.focus();
  await expect(tip).toContainText('Divine Strength');
  await page.getByLabel('wowsims export link').fill(mageLink);
  await page.getByRole('button', { name: 'Import settings', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'TestMage', exact: true })).toBeVisible();
  await expect(tip).toHaveCount(0);
  await page.getByRole('tab', { name: 'Talents', exact: true }).click();
  await expect(trees.getByRole('region', { name: 'Arcane talents', exact: true })).toBeVisible();
  await expect(trees.getByRole('button', { name: /Divine Strength/ })).toHaveCount(0);
});
