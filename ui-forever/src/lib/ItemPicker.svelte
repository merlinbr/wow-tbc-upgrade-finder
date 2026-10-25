<script>
  import { api } from './api.js';
  import { label, slotLabel } from './labels.js';

  let {
    open = false,
    slot = '',
    character,
    current = null,
    onclose = () => {},
    onapply = () => {},
  } = $props();

  let tab = $state('items');
  let query = $state('');
  let sourceFilter = $state('');
  let knownSources = $state([]);
  let items = $state([]);
  let itemTotal = $state(0);
  let selected = $state(null);
  let detail = $state(null);
  let variantIlvl = $state(0);
  let suffixId = $state(0);
  let enchants = $state([]);
  let selectedEnchant = $state(null);
  let error = $state('');

  let itemRequest = 0;
  let enchantRequest = 0;
  let itemSignature = '';
  let enchantSignature = '';

  $effect(() => {
    if (!open || tab !== 'items') {
      return;
    }
    const values = {
      className: character.class,
      slot,
      query,
      source: sourceFilter,
    };
    const timer = setTimeout(() => {
      loadItems(values);
    }, 150);
    return () => clearTimeout(timer);
  });

  $effect(() => {
    if (!open || tab !== 'enchants') {
      return;
    }
    const values = { className: character.class, slot, query };
    const timer = setTimeout(() => {
      loadEnchants(values);
    }, 150);
    return () => clearTimeout(timer);
  });

  async function loadItems({ className, slot: slotName, query: queryValue, source }) {
    const signature = [className, slotName, queryValue, source].join('|');
    if (signature === itemSignature) {
      return;
    }
    const sequence = ++itemRequest;
    const params = new URLSearchParams({ class: className, slot: slotName, limit: '50' });
    if (queryValue) {
      params.set('q', queryValue);
    }
    if (source) {
      params.set('source', source);
    }
    try {
      const data = await api.get(`/api/items?${params}`);
      if (sequence !== itemRequest) {
        return;
      }
      items = data.items ?? [];
      itemTotal = data.total ?? items.length;
      itemSignature = signature;
      const seen = new Set(knownSources);
      for (const item of items) {
        for (const entry of item.sources ?? []) {
          seen.add(entry.kind);
        }
      }
      knownSources = [...seen].sort();
      error = '';
    } catch (requestError) {
      if (sequence === itemRequest) {
        error = requestError.message ?? 'Could not search items.';
      }
    }
  }

  async function loadEnchants({ className, slot: slotName, query: queryValue }) {
    const signature = [className, slotName, queryValue].join('|');
    if (signature === enchantSignature) {
      return;
    }
    const sequence = ++enchantRequest;
    const params = new URLSearchParams({
      class: className,
      slot: slotName,
      includeIncompatible: '1',
      limit: '100',
    });
    if (queryValue) {
      params.set('q', queryValue);
    }
    try {
      const data = await api.get(`/api/enchants?${params}`);
      if (sequence !== enchantRequest) {
        return;
      }
      enchants = data.enchants ?? [];
      enchantSignature = signature;
      error = '';
    } catch (requestError) {
      if (sequence === enchantRequest) {
        error = requestError.message ?? 'Could not search enchants.';
      }
    }
  }

  async function selectItem(item) {
    selected = item;
    detail = null;
    variantIlvl = 0;
    suffixId = 0;
    try {
      const params = new URLSearchParams({ class: character.class, slot });
      detail = await api.get(`/api/items/${item.id}?${params}`);
    } catch {
      detail = item;
    }
  }

  function selectEnchant(enchant) {
    selectedEnchant = enchant;
  }

  let displayStats = $derived.by(() => {
    if (!detail) {
      return [];
    }
    if (variantIlvl) {
      const variant = (detail.variants ?? []).find(
        (entry) => Number(entry.ilvl) === Number(variantIlvl),
      );
      if (variant) {
        return variant.stats ?? [];
      }
    }
    return detail.stats ?? [];
  });

  let displayDamage = $derived.by(() => {
    if (!detail) {
      return null;
    }
    if (variantIlvl) {
      const variant = (detail.variants ?? []).find(
        (entry) => Number(entry.ilvl) === Number(variantIlvl),
      );
      if (variant) {
        return variant;
      }
    }
    return detail;
  });

  let itemCompatible = $derived(
    selected ? (detail?.compatible ?? selected.compatible ?? true) : false,
  );
  let itemReason = $derived(selected ? (detail?.reason ?? selected.reason ?? '') : '');

  let enchantReason = $derived.by(() => {
    if (!selectedEnchant) {
      return '';
    }
    if (!current) {
      return 'enchant_without_item';
    }
    if (!selectedEnchant.compatible) {
      return selectedEnchant.reason ?? 'incompatible_enchant_slot';
    }
    return '';
  });
  let canApplyEnchant = $derived(
    Boolean(selectedEnchant && current && selectedEnchant.compatible),
  );

  function applyItem() {
    if (!selected || !itemCompatible) {
      return;
    }
    onapply({
      itemId: selected.id,
      variantIlvl: Number(variantIlvl) || 0,
      randomSuffixId: Number(suffixId) || 0,
    });
  }

  function applyEnchant() {
    if (!selectedEnchant || !canApplyEnchant) {
      return;
    }
    onapply({
      ...current,
      enchant: {
        effectId: selectedEnchant.effectId,
        itemId: selectedEnchant.itemId || 0,
        spellId: selectedEnchant.spellId || 0,
      },
    });
  }

  function itemTypeLabel(item) {
    if (!item) {
      return '';
    }
    if (item.armorType) {
      return label(item.armorType);
    }
    if (item.weaponType) {
      return label(item.weaponType);
    }
    if (item.rangedWeaponType) {
      return label(item.rangedWeaponType);
    }
    return label(item.type);
  }

  function sourceLabel(item) {
    const kinds = (item?.sources ?? []).map((entry) => entry.kind);
    if (kinds.length === 0) {
      return 'Unknown source';
    }
    return kinds.map(label).join(', ');
  }

  function qualityClass(quality) {
    return quality ? `quality-${quality}` : '';
  }
</script>

{#if open}
  <div class="picker-backdrop">
    <dialog open class="panel picker" aria-modal="true" aria-labelledby="picker-title">
      <header class="picker-header">
        <div>
          <p class="kicker">{slotLabel(slot)}</p>
          <h2 id="picker-title">Choose an item</h2>
        </div>
        <button type="button" class="secondary-button" onclick={onclose}>Close</button>
      </header>

      <div class="picker-tabs" role="tablist" aria-label="Picker sections">
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'items'}
          onclick={() => (tab = 'items')}
        >
          Items
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'enchants'}
          onclick={() => (tab = 'enchants')}
        >
          Enchants
        </button>
      </div>

      <div class="picker-search">
        <div>
          <label for="picker-query">{tab === 'items' ? 'Search items' : 'Search enchants'}</label>
          <input
            id="picker-query"
            placeholder={tab === 'items' ? 'Search items' : 'Search enchants'}
            bind:value={query}
          />
        </div>
        {#if tab === 'items'}
          <div>
            <label for="picker-source">Source</label>
            <select id="picker-source" bind:value={sourceFilter}>
              <option value="">All sources</option>
              {#each knownSources as source (source)}
                <option value={source}>{label(source)}</option>
              {/each}
            </select>
          </div>
        {/if}
      </div>

      {#if error}
        <p class="error-text" role="alert">{error}</p>
      {/if}

      {#if tab === 'items'}
        <div class="picker-body">
          <ul class="picker-list">
            {#each items as item (item.id)}
              <li>
                <button
                  type="button"
                  class="picker-row"
                  class:selected={selected?.id === item.id}
                  data-testid="picker-row"
                  onclick={() => selectItem(item)}
                >
                  <span class="picker-row-name {qualityClass(item.quality)}">{item.name}</span>
                  <span class="picker-row-meta">
                    {#if item.ilvl}i{item.ilvl} · {/if}{itemTypeLabel(item)}
                  </span>
                  <span class="picker-row-source">{sourceLabel(item)}</span>
                  <span class="picker-row-fit" class:incompatible={!item.compatible}>
                    {item.compatible ? 'Fits' : item.reason}
                  </span>
                </button>
              </li>
            {:else}
              <li class="muted picker-empty">No items match this search.</li>
            {/each}
          </ul>
          <p class="muted picker-count">{itemTotal} items</p>

          <aside class="picker-detail">
            {#if selected && detail}
              <h3 class={qualityClass(detail.quality)}>{detail.name}</h3>
              <p class="muted">
                {itemTypeLabel(detail)}
                {#if detail.ilvl}· Item level {detail.ilvl}{/if}
              </p>

              {#if detail.variants?.length}
                <label for="variant-select">Variant</label>
                <select id="variant-select" bind:value={variantIlvl}>
                  <option value={0}>Default ({detail.ilvl})</option>
                  {#each detail.variants as variant (variant.ilvl)}
                    <option value={variant.ilvl}>Item level {variant.ilvl}</option>
                  {/each}
                </select>
              {/if}

              {#if detail.suffixIds?.length}
                <label for="suffix-select">Random suffix</label>
                <select id="suffix-select" bind:value={suffixId}>
                  <option value={0}>None</option>
                  {#each detail.suffixIds as id (id)}
                    <option value={id}>Suffix {id}</option>
                  {/each}
                </select>
              {/if}

              <h4>Stats</h4>
              {#if displayStats.length > 0}
                <ul class="stat-list">
                  {#each displayStats as stat (stat.name)}
                    <li>{label(stat.name)} <strong>{stat.value}</strong></li>
                  {/each}
                </ul>
              {:else}
                <p class="muted">No stats are recorded for this variant.</p>
              {/if}

              {#if displayDamage?.weaponDamageMin}
                <p class="muted">
                  Damage {displayDamage.weaponDamageMin}–{displayDamage.weaponDamageMax}
                  {#if detail.weaponSpeed}· Speed {detail.weaponSpeed}{/if}
                </p>
              {/if}
              {#if detail.setName}
                <p class="muted">Set: {detail.setName}</p>
              {/if}
              <p class="muted">Level requirement: unknown</p>
              {#if itemReason}
                <p class="error-text">{itemReason}</p>
              {/if}
              <button type="button" onclick={applyItem} disabled={!itemCompatible}>
                Apply to {slot}
              </button>
            {:else}
              <p class="muted">Select an item to inspect it.</p>
            {/if}
          </aside>
        </div>
      {:else}
        <div class="picker-body">
          <ul class="picker-list">
            {#each enchants as enchant (`${enchant.effectId}-${enchant.itemId ?? 0}-${enchant.spellId ?? 0}`)}
              <li>
                <button
                  type="button"
                  class="picker-row"
                  class:selected={selectedEnchant?.effectId === enchant.effectId &&
                    selectedEnchant?.itemId === enchant.itemId}
                  data-testid="enchant-row"
                  onclick={() => selectEnchant(enchant)}
                >
                  <span class="picker-row-name {qualityClass(enchant.quality)}">{enchant.name}</span>
                  <span class="picker-row-meta">Effect {enchant.effectId}</span>
                  <span class="picker-row-fit" class:incompatible={!enchant.compatible}>
                    {enchant.compatible ? 'Fits' : enchant.reason}
                  </span>
                </button>
              </li>
            {:else}
              <li class="muted picker-empty">No enchants match this search.</li>
            {/each}
          </ul>

          <aside class="picker-detail">
            {#if selectedEnchant}
              <h3 class={qualityClass(selectedEnchant.quality)}>{selectedEnchant.name}</h3>
              <p class="muted">
                Effect {selectedEnchant.effectId}
                {#if selectedEnchant.itemId}· Item {selectedEnchant.itemId}{/if}
                {#if selectedEnchant.spellId}· Spell {selectedEnchant.spellId}{/if}
              </p>
              {#if selectedEnchant.stats?.length}
                <h4>Stats</h4>
                <ul class="stat-list">
                  {#each selectedEnchant.stats as stat (stat.name)}
                    <li>{label(stat.name)} <strong>{stat.value}</strong></li>
                  {/each}
                </ul>
              {/if}
              {#if selectedEnchant.lvl}
                <p class="muted">Item level {selectedEnchant.lvl}</p>
              {/if}
              <p class="muted">Level requirement: unknown</p>
              {#if enchantReason}
                <p class="error-text">{enchantReason}</p>
              {/if}
              <button type="button" onclick={applyEnchant} disabled={!canApplyEnchant}>
                Apply enchant
              </button>
            {:else}
              <p class="muted">Select an enchant to inspect it.</p>
            {/if}
          </aside>
        </div>
      {/if}
    </dialog>
  </div>
{/if}
