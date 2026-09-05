<script>
  import { formatStatLine, qualityLabel, socketColors } from './labels.js';

  let { item, variant = 'full' } = $props();

  let iconFailed = $state(false);
  let summary = $derived(variant === 'summary');
  let name = $derived(item?.itemName ?? item?.name ?? '');
  let statEntries = $derived(Object.entries(item?.stats ?? {}));
  let suffixEntries = $derived(item?.randomSuffix ? Object.entries(item.randomSuffix.stats ?? {}) : []);

  function iconUrl(icon) {
    return `https://wow.zamimg.com/images/wow/icons/large/${icon}.jpg`;
  }

  function gemText(gem) {
    const lines = Object.entries(gem?.stats ?? {}).map(([key, value]) => formatStatLine(key, value));
    return lines.length ? lines.join(', ') : (gem?.name ?? '');
  }
</script>

{#if name}
  <span class="item-tooltip" role="tooltip">
    {#if summary}
      <span class="tooltip-summary">
        <span class="tooltip-name quality-text-{item.quality ?? 0}">
          {#if item.icon && !iconFailed}
            <img class="tooltip-icon" src={iconUrl(item.icon)} alt={name} onerror={() => (iconFailed = true)} />
          {/if}
          {name}
        </span>
        {#if item.phase}<span class="tooltip-phase">Phase {item.phase}</span>{/if}
      </span>
      <span class="tooltip-meta">{qualityLabel(item.quality)}</span>
    {:else}
      <span class="tooltip-header">
        <span class="tooltip-name quality-text-{item.quality ?? 0}">{name}</span>
        {#if item.phase}<span class="tooltip-phase">Phase {item.phase}</span>{/if}
      </span>
      {#if item.ilvl}<span class="tooltip-ilvl">Item Level {item.ilvl}</span>{/if}
      <span class="tooltip-meta">{item.slotName}</span>
      {#each statEntries as [key, value]}
        <span class="tooltip-stat">{formatStatLine(key, value)}</span>
      {/each}
      {#if item.randomSuffix}
        <span class="tooltip-meta">{item.randomSuffix.name}</span>
        {#each suffixEntries as [key, value]}
          <span class="tooltip-stat">{formatStatLine(key, value)}</span>
        {/each}
      {/if}
      {#if item.sockets?.length}
        <span class="tooltip-sockets">
          {#each item.sockets as socket}
            <span class="tooltip-socket">
              <span class="socket-dot socket-{socket.color}" title="{socketColors[socket.color] ?? 'Unknown'} socket"></span>
              {#if socket.gem}
                <span class="tooltip-gem">{gemText(socket.gem)}</span>
              {:else}
                <span class="tooltip-gem empty">{socketColors[socket.color] ?? 'Unknown'} socket (empty)</span>
              {/if}
            </span>
          {/each}
        </span>
      {/if}
      {#if item.socketBonus?.stats && Object.keys(item.socketBonus.stats).length}
        <span class="tooltip-socket-bonus" class:inactive={!item.socketBonus.active}>
          Socket Bonus: {Object.entries(item.socketBonus.stats).map(([key, value]) => formatStatLine(key, value)).join(', ')}
        </span>
      {/if}
      {#if item.enchant}
        <span class="tooltip-enchant">Equip: {item.enchant.description || item.enchant.name}</span>
      {/if}
    {/if}
  </span>
{/if}

<style>
  .item-tooltip {
    display: none;
    position: absolute;
    z-index: 30;
    top: 0;
    left: calc(100% + 10px);
    width: max-content;
    max-width: 280px;
    flex-direction: column;
    gap: 2px;
    padding: 8px 10px;
    background: #0a0d14;
    border: 1px solid #2b3444;
    border-radius: 6px;
    color: #f2f4f8;
    font-size: 0.85rem;
    line-height: 1.35;
    pointer-events: none;
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.55);
  }
  .tooltip-header { display: flex; justify-content: space-between; gap: 12px; }
  .tooltip-icon { border: 1px solid #2b3444; border-radius: 3px; height: 18px; object-fit: cover; vertical-align: middle; width: 18px; }
  .tooltip-phase { color: #b8c0cc; white-space: nowrap; }
  .tooltip-ilvl { color: #e6b23c; }
  .tooltip-meta { color: #d5dae2; }
  .tooltip-stat { color: #ffffff; }
  .tooltip-sockets { display: flex; flex-direction: column; gap: 2px; }
  .tooltip-socket { display: flex; align-items: center; gap: 6px; }
  .tooltip-gem { color: #3fd13f; }
  .tooltip-gem.empty { color: #8a93a3; }
  .socket-dot { width: 10px; height: 10px; border-radius: 2px; display: inline-block; border: 1px solid rgba(255, 255, 255, 0.35); }
  .socket-dot.socket-1 { background: #4a3a6b; }
  .socket-dot.socket-2 { background: #b32424; }
  .socket-dot.socket-3 { background: #2456b3; }
  .socket-dot.socket-4 { background: #d6c520; }
  .socket-dot.socket-5 { background: #2e9e44; }
  .socket-dot.socket-6 { background: #d67a20; }
  .socket-dot.socket-7 { background: #7a2e9e; }
  .socket-dot.socket-8 { background: #c8ccd4; }
  .tooltip-socket-bonus { color: #3fd13f; }
  .tooltip-socket-bonus.inactive { color: #8a93a3; }
  .tooltip-enchant { color: #3fd13f; }
</style>
