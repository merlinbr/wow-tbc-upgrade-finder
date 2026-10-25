import { expect, test } from '@playwright/test';

import { equipItem, getCurrentBuild, seedCharacter } from './helpers.js';

async function applyBadHat(page) {
  await page.getByTestId('gear-slot-head').click();
  await page.getByPlaceholder('Search items').fill('bad hat');
  await expect(page.getByTestId('picker-row')).toHaveCount(1);
  await page.getByTestId('picker-row').first().click();
  await page.getByRole('button', { name: 'Apply to head' }).click();
  await expect(page.getByTestId('gear-slot-head')).toContainText('Bad Hat');
}

test('edit, save, reload, and keep a failed save intact', async ({ page, request }, testInfo) => {
  test.skip(testInfo.project.name === 'touch', 'desktop flow');
  const ch = await seedCharacter(request);
  await page.goto(`/#/characters/${ch.id}/gear`);
  await expect(page.getByTestId('save-bar')).toContainText('All changes saved');

  await applyBadHat(page);
  await expect(page.getByTestId('save-bar')).toContainText('Unsaved changes');

  await page.getByRole('button', { name: 'Save changes' }).click();
  await expect(page.getByTestId('save-bar')).toContainText('All changes saved');

  await page.reload();
  await expect(page.getByTestId('gear-slot-head')).toContainText('Bad Hat');
});

test('keeps local edits when the save conflicts', async ({ page, request }, testInfo) => {
  test.skip(testInfo.project.name === 'touch', 'desktop flow');
  const ch = await seedCharacter(request);
  const current = await getCurrentBuild(request, ch.id);

  await page.goto(`/#/characters/${ch.id}/gear`);
  await applyBadHat(page);

  const external = await request.patch(`/api/builds/${current.id}`, {
    data: { version: current.version, equipment: { head: { itemId: 23019 } } },
  });
  expect(external.ok()).toBeTruthy();

  await page.getByRole('button', { name: 'Save changes' }).click();
  await expect(
    page.getByText('This build changed elsewhere. Reload to see the latest version.'),
  ).toBeVisible();
  await expect(page.getByTestId('save-bar')).toContainText('Unsaved changes');
  await expect(page.getByTestId('gear-slot-head')).toContainText('Bad Hat');
});

test('keeps edits and reports violations on a failed save', async ({ page, request }, testInfo) => {
  test.skip(testInfo.project.name === 'touch', 'desktop flow');
  const ch = await seedCharacter(request);
  await equipItem(request, ch.id, 'waist', 20041);

  await page.goto(`/#/characters/${ch.id}/gear`);
  await page.getByRole('button', { name: 'Edit character' }).click();
  await page.getByLabel('Race').selectOption('undead');
  await page.getByRole('button', { name: 'Save character', exact: true }).click();
  await expect(page.getByText('incompatible_faction').first()).toBeVisible();

  await applyBadHat(page);
  await page.getByRole('button', { name: 'Save changes' }).click();

  await expect(page.getByTestId('violation-list').last()).toContainText('incompatible_faction');
  await expect(page.getByTestId('save-bar')).toContainText('Unsaved changes');
  await expect(page.getByTestId('gear-slot-head')).toContainText('Bad Hat');
  await expect(page.getByTestId('gear-slot-waist')).toContainText("Highlander's Plate Girdle");
});

test('responsive layout, keyboard focus and touch flow', async ({ page, request }, testInfo) => {
  test.skip(testInfo.project.name !== 'touch', 'touch project only');
  const ch = await seedCharacter(request);
  await page.goto(`/#/characters/${ch.id}/gear`);

  const columns = await page
    .locator('.gear-grid')
    .evaluate((element) => getComputedStyle(element).gridTemplateColumns);
  expect(columns.trim().split(/\s+/).length).toBe(1);

  const noOverflow = await page.evaluate(
    () => document.documentElement.scrollWidth <= window.innerWidth + 1,
  );
  expect(noOverflow).toBeTruthy();
  await expect(page.getByTestId('save-bar')).toBeVisible();

  const reached = new Set();
  for (let step = 0; step < 80 && reached.size < 17; step += 1) {
    await page.keyboard.press('Tab');
    const testId = await page.evaluate(
      () => document.activeElement?.getAttribute('data-testid') ?? '',
    );
    if (testId.startsWith('gear-slot-')) {
      reached.add(testId);
    }
  }
  expect(reached.size).toBe(17);
  const outlineWidth = await page.evaluate(() => {
    const element = document.activeElement;
    return element ? getComputedStyle(element).outlineWidth : '';
  });
  expect(outlineWidth).toBe('3px');

  await page.getByTestId('gear-slot-head').tap();
  await page.getByPlaceholder('Search items').fill('bad hat');
  await expect(page.getByTestId('picker-row')).toHaveCount(1);
  await page.getByTestId('picker-row').first().tap();
  await page.getByRole('button', { name: 'Apply to head' }).tap();
  await expect(page.getByTestId('gear-slot-head')).toContainText('Bad Hat');

  await page.getByRole('tab', { name: 'Talents' }).tap();
  await expect(page.getByRole('tab', { name: 'Holy', exact: true })).toBeVisible();
  await expect(page.locator('.talent-tree:visible')).toHaveCount(1);

  await page.getByRole('tab', { name: 'Talents' }).focus();
  const talentReached = new Set();
  for (let step = 0; step < 80 && talentReached.size < 17; step += 1) {
    await page.keyboard.press('Tab');
    const talentId = await page.evaluate(
      () => document.activeElement?.getAttribute('data-talent-id') ?? '',
    );
    if (talentId) {
      talentReached.add(talentId);
    }
  }
  expect(talentReached.size).toBe(17);
  const talentOutline = await page.evaluate(() => {
    const element = document.activeElement;
    return element ? getComputedStyle(element).outlineWidth : '';
  });
  expect(talentOutline).toBe('3px');

  await page.getByTestId('talent-node').first().tap();
  await page.getByRole('button', { name: '+ Rank' }).tap();
  await expect(page.getByTestId('point-total')).toContainText('1 / 51');
});
