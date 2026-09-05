# TBC Upgrade Finder

A local browser application that ranks practical single-item DPS upgrades for one The Burning Crusade character. It imports an individual-sim configuration and evaluates candidates with the bundled [wowsims/tbc-new](https://github.com/wowsims/tbc-new) simulator and item data.

## Quick start

```bash
rtk go run ./cmd/wowsimcli rank-upgrades
```

The command starts a loopback-only local server and opens the application in your browser. Add `--no-browser` when you only need the URL printed to the terminal.

For frontend changes, use the Vite dev server with hot reload instead of
rebuilding the embedded UI: [UI development (hot reload)](docs/upgrade-finder.md#ui-development-hot-reload).

## Use it

1. Paste an individual-sim wowsims export link.
2. Review the generated 17-slot armory.
3. Choose content filters and simulation iteration budgets.
4. Start ranking and inspect the DPS upgrade report.

## What it does

Rankings cover practical single-item DPS upgrades. The imported baseline is never changed. Character data, jobs, and reports remain in process memory only; nothing is persisted remotely.

## Documentation

- [Upgrade Finder guide](docs/upgrade-finder.md)
- [Installation Guide](docs/installation.md)
- [Development Commands](docs/commands.md)
- [Adding a New Sim](docs/adding_sim.md)
- [Internationalization](docs/i18n_guide.md)

## Upstream simulator

The bundled simulator and item data come from [wowsims/tbc-new](https://github.com/wowsims/tbc-new). See [UPSTREAM.md](UPSTREAM.md) for the pinned simulator and database revisions.

The upstream project is licensed under MIT and requests a visible link back to the original project when its software is used elsewhere.

- [Live TBC sims](https://wowsims.com/tbc)
- [Join the upstream Discord](https://discord.gg/jJMPr9JWwx)
- [Support wowsims on Patreon](https://www.patreon.com/wowsims)
