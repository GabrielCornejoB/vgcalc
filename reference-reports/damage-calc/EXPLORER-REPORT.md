> **Reference note (moved 2026-10-07):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/damage-calc`. Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Explorer Report — damage-calc

## Stack
- Language(s): TypeScript (core library + tooling), vanilla JavaScript + HTML/CSS (UI)
- Framework(s): None (plain DOM/jQuery for UI); Express used only by optional `server.js`
- Package manager / build: npm (four `package.json` workspaces via `subpkg`); `calc/` bundled with Babel + Terser (`bundler.js`), compiled with `tsc`
- Scripts (build/test/lint/dev):
  - Root: `lint`, `fix`, `compile`, `bundle`, `build`, `build-ui`, `test`, `postinstall` (`subpkg run ...` + `node build view`)
  - `calc/`: `compile` (tsc -p .), `bundle` (node bundle), `build`, `test` (jest), `coverage`, `lint`/`fix`
  - `import/`: `compile` (tsc -p .), `build`, `lint`/`fix` (test is a no-op)
  - `smoketest/`: `compile` (tsc -p .), `build`, `lint`/`fix` (test is a no-op)
  - Dev UI: `node build` (full) or `node build view` (UI only, copies to `dist/`)

## Entry-points
- `calc/src/index.ts` — public `@smogon/calc` entry (exports calculate/Pokemon/Move/Field/Result)
- `calc/src/adaptable.ts` — adapter entry point for pluggable data layers
- `calc/src/data/index.ts` — bundled Generations data layer
- `src/index.html` — main UI page (plus `randoms.html`, `honkalculate.html`, `oms.html`, `champions.html`, `shared.template.html`)
- `server.js` — optional Express API (`GET /calculate`, `POST`-style body; serves `dist/` on port 3000)
- `build` — Node build script (compiles templates, copies `calc/` + `src/` into `dist/`, adds cache-busted hashes)
- `bundler.js` — Babel/Terser bundling helper used by `calc/`
- `import/src/set-import.ts` — set/usage-stat import tool (compiles to `dist/set-import.js`)
- `smoketest/src/smoketest.ts` — simulation-based smoke tester against `@pkmn/sim`

## Top-level structure
- `calc/` — core `@smogon/calc` TypeScript package: data, game mechanics per gen, calc/desc/result, Jest tests
- `src/` — vanilla JS + HTML/CSS UI layer; `js/` holds controls and per-gen sets/tiers data, `css/` and `img/` assets
- `import/` — internal tooling package for populating sets and random-battle data from Pokémon Showdown
- `smoketest/` — internal tooling package that runs battles and validates calc output
- `.github/` — CI: `workflows/test.yml` (build/test + GitHub Pages publish) and `update.yml`
- `publish` — shell script to build and push `dist/` to the `gh-pages` branch
- `build` — Node build orchestration script (template hashing, copying, compile of `calc/`)
- `bundler.js` — Babel/Terser bundler config/helper
- `server.js` — optional Express `/calculate` API
- `Dockerfile`, `.dockerignore` — container build (node:12, runs `node build`, then `server.js`)

## Config & tooling
- `tsconfig.json` (root) — strict, CommonJS, ES3 target, babel-transpiled output
- `calc/tsconfig.json`, `import/tsconfig.json`, `smoketest/tsconfig.json` — per-package TS configs
- `calc/jest.config.js` — jest/ts-jest test config
- `.eslintrc`, `.eslintignore`, `.editorconfig` — lint/style rules
- `package-lock.json` in root, `calc/`, `import/`, `smoketest/`
- `LICENSE` (MIT), `README.md`, `TASKS.md` — docs/instructions for set updates and releases

## Notes / unknowns
- Not a git repository (`.git` absent; `git log` unavailable).
- This is the official smogon/damage-calc repository (Pokémon Showdown damage calculator) — README confirms it houses both `@smogon/calc` and the UI.
- `calc/` is the only root `subPackage`; `import/` and `smoketest/` require separate `npm install` and are not built by the root default flow.
- UI is statically built to `dist/` (not committed); a `dist/` may or may not exist locally — not listed at top level.
