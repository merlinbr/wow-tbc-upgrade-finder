import { expect, test } from '@playwright/test';

test('creates a character and reopens it', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Create character' }).click();
  await page.getByLabel('Name').fill('Elara');
  await page.getByLabel('Class').selectOption('paladin');
  await page.getByLabel('Specialization').selectOption('paladin_retribution');
  await page.getByLabel('Race').selectOption('human');
  await page.getByRole('button', { name: 'Create', exact: true }).click();

  await expect(page.getByRole('heading', { name: 'Elara' })).toBeVisible();
  await page.reload();
  await expect(page.getByRole('heading', { name: 'Elara' })).toBeVisible();
});

test('rejects an invalid name and an invalid race/class pair', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Create character' }).click();
  await page.getByLabel('Class').selectOption('paladin');

  await page.getByRole('button', { name: 'Create', exact: true }).click();
  await expect(page.getByText('invalid_character_name')).toBeVisible();
  await expect(page.getByRole('dialog')).toBeVisible();

  await page.getByLabel('Name').fill('Orcadin');
  // Force the invalid race past the client-side disabled option.
  await page.evaluate(() => {
    const select = document.querySelector('#character-race');
    const option = Array.from(select.options).find((entry) => entry.value === 'orc');
    option.disabled = false;
    select.value = 'orc';
    select.dispatchEvent(new Event('change', { bubbles: true }));
  });
  await page.getByRole('button', { name: 'Create', exact: true }).click();
  await expect(page.getByText('race_class_mismatch')).toBeVisible();
  await expect(page.getByRole('dialog')).toBeVisible();

  await page.getByRole('button', { name: 'Cancel' }).click();
  await expect(page.getByRole('heading', { name: 'Characters' })).toBeVisible();
});
