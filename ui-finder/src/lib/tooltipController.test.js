import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createTooltipController } from './tooltipController.js';

function fakeTiming() {
  let next = 0;
  const queue = [];
  return {
    setTimer(fn) {
      const id = ++next;
      queue.push({ id, fn });
      return id;
    },
    clearTimer(id) {
      const index = queue.findIndex((timer) => timer.id === id);
      if (index >= 0) queue.splice(index, 1);
    },
    fireNext() {
      const entry = queue.shift();
      if (!entry) throw new Error('no pending timer');
      entry.fn();
    },
    pending() {
      return queue.length;
    },
  };
}

function harness() {
  const changes = [];
  const timing = fakeTiming();
  const controller = createTooltipController((value) => changes.push(value), timing);
  return { changes, timing, controller };
}

function record(owner, item = {}, extra = {}) {
  return { owner, id: 'tip', item, variant: 'full', preferredSide: 'right', anchor: extra };
}

test('focus opens immediately and release closes', () => {
  const { changes, controller } = harness();
  const owner = Symbol('chest');
  const entered = record(owner, { itemName: 'Chest' });
  controller.enter(entered, 'focus');
  assert.equal(changes.at(-1)?.owner, owner);
  controller.release(owner);
  assert.equal(changes.at(-1), null);
  controller.destroy();
});

test('pointer opens after the delay and a leave before the delay cancels it', () => {
  const { changes, timing, controller } = harness();
  const owner = Symbol('neck');
  controller.enter(record(owner, { itemName: 'Neck' }), 'pointer');
  assert.equal(timing.pending(), 1);
  assert.equal(changes.length, 0);
  controller.leave(owner, 'pointer');
  assert.equal(timing.pending(), 0);
  assert.equal(changes.length, 0);
});

test('pointer close is delayed and is kept open by panel hover', () => {
  const { changes, timing, controller } = harness();
  const owner = Symbol('head');
  controller.enter(record(owner, { itemName: 'Head' }), 'pointer');
  timing.fireNext();
  assert.equal(changes.at(-1)?.owner, owner);
  controller.enterPanel(owner);
  controller.leave(owner, 'pointer');
  timing.fireNext();
  assert.equal(changes.at(-1)?.owner, owner, 'panel hover keeps it open');
  controller.leavePanel(owner);
  timing.fireNext();
  assert.equal(changes.at(-1), null);
});

test('focus wins over a pending hover and opens immediately', () => {
  const { changes, timing, controller } = harness();
  const owner = Symbol('chest');
  controller.enter(record(owner, { itemName: 'Chest' }), 'pointer');
  controller.enter(record(owner, { itemName: 'Chest' }), 'focus');
  assert.equal(changes.at(-1)?.owner, owner);
  assert.equal(timing.pending(), 0);
});

test('escape suppresses the still-active source until it leaves and re-enters', () => {
  const { changes, controller } = harness();
  const owner = Symbol('chest');
  controller.enter(record(owner, { itemName: 'Chest' }), 'focus');
  assert.equal(changes.at(-1)?.owner, owner);
  controller.dismiss();
  assert.equal(changes.at(-1), null);
  controller.enter(record(owner, { itemName: 'Chest' }), 'focus');
  assert.equal(changes.at(-1), null, 'suppressed while focus remains');
  controller.leave(owner, 'focus');
  controller.enter(record(owner, { itemName: 'Chest' }), 'focus');
  assert.equal(changes.at(-1)?.owner, owner);
});

test('a new owner closes the old one immediately and opens after its own delay', () => {
  const { changes, timing, controller } = harness();
  const first = Symbol('chest');
  const second = Symbol('neck');
  controller.enter(record(first, { itemName: 'Chest' }), 'pointer');
  timing.fireNext();
  controller.enter(record(second, { itemName: 'Neck' }), 'pointer');
  assert.equal(changes.at(-1), null, 'old panel closed immediately');
  timing.fireNext();
  assert.equal(changes.at(-1)?.owner, second);
});

test('moving to another owner before the old delay expires opens only the new owner', () => {
  const { changes, timing, controller } = harness();
  const first = Symbol('chest');
  const second = Symbol('neck');
  controller.enter(record(first, { itemName: 'Chest' }), 'pointer');
  assert.equal(timing.pending(), 1);
  controller.enter(record(second, { itemName: 'Neck' }), 'pointer');
  timing.fireNext();
  assert.equal(changes.at(-1)?.owner, second);
  assert.equal(changes.filter((value) => value?.owner === first).length, 0);
});

test('moving between two triggers of the same item keeps one panel with the latest anchor', () => {
  const { changes, timing, controller } = harness();
  const owner = Symbol('chest');
  const iconAnchor = { icon: true };
  const nameAnchor = { name: true };
  controller.enter(record(owner, { itemName: 'Chest' }, iconAnchor), 'pointer');
  timing.fireNext();
  controller.enter(record(owner, { itemName: 'Chest' }, nameAnchor), 'pointer');
  assert.equal(changes.at(-1)?.owner, owner);
  assert.equal(changes.at(-1)?.anchor, nameAnchor);
  assert.equal(changes.length, 2, 'no intermediate close when switching triggers');
});

test('the old owner cannot close the new panel from its pending close', () => {
  const { changes, timing, controller } = harness();
  const first = Symbol('chest');
  const second = Symbol('neck');
  controller.enter(record(first, { itemName: 'Chest' }), 'pointer');
  timing.fireNext();
  assert.equal(changes.at(-1)?.owner, first);
  controller.leave(first, 'pointer');
  assert.equal(timing.pending(), 1, 'close scheduled for the old owner');
  controller.enter(record(second, { itemName: 'Neck' }), 'focus');
  assert.equal(changes.at(-1)?.owner, second);
  assert.equal(timing.pending(), 0, 'new owner cleared the stale close');
});

test('release closes only the associated owner', () => {
  const { changes, controller } = harness();
  const owner = Symbol('chest');
  const other = Symbol('neck');
  controller.enter(record(owner, { itemName: 'Chest' }), 'focus');
  controller.release(other);
  assert.equal(changes.at(-1)?.owner, owner);
  controller.release(owner);
  assert.equal(changes.at(-1), null);
});

test('destroy closes and clears pending timers', () => {
  const { changes, timing, controller } = harness();
  const owner = Symbol('chest');
  controller.enter(record(owner, { itemName: 'Chest' }), 'pointer');
  assert.equal(timing.pending(), 1);
  controller.destroy();
  assert.equal(timing.pending(), 0);
  assert.equal(changes.length, 0, 'pending open never emitted after destroy');
});
