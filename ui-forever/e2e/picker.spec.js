import { expect, test } from '@playwright/test';

import { equipItem, seedCharacter } from './helpers.js';

test('filters by slot and applies an item locally', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await page.goto(`/#/characters/${ch.id}/gear`);
  await page.getByTestId('gear-slot-head').click();
  await page.getByPlaceholder('Search items').fill('bad hat');
  await expect(page.getByTestId('picker-row')).toHaveCount(1);
  await page.getByTestId('picker-row').first().click();
  await page.getByRole('button', { name: 'Apply to head' }).click();
  await expect(page.getByTestId('gear-slot-head')).toContainText('Bad Hat');
});

test('shows flattened stats, unknown sources and keeps the selection', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await equipItem(request, ch.id, 'head', 23019);
  await page.goto(`/#/characters/${ch.id}/gear`);
  await expect(page.getByTestId('gear-slot-head')).toContainText('Icebane Helmet');

  await page.getByTestId('gear-slot-head').click();
  await page.getByPlaceholder('Search items').fill('icebane helmet');
  await expect(page.getByTestId('picker-row')).toHaveCount(1);
  await page.getByTestId('picker-row').first().click();
  await expect(page.getByRole('heading', { name: 'Icebane Helmet' })).toBeVisible();
  await expect(page.getByText('757').first()).toBeVisible();
  await expect(page.getByText('Level requirement: unknown')).toBeVisible();

  await page.getByPlaceholder('Search items').fill('bad hat');
  await expect(page.getByTestId('picker-row')).toHaveCount(1);
  await expect(page.getByTestId('picker-row').first()).toContainText('Bad Hat');
  await page.getByTestId('picker-row').first().click();
  await expect(page.getByRole('heading', { name: 'Bad Hat' })).toBeVisible();
  await expect(page.getByText('Unknown source').first()).toBeVisible();

  await page.getByRole('tab', { name: 'Enchants' }).click();
  await page.getByRole('tab', { name: 'Items' }).click();
  await expect(page.getByRole('heading', { name: 'Bad Hat' })).toBeVisible();
});

test('disables incompatible items without changing the slot', async ({ page, request }) => {
  const ch = await seedCharacter(request, {
    name: 'Mira',
    class: 'mage',
    spec: 'mage_frost',
    race: 'gnome',
  });
  await page.goto(`/#/characters/${ch.id}/gear`);
  await page.getByTestId('gear-slot-head').click();
  await page.getByPlaceholder('Search items').fill('icebane helmet');
  await page.getByTestId('picker-row').first().click();
  await expect(page.getByText('incompatible_armor').first()).toBeVisible();
  await expect(page.getByRole('button', { name: 'Apply to head' })).toBeDisabled();
  await expect(page.getByTestId('gear-slot-head')).toContainText('Empty');
});

test('disables a weapon-only enchant on an armor slot', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await equipItem(request, ch.id, 'head', 23019);
  await page.goto(`/#/characters/${ch.id}/gear`);
  await page.getByTestId('gear-slot-head').click();
  await page.getByRole('tab', { name: 'Enchants' }).click();
  await page.getByPlaceholder('Search enchants').fill('counterweight');
  await expect(page.getByTestId('enchant-row')).toHaveCount(1);
  await page.getByTestId('enchant-row').first().click();
  await expect(page.getByText('incompatible_enchant_slot').first()).toBeVisible();
  await expect(page.getByRole('button', { name: 'Apply enchant' })).toBeDisabled();
});

test('applies a compatible enchant locally', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await equipItem(request, ch.id, 'head', 23019);
  await page.goto(`/#/characters/${ch.id}/gear`);
  await page.getByTestId('gear-slot-head').click();
  await page.getByRole('tab', { name: 'Enchants' }).click();
  await page.getByPlaceholder('Search enchants').fill('Lesser Arcane Amalgamation');
  await page.getByTestId('enchant-row').first().click();
  await page.getByRole('button', { name: 'Apply enchant' }).click();
  await expect(page.getByTestId('gear-slot-head')).toContainText('Enchant effect');
});
