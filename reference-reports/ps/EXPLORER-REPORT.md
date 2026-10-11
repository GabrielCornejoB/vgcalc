> **Reference note (moved 2026-10-07):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/ps`. Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Explorer Report — ps

## Stack
- **Language(s):** TypeScript (primary), JavaScript/Node (build scripts), JSON data.
- **Framework(s):** None in the web-app sense. This is the `pkmn/ps` monorepo — Pokémon Showdown modularized into `@pkmn/*` libraries (simulator, dex, protocol parsers, client battle engine, UI helpers). Uses esbuild + tsup for bundling; jest (and mocha for `sim`).
- **Package manager / build:** npm (root `package.json`, `private: true`, `package-lock=false`). Multi-package managed via [`subpkg`](package.json) with a `subPackages` list. Shared root `build` / `import` Node scripts drive generation/bundling.
- **Scripts (build/test/lint/dev):**
  - Root: `lint`, `fix`, `compile`, `build`, `start`, `test` (all `subpkg run ...`), `test:integration`, `postinstall` (`subpkg install` + integration install).
  - Per-package pattern: `compile` (`tsc -p .`), `build` (`npm run compile && node ../build ...`), `test` (jest, or mocha for `sim`), `lint` (`eslint`). `pretest` usually builds; `posttest` lints.
  - Integration: `start` (parcel dev server), `build`/`gh-pages` (parcel), `test` (jest), `compile` (tsc + vendored showdown build).
  - `sim` uses mocha: `mocha test/main.js test/sim/*.js ...`.

## Entry-points
- Library entry files (each package): `sim/` (tsup/esbuild bundle), `dex/index.ts`, `data/index.ts`, `sets/src/index.ts`, `protocol/src/index.ts`, `view/src/index.ts`, `client/src/index.ts`, `img/src/index.ts`, `login/index.ts`, `streams/index.ts`, `mods/src/*/index.ts`, `randoms/src/index.ts`, `types/index.d.ts`.
- Build tooling: `./build` (esbuild/tsup bundle driver), `./import` (submodule sync + code generation from vendored Showdown), `./publish`.
- HTML/UI demos: `integration/index.html`, `img/index.html`, `login/index.html`.
- Integration test harness: `integration/test` (`test sim`, `test client`), UI demo `integration/src/ui/index.ts`.
- CI: `.github/workflows/test.yml` (checkout submodules → `npm install` → `npm run test:integration`).

## Top-level structure
- `client/` — `@pkmn/client`: fork of Showdown client's battle engine (battle, field, side, pokemon, handler).
- `data/` — `@pkmn/data`: higher-level data API wrapper over dex/sim.
- `dex/` — `@pkmn/dex`: unified Showdown data layer (`data/*.json`, `index.ts`); `dex/types/` — `@pkmn/dex-types` TypeScript defs.
- `img/` — `@pkmn/img`: sprite/icon display logic (`src/data/data.json`, `import-data`).
- `integration/` — cross-package integration tests + browser UI demo (parcel).
- `login/` — `@pkmn/login`: authentication against Pokémon Showdown servers.
- `mods/` — `@pkmn/mods`: non-standard format/mod modifications (gen1–gen9 variants like `gen9dlc1`, `champions`).
- `protocol/` — `@pkmn/protocol`: PROTOCOL/SIM-PROTOCOL parsing + `verifier`.
- `randoms/` — `@pkmn/randoms`: Random Battle team generation per generation (`gen1.ts`…`gen9.ts`).
- `sets/` — `@pkmn/sets`: import/export of Showdown set & team specs.
- `sim/` — `@pkmn/sim`: generated simulator extraction (`sim/`, `data/`, `lib/`, `config/`, `test/`).
- `streams/` — `@pkmn/streams`: generated extraction of Showdown's streams library.
- `types/` — `@pkmn/types`: shared TypeScript type definitions.
- `view/` — `@pkmn/view`: helpers for building Showdown client UIs (choices, formatter, `format-battle`).
- `vendor/` — git submodules `pokemon-showdown` and `pokemon-showdown-client` (currently **empty/not checked out**).
- `img/` — (also) sprite data + import tooling.
- `.github/workflows/` — CI (`test.yml`).
- Root files: `README.md`, `CONTRIBUTING.md`, `FAQ.md`, `LICENSE`, `package.json`, `tsconfig.json`, `tsconfig.test.json`, `eslint.config.mjs`, `.gitmodules`, `.npmrc`, `.gitignore`, `build`, `import`, `publish`.

## Config & tooling
- TypeScript config: root `tsconfig.json` (strict, `module: esnext`, `moduleResolution: bundler`, `noEmit: true`, target es2018) plus per-package `tsconfig.json` and root `tsconfig.test.json`.
- ESLint: root `eslint.config.mjs` re-exporting `@pkmn/eslint-config`; most packages have their own `eslint.config.mjs`.
- Jest (root dep `ts-jest`, `jest`, `@types/jest`); `sim` uses `mocha` via its own package scripts.
- Bundling: `esbuild` + `tsup` (driven by `./build`), `terser` for minification; `parcel` in `integration`.
- Tooling helpers: `subpkg` (monorepo runner), `minimist`, `json-stringify-pretty-compact`.
- `.npmrc`: `package-lock=false`, `loglevel=silent`.

## Notes / unknowns
- The workspace directory is **not a git repo** (`git` commands exit 128; no `.git`). `.gitmodules` references `vendor/pokemon-showdown` and `vendor/pokemon-showdown-client`, but both vendor dirs are currently empty — `sim`/`dex`/`randoms`/`mods` generation and `integration` tests that need the vendored Showdown will require `git submodule update`.
- `node_modules` is absent, so packages are not installed/built. Many `pretest` steps build first.
- Exact per-file logic intentionally not examined (out of scope for this shallow pass).
