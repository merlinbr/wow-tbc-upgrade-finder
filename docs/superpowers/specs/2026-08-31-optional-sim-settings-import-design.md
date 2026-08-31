# Optional Simulation Settings Import

## Goal

Accept current valid individual-sim exports that omit `IndividualSimSettings.settings` while preserving the existing ranking behavior for exports that include it.

## Evidence

The supplied Retribution Paladin export decodes as an API-v14 individual configuration with player, 17-slot equipment, encounter, buffs, consumables, rotation, and talents. It has no `settings` protobuf message. `upgrades.Import` currently rejects that missing optional message before its existing nil-safe simulation request construction can run.

## Design

`settings` becomes optional at import. Validation still requires the player, encounter, and equipped gear that ranking needs. Character-summary fields derived from missing simulation settings use protobuf zero values: phase and iterations `0`, fixed seed `false`.

The import API defaults the maximum content phase to `5` when the imported phase is less than `1`. Exports that specify a positive phase retain it. This matches the frontend's existing fallback and avoids accidentally sending an unbounded phase (`0`) to ranking.

No decoded payload is exposed to the browser. Accepting the valid export removes the observed error; malformed and incompatible exports retain their existing typed validation responses.

## Verification

Add a focused import test built from the supplied link. It must assert successful import, the Retribution Paladin identity, and absent simulation-settings summary values. Add or update the API default assertion for maximum phase `5`. Run the focused upgrades Go tests and import the supplied link through the local HTTP endpoint.

## Boundaries

No protocol migration, ranking-domain change, new diagnostics endpoint, or UI redesign.