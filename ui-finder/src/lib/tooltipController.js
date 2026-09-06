export const ITEM_TOOLTIP_CONTEXT = Symbol('item-tooltip');

// App-scoped hover/focus coordination for the armory item tooltip.
//
// onChange receives null (closed) or the active record:
//   { owner, id, item, variant, preferredSide, anchor }
// owner is a Symbol per trigger wrapper; anchor is the actual hovered or
// focused element used for positioning.
//
// Behavior:
//   - pointer opens after `delay`; focus opens immediately
//   - leaving closes after `delay` unless trigger input or panel hover holds it
//   - a new owner closes the visible owner immediately; its own delay applies
//   - Escape closes and suppresses the still-active sources until they leave/re-enter
//   - release/destroy cancel associated timers and close their panels
export function createTooltipController(onChange, timing = {}) {
  const setTimer = timing.setTimer ?? setTimeout;
  const clearTimer = timing.clearTimer ?? clearTimeout;
  const delay = timing.delay ?? 120;

  let generation = 0;
  let active = null;
  let pointerOwner = null;
  let focusOwner = null;
  let panelOwner = null;
  const suppressed = new Set();
  let openTimer = null;
  let closeTimer = null;

  const emit = (value) => onChange(value);

  function clearOpen() {
    if (openTimer !== null) {
      clearTimer(openTimer);
      openTimer = null;
    }
  }

  function clearClose() {
    if (closeTimer !== null) {
      clearTimer(closeTimer);
      closeTimer = null;
    }
  }

  function scheduleOpen(record) {
    clearOpen();
    const wasActive = active;
    const gen = ++generation;
    openTimer = setTimer(() => {
      openTimer = null;
      if (gen !== generation || suppressed.has(record.owner)) return;
      if (pointerOwner !== record.owner) return;
      active = record;
      emit(active);
    }, delay);
    // An open panel replaced by a pointer entry was already closed before
    // scheduling; nothing stale can reopen it.
    return wasActive;
  }

  function scheduleClose() {
    clearClose();
    const gen = ++generation;
    closeTimer = setTimer(() => {
      closeTimer = null;
      if (gen !== generation || !active) return;
      if (pointerOwner === active.owner || focusOwner === active.owner || panelOwner === active.owner) return;
      active = null;
      emit(null);
    }, delay);
  }

  return {
    enter(record, source) {
      if (!record?.owner || !record?.anchor) return;
      clearClose();
      if (source === 'pointer') {
        pointerOwner = record.owner;
        if (suppressed.has(record.owner)) return;
        if (active && active.owner === record.owner) {
          // Same item via icon or name: keep one panel, use the latest anchor.
          active = record;
          emit(active);
          return;
        }
        if (active) {
          // New owner closes the visible old owner immediately.
          active = null;
          emit(null);
        }
        scheduleOpen(record);
      } else if (source === 'focus') {
        focusOwner = record.owner;
        if (suppressed.has(record.owner)) return;
        clearOpen();
        active = record;
        emit(active);
      }
    },
    leave(owner, source) {
      if (source === 'pointer' && pointerOwner === owner) {
        pointerOwner = null;
        if (openTimer !== null) clearOpen();
      }
      if (source === 'focus' && focusOwner === owner) focusOwner = null;
      suppressed.delete(owner);
      if (active?.owner === owner) scheduleClose();
    },
    enterPanel(owner) {
      if (!active || active.owner !== owner) return;
      panelOwner = owner;
      clearClose();
    },
    leavePanel(owner) {
      if (panelOwner !== owner) return;
      panelOwner = null;
      if (active?.owner === owner) scheduleClose();
    },
    dismiss() {
      if (pointerOwner) suppressed.add(pointerOwner);
      if (focusOwner) suppressed.add(focusOwner);
      generation++;
      clearOpen();
      clearClose();
      if (active) {
        active = null;
        emit(null);
      }
    },
    release(owner) {
      suppressed.delete(owner);
      if (pointerOwner === owner) pointerOwner = null;
      if (focusOwner === owner) focusOwner = null;
      if (panelOwner === owner) panelOwner = null;
      if (active?.owner === owner) {
        generation++;
        clearOpen();
        clearClose();
        active = null;
        emit(null);
      }
    },
    destroy() {
      generation++;
      clearOpen();
      clearClose();
      pointerOwner = null;
      focusOwner = null;
      panelOwner = null;
      suppressed.clear();
      if (active) {
        active = null;
        emit(null);
      }
    },
  };
}
