<script>
  const backdrops = import.meta.glob('../../../assets/img/*.jpg', { eager: true, import: 'default' });

  const backdropFor = {
    BalanceDruid: 'balance_druid_background.jpg',
    FeralCatDruid: 'feral_druid_background.jpg',
    FeralBearDruid: 'feral_druid_tank_background.jpg',
    RestorationDruid: 'resto_druid_background.jpg',
    Hunter: 'hunter_background.jpg',
    Mage: 'mage_background.jpg',
    HolyPaladin: 'holy_paladin_background.jpg',
    ProtectionPaladin: 'prot_paladin.jpg',
    RetributionPaladin: 'retribution_paladin.jpg',
    Priest: 'healing_priest_background.jpg',
    Rogue: 'rogue_background.jpg',
    ElementalShaman: 'elemental_shaman_background.jpg',
    EnhancementShaman: 'enhancement_shaman_background.jpg',
    RestorationShaman: 'resto_shaman_background.jpg',
    Warlock: 'warlock_background.jpg',
    DpsWarrior: 'warrior_background.jpg',
    ProtectionWarrior: 'warrior_background.jpg',
  };

  let { race = '', class: playerClass = '', spec = '' } = $props();

  let backdropUrl = $derived.by(() => {
    const fileName = backdropFor[spec] ?? backdropFor[playerClass];
    if (!fileName) return '';
    const entry = Object.entries(backdrops).find(([path]) => path.endsWith(`/${fileName}`));
    return entry?.[1] ?? '';
  });
</script>

<div class="character-stage" data-region="character-stage">
  <div class="stage-backdrop" aria-hidden="true" style:background-image={backdropUrl ? `url('${backdropUrl}')` : 'none'}></div>
  <div class="stage-placeholder">
    <span class="stage-kicker">Character preview</span>
    <span class="stage-note">Appearance not imported</span>
    <button type="button" class="secondary-button" disabled title="3D model integration is not available yet">Activate 3D</button>
  </div>
</div>
