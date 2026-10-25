import { expect, test } from '@playwright/test';

import { seedCharacter, seedTalents } from './helpers.js';

test('allocates points, locks lower tiers and confirms reset', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await page.goto(`/#/characters/${ch.id}/talents`);

  const first = page.getByTestId('talent-node').first();
  await first.click();
  await page.getByRole('button', { name: '+ Rank' }).click();
  await expect(page.getByTestId('point-total')).toContainText('1 / 51');
  await expect(page.getByTestId('talent-node').nth(4)).toHaveAttribute('data-state', 'locked');

  await page.getByRole('button', { name: 'Reset talents…' }).click();
  await page.getByRole('button', { name: 'Cancel' }).click();
  await expect(page.getByTestId('point-total')).toContainText('1 / 51');
});

test('keeps broken talent points and excludes them from unlocking', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await seedTalents(request, ch.id, {
    'def:135164': 5,
    'def:134827': 5,
    'def:134828': 3,
    'def:134830': 2,
    'def:142623': 3,
    'def:134824': 1,
  });

  await page.goto(`/#/characters/${ch.id}/talents`);
  await expect(page.getByTestId('point-total')).toContainText('19 / 51');

  await page.locator('[data-talent-id="def:142623"]').click();
  await page.getByRole('button', { name: '− Rank' }).click();
  await page.getByRole('button', { name: '− Rank' }).click();
  await page.getByRole('button', { name: '− Rank' }).click();

  await expect(page.getByTestId('point-total')).toContainText('16 / 51');
  const child = page.locator('[data-talent-id="def:134824"]');
  await expect(child).toHaveAttribute('data-state', 'broken');
  await expect(child).toContainText('1/5');
});
