import { expect, test } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const fixtureLink = readFileSync(
  fileURLToPath(new URL('../../cmd/wowsimcli/cmd/upgrades/testdata/fixed_individual_link.txt', import.meta.url)),
  'utf8',
).trim();

const svgIcon = (label) => `<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48"><rect width="48" height="48" fill="#345072"/><text x="24" y="29" font-size="9" text-anchor="middle" fill="#fff">${label}</text></svg>`;

// Deterministic icon artwork; abort() routes are registered after the generic
// stub so the more specific pattern takes precedence.
async function stubIcons(page, failedIcons = []) {
  await page.route('**/images/wow/icons/large/**', async (route) => {
    const name = decodeURIComponent(route.request().url().split('/').pop()).replace('.jpg', '');
    await route.fulfill({ contentType: 'image/svg+xml', body: svgIcon(name.slice(0, 12)) });
  });
  for (const icon of failedIcons) {
    await page.route(`**/images/wow/icons/large/${icon}.jpg`, (route) => route.abort());
  }
}

async function importArmory(page, modify) {
  await page.goto('/');
  const state = { calls: 0 };
  await page.route('**/api/import', async (route) => {
    const response = await route.fetch();
    const payload = await response.json();
    state.calls += 1;
    modify?.(payload, state.calls);
    await route.fulfill({ response, json: payload });
  });
  await page.getByLabel('wowsims export link').fill(fixtureLink);
  await page.getByRole('button', { name: 'Import settings', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'TestMage', exact: true })).toBeVisible();
  return state;
}

function slotOf(payload, name) {
  return payload.gear.find((slot) => slot.slotName === name);
}

function plainNeck(payload) {
  Object.assign(slotOf(payload, 'Neck'), {
    itemName: 'Plain Neck Fixture', icon: 'fixture_neck', quality: 3, ilvl: 120, phase: 2, armorType: 2,
    stats: { intellect: 20, stamina: 15 }, randomSuffix: null, enchant: null, sockets: [],
    socketBonus: { stats: {}, active: false },
  });
}

function gemmedChest(payload) {
  Object.assign(slotOf(payload, 'Chest'), {
    itemName: 'Tooltip Chest Fixture', icon: 'fixture_chest', quality: 4, ilvl: 146, phase: 3, armorType: 4,
    stats: { armor: 1825, strength: 56, stamina: 48, intellect: 31, melee_crit_rating: 31 },
    randomSuffix: null,
    enchant: { id: 0, name: 'Fixture enchant', description: '+6 All Stats', icon: '', stats: {} },
    sockets: [
      { color: 2, gem: { id: 101, name: 'Red Gem Fixture', icon: 'fixture_red', color: 2, stats: { strength: 10 } } },
      { color: 4, gem: { id: 102, name: 'Orange Gem Fixture', icon: 'fixture_orange', color: 6, stats: { strength: 5, melee_crit_rating: 5 } } },
      { color: 3, gem: { id: 103, name: 'Purple Gem Fixture', icon: 'fixture_purple', color: 7, stats: { stamina: 7, strength: 5 } } },
    ],
    socketBonus: { stats: { spell_damage: 5, healing_power: 5 }, active: true },
  });
}

test.describe('item tooltip', () => {
  test.use({ viewport: { width: 1280, height: 900 } });

  test('gemmed chest shows item icon, real gem icons, ordered sections, active bonus', async ({ page }, testInfo) => {
    await stubIcons(page);
    await importArmory(page, gemmedChest);

    const chest = page.locator('[data-slot="Chest"]');
    await chest.locator('.name-trigger').hover();
    const tip = page.getByRole('tooltip');
    await expect(tip).toHaveCount(1);
    await expect(tip).toBeVisible();

    // External item icon and the three actual gem icons.
    await expect(tip.getByRole('img', { name: 'Tooltip Chest Fixture', exact: true })).toBeVisible();
    await expect(tip.locator('.tooltip-socket img')).toHaveCount(3);
    await expect(tip.locator('[data-gem-id="102"] img')).toHaveAttribute('src', /fixture_orange\.jpg$/);
    await expect(tip.getByRole('img', { name: 'Red Gem Fixture', exact: true })).toBeVisible();

    // Base stats, enchant wording without an Equip: prefix, bonus state.
    await expect(tip.locator('.tooltip-stat').first()).toHaveText('1825 Armor');
    await expect(tip.locator('.tooltip-enchant')).toHaveText('+6 All Stats');
    await expect(tip).not.toContainText('Equip: +6 All Stats');
    await expect(tip.locator('.tooltip-socket-bonus')).not.toHaveClass(/inactive/);
    await expect(tip.locator('.tooltip-socket-bonus')).toContainText('+5 Healing Power, +5 Spell Damage');

    // Section order: base stats, then enchant, then gems, then bonus.
    const enchants = await tip.locator('.tooltip-enchant').boundingBox();
    const firstSocket = await tip.locator('.tooltip-socket').first().boundingBox();
    const bonus = await tip.locator('.tooltip-socket-bonus').boundingBox();
    expect(enchants.y).toBeLessThan(firstSocket.y);
    expect(firstSocket.y).toBeLessThan(bonus.y);

    await testInfo.attach('gemmed-chest-tooltip.png', { body: await tip.screenshot(), contentType: 'image/png' });
  });

  test('plain neck keeps name, level, phase and type without sockets', async ({ page }) => {
    await stubIcons(page);
    await importArmory(page, plainNeck);

    await page.locator('[data-slot="Neck"] .name-trigger').hover();
    const tip = page.getByRole('tooltip');
    await expect(tip).toContainText('Plain Neck Fixture');
    await expect(tip).toContainText('Item Level 120');
    await expect(tip).toContainText('Phase 2');
    await expect(tip).toContainText('Leather');
    await expect(tip.locator('.tooltip-socket')).toHaveCount(0);
    await expect(tip.locator('.tooltip-enchant')).toHaveCount(0);
  });

  test('empty sockets, mismatched gems, name-only gems and inactive bonus stay distinct', async ({ page }) => {
    await stubIcons(page);
    await importArmory(page, (payload) => {
      Object.assign(slotOf(payload, 'Head'), {
        itemName: 'Odd Head Fixture', icon: 'fixture_head', quality: 2, ilvl: 100, phase: 2, armorType: 1,
        stats: { strength: 10 }, randomSuffix: null, enchant: null,
        sockets: [
          // Red gem in a blue socket: mismatched but still shown.
          { color: 3, gem: { id: 201, name: 'Mismatched Red', icon: 'fixture_mismatch', color: 2, stats: { strength: 8 } } },
          // Genuinely empty socket.
          { color: 2, gem: null },
          // Gem with no numeric effects: effect text falls back to the name.
          { color: 4, gem: { id: 202, name: 'Name Only Gem', icon: 'fixture_nameonly', color: 4, stats: {} } },
        ],
        socketBonus: { stats: { strength: 4 }, active: false },
      });
    });

    await page.locator('[data-slot="Head"] .name-trigger').hover();
    const tip = page.getByRole('tooltip');
    await expect(tip.locator('.tooltip-socket')).toHaveCount(3);
    await expect(tip.locator('[data-gem-id="201"] img')).toHaveAttribute('src', /fixture_mismatch\.jpg$/);
    await expect(tip.locator('[data-gem-id="201"] .tooltip-gem')).toHaveText('+8 Strength');

    const emptySocket = tip.locator('[data-gem-id=""]');
    await expect(emptySocket).toHaveCount(1);
    await expect(emptySocket).toContainText('Red socket (empty)');
    await expect(emptySocket.locator('img')).toHaveCount(0);

    await expect(tip.locator('[data-gem-id="202"] .tooltip-gem')).toHaveText('Name Only Gem');
    await expect(tip.locator('.tooltip-socket-bonus')).toHaveClass(/inactive/);
  });

  test('failed icon falls back but gem text survives; re-import recovers the icon', async ({ page }) => {
    await stubIcons(page, ['fixture_chest']);
    await importArmory(page, (payload, calls) => {
      gemmedChest(payload);
      if (calls === 2) {
        slotOf(payload, 'Chest').icon = 'fixture_chest_ok';
      }
    });

    await page.locator('[data-slot="Chest"] .name-trigger').hover();
    let tip = page.getByRole('tooltip');
    await expect(tip).toBeVisible();
    await expect(tip.getByRole('img', { name: 'Tooltip Chest Fixture icon unavailable', exact: true })).toBeVisible();
    await expect(tip.locator('.tooltip-socket img')).toHaveCount(3, 'gem art must survive the item icon failure');

    // Re-import with a recovered icon URL: the failure state must reset.
    await page.getByLabel('wowsims export link').fill(fixtureLink);
    await page.getByRole('button', { name: 'Import settings', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'TestMage', exact: true })).toBeVisible();
    await page.locator('[data-slot="Chest"] .name-trigger').hover();
    tip = page.getByRole('tooltip');
    await expect(tip.getByRole('img', { name: 'Tooltip Chest Fixture', exact: true })).toBeVisible();
    await expect(tip.locator('.tooltip-icon-fallback')).toHaveCount(0);
  });

  test('weapon fixture shows hand type, weapon type, damage and speed', async ({ page }) => {
    await stubIcons(page);
    await importArmory(page, (payload) => {
      Object.assign(slotOf(payload, 'Main Hand'), {
        itemName: 'Sword Fixture', icon: 'fixture_sword', quality: 4, ilvl: 130, phase: 2,
        weaponType: 9, handType: 2, weaponDamageMin: 100, weaponDamageMax: 200, weaponSpeed: 3.5,
        stats: { attack_power: 40 }, randomSuffix: null, enchant: null, sockets: [],
        socketBonus: { stats: {}, active: false },
      });
    });

    await page.locator('[data-slot="Main Hand"] .name-trigger').hover();
    const tip = page.getByRole('tooltip');
    await expect(tip).toContainText('One-Hand');
    await expect(tip).toContainText('Sword');
    await expect(tip).toContainText('100 - 200 Damage');
    await expect(tip).toContainText('Speed 3.5');
    await expect(tip.locator('.tooltip-equip')).toHaveText('Equip: +40 Attack Power');
  });

  test('keyboard focus opens immediately, Escape suppresses until refocus', async ({ page }) => {
    await stubIcons(page);
    await importArmory(page, gemmedChest);

    const nameButton = page.locator('[data-slot="Chest"] .name-trigger');
    await nameButton.focus();
    const tip = page.getByRole('tooltip');
    await expect(tip).toBeVisible();
    const tipId = await tip.getAttribute('id');
    await expect(nameButton).toHaveAttribute('aria-describedby', tipId);

    await page.keyboard.press('Escape');
    await expect(tip).toHaveCount(0);
    await page.keyboard.press('Escape');
    await expect(tip).toHaveCount(0, 'suppressed while still focused');

    await nameButton.blur();
    await nameButton.focus();
    await expect(tip).toBeVisible();
  });

  test('moving between triggers and into the panel keeps exactly one tooltip', async ({ page }) => {
    await stubIcons(page);
    await importArmory(page, gemmedChest);

    const triggerId = await page.locator('[data-slot="Chest"] .name-trigger').getAttribute('aria-describedby');
    await page.locator('[data-slot="Chest"] .name-trigger').hover();
    let tip = page.getByRole('tooltip');
    await expect(tip).toBeVisible();
    const firstId = await tip.getAttribute('id');
    expect(firstId).toBe(triggerId);

    // Hover into the panel: it must stay open without any trigger hover.
    const body = tip.locator('.tooltip-body');
    await body.hover();
    await expect(tip).toBeVisible({ timeout: 2_000 });
    await expect(tip).toHaveCount(1);

    // Switch to the icon trigger of the same item: same single panel.
    await page.locator('[data-slot="Chest"] .gear-trigger').hover();
    await expect(tip).toHaveCount(1);
    await expect(tip).toBeVisible();
  });

  test('rapid movement across items ends with exactly one panel for the last item', async ({ page }) => {
    await stubIcons(page);
    await importArmory(page, (payload) => {
      plainNeck(payload);
      gemmedChest(payload);
    });

    await page.locator('[data-slot="Head"] .name-trigger').hover();
    await page.locator('[data-slot="Neck"] .name-trigger').hover();
    const tip = page.getByRole('tooltip');
    await expect(tip).toHaveCount(1);
    await expect(tip).toContainText('Plain Neck Fixture');
    await expect(tip).not.toContainText('Tooltip Chest Fixture');
  });

  test('both columns, bottom weapons, scroll and report table stay in viewport bounds', async ({ page }) => {
    await stubIcons(page);
    await importArmory(page, gemmedChest);

    const assertInViewport = async (slotName, side) => {
      const trigger = page.locator(`[data-slot="${slotName}"] .name-trigger`);
      await trigger.hover();
      const tip = page.getByRole('tooltip');
      await expect(tip).toBeVisible();
      const box = await page.locator('.tooltip-layer').boundingBox();
      const triggerBox = await trigger.boundingBox();
      const viewport = page.viewportSize();
      expect(box.x).toBeGreaterThanOrEqual(8);
      expect(box.y).toBeGreaterThanOrEqual(8);
      expect(box.x + box.width).toBeLessThanOrEqual(viewport.width - 8);
      expect(box.y + box.height).toBeLessThanOrEqual(viewport.height - 8);
      if (side === 'right') {
        expect(box.x).toBeGreaterThan(triggerBox.x + triggerBox.width);
      } else {
        expect(box.x + box.width).toBeLessThanOrEqual(triggerBox.x);
      }
    };

    // Head is in the left column (tooltip flips to the right side).
    await assertInViewport('Head', 'right');
    // Waist is in the right column (tooltip flips to the left side).
    await assertInViewport('Waist', 'left');

    // Bottom weapon strip: Main Hand is the first weapon slot.
    await assertInViewport('Main Hand', 'right');

    // Scrolling while open repositions the panel. A wheel scroll fires real
    // pointer boundary events on the hovered trigger (the trigger moves away
    // from the stationary pointer), so open via keyboard focus instead: the
    // panel must stay open and follow the anchor.
    await page.locator('[data-slot="Feet"] .name-trigger').focus();
    await expect(page.getByRole('tooltip')).toBeVisible();
    await page.mouse.wheel(0, 400);
    await expect(page.getByRole('tooltip')).toBeVisible();
    await expect.poll(async () => {
      const box = await page.locator('.tooltip-layer').boundingBox();
      const viewport = page.viewportSize();
      return box.y >= 8 && box.y + box.height <= viewport.height - 8;
    }).toBe(true);
  });

  test('long title wraps without crushing the phase, mobile layout keeps icon in header', async ({ page }) => {
    await stubIcons(page);
    await importArmory(page, (payload) => {
      Object.assign(slotOf(payload, 'Neck'), {
        itemName: 'Extraordinarily Long Neck Fixture Name That Should Wrap Cleanly Across Many Lines '.repeat(3).trim(),
        icon: 'fixture_long', quality: 4, ilvl: 146, phase: 3, armorType: 2,
        stats: { intellect: 41 }, randomSuffix: null, enchant: null, sockets: [],
        socketBonus: { stats: {}, active: false },
      });
    });

    await page.locator('[data-slot="Neck"] .name-trigger').hover();
    const tip = page.getByRole('tooltip');
    await expect(tip).toBeVisible();
    const nameBox = await tip.locator('.tooltip-name').boundingBox();
    const phaseBox = await tip.locator('.tooltip-phase').boundingBox();
    expect(nameBox.x + nameBox.width).toBeLessThanOrEqual(phaseBox.x);

    // Mobile-sized viewport: single-column footprint with the icon in the header.
    await page.setViewportSize({ width: 390, height: 844 });
    await expect(tip).toBeVisible();
    await expect(tip.locator('.tooltip-inline-icon')).toBeVisible();
    await expect(tip.locator('.tooltip-icon-slot')).toBeHidden();
    // Wait for the layer remeasure: mobile footprint drops the 38px icon column.
    await expect.poll(async () => (await page.locator('.tooltip-layer').boundingBox())?.width ?? 0)
      .toBeLessThanOrEqual(390 - 16);
    const box = await page.locator('.tooltip-layer').boundingBox();
    expect(box.x).toBeGreaterThanOrEqual(8);
    expect(box.x + box.width).toBeLessThanOrEqual(390 - 8);
    expect(box.y + box.height).toBeLessThanOrEqual(844 - 8);

    // Details remain readable at touch size via the existing disclosure.
    const details = page.locator('[data-slot="Neck"] details');
    await details.locator('summary').click();
    await expect(details.locator('dd').filter({ hasText: 'Leather' })).toBeVisible();
  });

  test('report summary tooltip uses the shared shell without inventing gems', async ({ page }) => {
    await stubIcons(page);
    await importArmory(page, gemmedChest);

    const maxPhase = page.getByLabel('Maximum phase');
    const screening = page.getByLabel('Screening iterations');
    const confirmation = page.getByLabel('Confirmation iterations');
    await maxPhase.fill('1');
    await screening.fill('1');
    await confirmation.fill('1');
    const kindBoxes = page.locator('.source-kind-group input[type="checkbox"]');
    for (let index = 0; index < await kindBoxes.count(); index += 1) {
      await kindBoxes.nth(index).check();
    }
    await page.getByRole('button', { name: 'Start ranking', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Upgrade report', exact: true })).toBeVisible({ timeout: 60_000 });

    const reportTrigger = page.locator('.report-item-trigger').first();
    await reportTrigger.hover();
    const tip = page.getByRole('tooltip');
    await expect(tip).toBeVisible();
    const itemName = (await reportTrigger.textContent()).split('(')[0].trim();
    await expect(tip).toContainText(itemName);
    await expect(tip.locator('.tooltip-socket')).toHaveCount(0, 'summary must not invent candidate gems');
    await expect(tip).toContainText(/Epic|Rare|Uncommon/);

    // The report table has horizontal overflow; the floating layer must escape it.
    const box = await page.locator('.tooltip-layer').boundingBox();
    const viewport = page.viewportSize();
    expect(box.x).toBeGreaterThanOrEqual(8);
    expect(box.x + box.width).toBeLessThanOrEqual(viewport.width - 8);
  });
});
