import { expect, test } from '@playwright/test';

import { buildIdFromUrl, seedCharacter } from './helpers.js';

test('shows all 17 slots and the empty current build', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await page.goto(`/#/characters/${ch.id}/gear`);
  await expect(page.getByRole('heading', { name: ch.name })).toBeVisible();
  await expect(page.getByTestId('gear-slot')).toHaveCount(17);
  await expect(page.getByText('Empty')).toHaveCount(17);
  await expect(page.getByRole('tab', { name: 'Talents' })).toBeVisible();
});

test('manages plan lifecycle without mutating current', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await page.goto(`/#/characters/${ch.id}/gear`);

  await page.getByRole('button', { name: 'New plan' }).click();
  await page.getByLabel('Plan name').fill('Pre-raid');
  await page.getByLabel('Start from').selectOption({ label: 'Current' });
  await page.getByRole('button', { name: 'Create plan' }).click();
  await expect(page.getByRole('dialog')).toBeHidden();

  const planId = buildIdFromUrl(page.url());
  expect(planId).not.toBe('');
  await expect(page.locator('#build-select')).toHaveValue(planId);

  await page.locator('#build-select').selectOption({ label: 'Current' });
  await expect(page.getByText('Empty')).toHaveCount(17);

  await page.locator('#build-select').selectOption({ label: 'Pre-raid' });
  await expect(page.locator('#build-select')).toHaveValue(planId);

  await page.getByRole('button', { name: 'Duplicate' }).click();
  await expect(page.locator('#build-select')).not.toHaveValue(planId);
  await expect(page.getByRole('dialog')).toBeHidden();
  const duplicateId = buildIdFromUrl(page.url());
  expect(duplicateId).not.toBe(planId);

  await page.getByRole('button', { name: 'Rename' }).click();
  await page.getByLabel('Plan name').fill('Pre-raid v2');
  await page.getByRole('button', { name: 'Rename plan' }).click();
  await expect(page.getByRole('dialog')).toBeHidden();
  await expect(page.locator('#build-select')).toHaveValue(duplicateId);
  await expect(
    page.locator('#build-select option', { hasText: /^Pre-raid v2$/ }),
  ).toHaveCount(1);

  await page.locator('#build-select').selectOption({ label: 'Pre-raid' });
  await page.getByRole('button', { name: 'Delete' }).click();
  await page.getByRole('button', { name: 'Delete plan' }).click();
  await expect(page.getByRole('dialog')).toBeHidden();
  await expect(
    page.locator('#build-select option', { hasText: /^Pre-raid$/ }),
  ).toHaveCount(0);

  await page.locator('#build-select').selectOption({ label: 'Pre-raid v2' });
  await expect(page.locator('#build-select')).toHaveValue(duplicateId);
  await page.reload();
  await expect(page.locator('#build-select')).toHaveValue(duplicateId);
});

test('identity edit warns without deleting equipment', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  const buildsResponse = await request.get(`/api/characters/${ch.id}/builds`);
  const current = (await buildsResponse.json()).builds.find((build) => build.kind === 'current');
  const saved = await request.patch(`/api/builds/${current.id}`, {
    data: { version: current.version, equipment: { waist: { itemId: 20041 } } },
  });
  expect(saved.ok()).toBeTruthy();

  await page.goto(`/#/characters/${ch.id}/gear`);
  await expect(page.getByTestId('gear-slot-waist')).toContainText("Highlander's Plate Girdle");

  await page.getByRole('button', { name: 'Edit character' }).click();
  await page.getByLabel('Race').selectOption('undead');
  await page.getByRole('button', { name: 'Save character', exact: true }).click();

  await expect(page.getByText('incompatible_faction')).toBeVisible();
  await expect(page.getByTestId('gear-slot-waist')).toContainText("Highlander's Plate Girdle");
});
