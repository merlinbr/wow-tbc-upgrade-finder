<script>
  import { label } from './labels.js';

  let { character, meta = null, onedit = () => {} } = $props();

  function specName() {
    const known = meta?.specs?.[character.class]?.find((spec) => spec.key === character.spec);
    if (known) {
      return known.name;
    }
    const parts = character.spec.split('_');
    if (parts[0] === character.class) {
      parts.shift();
    }
    return label(parts.join('_'));
  }
</script>

<header class="identity-header">
  <div>
    <h1 class="character-name class-{character.class}">{character.name}</h1>
    <p class="identity-subtitle">
      Level {character.level} · {label(character.race)} · {label(character.class)} · {specName()}
    </p>
    {#if character.professions?.length}
      <p class="muted identity-professions">
        {character.professions.map(label).join(' · ')}
      </p>
    {/if}
  </div>
  <button type="button" class="secondary-button" onclick={onedit}>Edit character</button>
</header>
