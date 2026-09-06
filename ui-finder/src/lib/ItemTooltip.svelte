<script>
  import { buildItemTooltip } from './itemTooltip.js';
  import TooltipIcon from './TooltipIcon.svelte';

  let { item, variant = 'full', id = '' } = $props();

  let vm = $derived.by(() => buildItemTooltip(item ?? {}, variant));
</script>

{#if vm.name}
  <div class="item-tooltip" role="tooltip" {id}>
    <div class="tooltip-icon-slot">
      <TooltipIcon icon={vm.icon} name={vm.name} size={38} />
    </div>
    <div class="tooltip-body">
      <div class="tooltip-header">
        <div class="tooltip-inline-icon" aria-hidden="true">
          <TooltipIcon icon={vm.icon} name={vm.name} size={38} />
        </div>
        <div class="tooltip-name">
          <span class="tooltip-name-text quality-text-{vm.quality}">{vm.name}</span>
          {#if vm.suffixLabel}<span class="tooltip-suffix">{vm.suffixLabel}</span>{/if}
        </div>
        {#if vm.phase}<span class="tooltip-phase">Phase {vm.phase}</span>{/if}
      </div>
      {#if variant === 'summary'}
        <div class="tooltip-section"><span class="tooltip-muted">{vm.qualityLabel}</span></div>
      {:else}
        {#if vm.ilvl}
          <div class="tooltip-section"><span class="tooltip-ilvl">Item Level {vm.ilvl}</span></div>
        {/if}
        {#if vm.slotLabel || vm.typeLabel}
          <div class="tooltip-section tooltip-slot-row">
            <span class="tooltip-muted">{vm.slotLabel}</span>
            {#if vm.typeLabel}<span class="tooltip-muted">{vm.typeLabel}</span>{/if}
          </div>
        {/if}
        {#if vm.weaponLines.length}
          <div class="tooltip-section tooltip-weapon">
            {#each vm.weaponLines as line, index (index)}
              <span>{line}</span>
            {/each}
          </div>
        {/if}
        {#if vm.baseLines.length}
          <div class="tooltip-section">
            {#each vm.baseLines as line, index (index)}
              <span class="tooltip-stat">{line.text}</span>
            {/each}
          </div>
        {/if}
        {#if vm.enchantLine}
          <div class="tooltip-section"><span class="tooltip-enchant">{vm.enchantLine}</span></div>
        {/if}
        {#if vm.sockets.length}
          <div class="tooltip-section">
            {#each vm.sockets as socket, index (index)}
              <div class="tooltip-socket" data-gem-id={socket.gem?.id ?? ''}>
                {#if socket.gem}
                  <TooltipIcon icon={socket.gem.icon} name={socket.gem.name} size={16} />
                  <span class="tooltip-gem">{socket.text}</span>
                {:else}
                  <span class="empty-socket-icon socket-{socket.color}" aria-hidden="true"></span>
                  <span class="tooltip-muted">{socket.text}</span>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
        {#if vm.socketBonus}
          <div class="tooltip-section">
            <span class="tooltip-socket-bonus" class:inactive={!vm.socketBonus.active}>Socket Bonus: {vm.socketBonus.text}</span>
          </div>
        {/if}
        {#if vm.restrictionLines.length}
          <div class="tooltip-section">
            {#each vm.restrictionLines as line, index (index)}
              <span class="tooltip-restriction">{line}</span>
            {/each}
          </div>
        {/if}
        {#if vm.equipLines.length}
          <div class="tooltip-section">
            {#each vm.equipLines as line, index (index)}
              <span class="tooltip-equip">{line.text}</span>
            {/each}
          </div>
        {/if}
        {#if vm.setName}
          <div class="tooltip-section"><span class="tooltip-set">{vm.setName}</span></div>
        {/if}
      {/if}
    </div>
  </div>
{/if}

<style>
  .item-tooltip {
    display: grid;
    grid-template-columns: 38px minmax(0, 320px);
    gap: 5px;
    text-align: left;
    width: max-content;
    max-width: calc(100vw - 16px);
  }
  .tooltip-icon-slot { align-self: start; }
  .tooltip-body {
    box-sizing: border-box;
    min-width: 0;
    padding: 8px 10px;
    color: #f2f2f4;
    background: rgba(18, 21, 38, 0.98);
    border: 1px solid #85848b;
    border-radius: 3px;
    box-shadow: 0 4px 14px #0008;
    font-size: 13px;
    line-height: 1.25;
    max-height: calc(100vh - 16px);
    overflow-y: auto;
  }
  .tooltip-header {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 12px;
  }
  .tooltip-inline-icon { display: none; }
  .tooltip-name { min-width: 0; overflow-wrap: anywhere; }
  .tooltip-name-text { font-size: 15px; }
  .tooltip-suffix { color: #d5dae2; }
  .tooltip-phase { color: #aaaab4; white-space: nowrap; }
  .tooltip-ilvl, .tooltip-set { color: #ffd100; }
  .tooltip-enchant, .tooltip-gem, .tooltip-equip, .tooltip-socket-bonus { color: #1eff00; }
  .tooltip-muted, .tooltip-restriction, .tooltip-socket-bonus.inactive { color: #9d9da8; }
  .tooltip-section + .tooltip-section { margin-top: 8px; }
  .tooltip-section { display: grid; gap: 2px; }
  .tooltip-weapon { display: flex; gap: 8px; }
  .tooltip-slot-row { display: flex; gap: 8px; justify-content: space-between; }
  .tooltip-socket { display: flex; align-items: flex-start; gap: 4px; }
  .empty-socket-icon {
    border-radius: 2px;
    display: inline-block;
    height: 16px;
    width: 16px;
    border: 1px solid rgba(255, 255, 255, 0.35);
  }
  .empty-socket-icon.socket-0 { background: transparent; }
  .empty-socket-icon.socket-1 { background: #4a3a6b; }
  .empty-socket-icon.socket-2 { background: #b32424; }
  .empty-socket-icon.socket-3 { background: #2456b3; }
  .empty-socket-icon.socket-4 { background: #d6c520; }
  .empty-socket-icon.socket-5 { background: #2e9e44; }
  .empty-socket-icon.socket-6 { background: #d67a20; }
  .empty-socket-icon.socket-7 { background: #7a2e9e; }
  .empty-socket-icon.socket-8 { background: #c8ccd4; }

  @media (max-width: 420px) {
    .item-tooltip { display: block; max-width: calc(100vw - 16px); }
    .tooltip-icon-slot { display: none; }
    .tooltip-header {
      grid-template-columns: 38px minmax(0, 1fr) auto;
    }
    .tooltip-inline-icon { display: block; }
  }
</style>
