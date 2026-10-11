> **Reference note (moved 2026-10-07):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/engine`. Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Explorer Report — engine

`@pkmn/engine` (pkmn/engine) — "A minimal, complete, Pokémon battle simulation engine optimized for performance". WIP (pre-v0.1). Not a git repository (no `.git`).

## Stack
- Language(s): Zig (core engine, `build.zig` / `src/lib`, min Zig 0.16.0); TypeScript (JS bindings/package, tooling, tests); C (C bindings/example, `src/include/pkmn.h`); small amount of JSX per tsconfig and TSX tool sources.
- Framework(s): Zig build system; Node-API / WASM bindings (napi.h); Vitest (test runner); Vite (example web app); VS Code extension tooling; ESLint via `@pkmn/eslint-config`; Parcel (optional example bundler).
- Package manager / build: npm (Node `>=16`, `package.json`, `package-lock` not inspected). Orchestrated by `Makefile` which drives both `zig build` and npm.
- Scripts (build/test/lint/dev):
  - npm: `lint`, `fix`, `compile` (= `build`, tsc), `test` (vitest src/pkg), `test:integration`, `debug`, `demo`, `fuzz`, `generate`, `release`, `benchmark`, `integration`, `extension`, `postinstall` (install-pkmn-engine).
  - Makefile: `build` (zig-build + js-build), `test` (zig-test + js-test), `lint`, `fix`, `coverage`, `check`, `example` (c/js/zig), `addon`, `integration`, `benchmark`, `release`, `patch`, `clean`, `t`/`it`.
  - zig build: `zig build --summary all -Dlog -Dchance -Dcalc [-Dshowdown] [-p build] [test]`.

## Entry-points
- Package main: `build/pkg/index.js`, types `build/pkg/index.d.ts` (built from `src/pkg/index.ts`).
- npm bins: `install-pkmn-engine` → `src/bin/install-pkmn-engine`; `pkmn-debug` → `src/bin/pkmn-debug`.
- Zig core: `src/lib/pkmn.zig` (+ bindings `src/lib/node.zig`, `src/lib/wasm.zig`, `src/lib/c.zig`).
- Build entry: `build.zig` (root), `build.zig.zon` (package manifest).
- Example apps: `examples/c/example.c`, `examples/js/example.ts` + `index.html`, `examples/zig/example.zig`.
- Print/debug tools (built): `build/tools/debug`, `build/tools/demo`, `build/tools/generate`, `build/tools/release`.
- Test entries: `src/pkg/*.test.ts` (unit), `src/test/*.test.ts` (integration/benchmark/fuzz/regression).

## Top-level structure
- `src/` — source root: `lib/` (Zig engine), `pkg/` (TS npm package), `bin/`, `tools/`, `test/`, `data/`, `include/`.
- `examples/` — runnable samples: `c/` (C Makefile example), `js/` (Vite/Parcel web+node example), `zig/` (Zig build example).
- `docs/` — `DESIGN.md`, `PROTOCOL.md`, `TESTING.md`, `CONTRIBUTING.md`, `RESEARCH.md`, `NOTES.md`.
- `.github/` — CI workflows: `test.yml`, `fuzz.yml`, `release.yml`.
- `README.md` — project overview/status (WIP warning).
- `Makefile` — unified Zig + npm orchestration.
- `package.json` — npm package `@pkmn/engine`, scripts, deps.
- `build.zig` / `build.zig.zon` — Zig build/config.
- `tsconfig.json`, `vite.config.mts`, `eslint.config.mjs` — TS/Vitest/ESLint config.
- `CITATION.cff`, `LICENSE`, `.gitattributes`, `.gitignore`.

### `src/` subdirs
- `src/lib/` — Zig engine: `common/`, `gen1/`, `gen2/`, `bindings/`, plus `pkmn.zig`, `node.zig`, `wasm.zig`, `c.zig`, `test.zig`, `napi.h`.
- `src/pkg/` — TS package surface: `index.ts`, `addon.ts`, `common.ts`, `data.ts`, `gen1.ts`, `protocol.ts`, plus `data/` JSON (`ids.json`, `layout.json`, `mapping.json`) and `fixtures/`.
- `src/bin/` — CLI node scripts `install-pkmn-engine`, `pkmn-debug`.
- `src/tools/` — dev tooling: `debug.ts`, `demo.ts`, `generate.ts`, `release.ts`, `serde.ts`, display UI (`display/`), Zig tools (`analyze.zig`, `dump.zig`, `transitions.zig`, `serde.zig`), `vscode/` extension subpackage.
- `src/test/` — integration/fuzz/benchmark/regression harness, `showdown/` fixtures, `stats.ts`.
- `src/data/` — shared JSON: `data.json`, `layout.json`, `protocol.json`.
- `src/include/` — public C header `pkmn.h`.

## Config & tooling
- `package.json` deps: `@pkmn/data`, `@pkmn/protocol`; dev deps include `@pkmn/sim`, `@pkmn/dex-types`, `@pkmn/sets`, `@pkmn/smogon`, `vitest` + `@vitest/coverage-v8`, `typescript`, `eslint` + `@pkmn/eslint-config`, `esbuild`, `binaryen`, `@vscode/vsce`, JSX/display helpers.
- `tsconfig.json`: strict, `module: nodenext`, `outDir: build`, `rootDir: src`, JSX via `h`/`Fragment`, types `jest`+`node`.
- `vite.config.mts`: Vitest config (no watch, 60s timeout; excludes node_modules/build/examples; coverage → `coverage/js`).
- `eslint.config.mjs`: flat config extending `@pkmn/eslint-config`.
- CI in `.github/workflows/` (test, fuzz, release).

## Notes / unknowns
- Package publishes only `build/` artifacts + a subset of `src/lib` Zig sources + some `src/tools` sources (see `files` in package.json).
- `main`/`types` point at `build/` which is generated by tsc + the Zig→npm install step; a clean checkout cannot run JS without building (`make build`).
- No `package-lock.json`/`node_modules` visible at repo root (lockfile state unknown).
- Version `0.1.0` but README marks project WIP pre-release; Zig `0.16.0` minimum.
- Not a git repo, so no git log/status available.
