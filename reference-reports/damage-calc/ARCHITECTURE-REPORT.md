> **Reference note (moved 2026-10-07):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/damage-calc`. Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Architecture Report — damage-calc

## Overview
- Monorepo with two halves: pure-TypeScript battle library `@smogon/calc` in `calc/` plus jQuery/HTML UI in `src/` statically built to `dist/` (`EXPLORER-REPORT.md:3-12`, `build:64-74`, `server.js:8-30`).
- Library is side-effect-free math: `Generations + Pokemon/Move/Field -> calculate() -> Result/desc`; UI is a thin event-driven shell that scrapes DOM into those models and re-renders text (`calc/src/calc.ts:27-41`, `src/js/shared_controls.js:1066-1265`).
- Data is table-driven and generation-layered: base tables `RBY/GSC/.../SV` plus `*_PATCH` deltas merged with `extend(true,...)`; behavior is generation-stratified with one engine per era plus shared helpers (`calc/src/data/species.ts:1562`, `calc/src/data/moves.ts:341,571,1177`, `calc/src/mechanics/util.ts:1-717`).
- No framework, no store, no router: global `calc`, `pokedex`, `moves`, `GENERATION` in browser; `MECHANICS[gen.num]` dispatch in library; Express `server.js` is an optional thin wrapper (`calc/src/index.ts:1-58`, `src/js/shared_controls.js:1415-1421`, `server.js:11-28`).
- Tooling closes the loop: `import/` regenerates `src/js/data/sets/*.js` from Showdown data/usage; `smoketest/` cross-checks calc against `@pkmn/sim`; `calc/src/test/*` pins ranges/descs per gen (`import/src/set-import.ts:465-603`, `smoketest/src/smoketest.ts:123-200`, `calc/src/test/helper.ts:51-82`).

## Core systems (responsibility + key files)
- Public facade + pluggable-data split: `calc/src/index.ts` (gen-number convenience wrapper, `Pokemon/Move` subclasses resolving `Generations.get(n)`) over `calc/src/adaptable.ts` (pure `Generation`-parameterized exports for custom data layers) (`calc/src/index.ts:46-121`, `calc/src/adaptable.ts:1-12`, `calc/src/data/index.ts:10-36`).
- Dispatcher: `calc/src/calc.ts` clones inputs and indexes `MECHANICS[gen.num]` (`calculateChampions,RBYGSC,ADV,DPP,BWXY,SMSSSV`) so callers never branch on gen (`calc/src/calc.ts:14-41`).
- Domain entities (mutable working state, `clone()` before calc): `Pokemon` (stats/IV/EV/boosts/Tera/Dynamax/formes), `Move` (Z/Max resolution, multihit, `overrides`), `Field/Side` (weather/terrain/rooms/hazards/screens) plus plain `State.*` DTO interfaces (`calc/src/pokemon.ts:10-177`, `calc/src/move.ts:7-198`, `calc/src/field.ts:4-132`, `calc/src/state.ts:4-84`).
- Generation engines: `calc/src/mechanics/gen12.ts`, `gen3.ts`, `gen4.ts`, `gen56.ts`, `gen789.ts`, `champions.ts` — each exports `calculate*` plus `calculateBasePower|BPMods|Attack|AtMods|Defense|DfMods|FinalMods` stages (`calc/src/mechanics/gen789.ts:52,785,1040,1317,1367,1524,1580,1718`).
- Shared mechanics kernel: `calc/src/mechanics/util.ts` (`computeFinalStats`, `getModifiedStat`, `getFinalSpeed`, `checkIntimidate/Download/Item/Forecast/SeedBoost/MultihitBoost`, `chainMods`, `getBaseDamage`, `getFinalDamage`, `getStabMod`, `pokeRound/OF16/OF32`, `isGrounded`) reused by all engines (`calc/src/mechanics/util.ts:37-137,168-469,471-717`).
- Stat/item math: `calc/src/stats.ts` (`calcStatADV/RBY/Champions`, Hidden Power, DV/IV conversion), `calc/src/items.ts` (`getItemBoostType`, `getBerryResistType`, `getFlingPower`, `getNaturalGift/TechnoBlast/MultiAttack`) (`calc/src/stats.ts:118-251`, `calc/src/items.ts:11-424`).
- Presentation: `calc/src/result.ts` (`Result`, `damageRange/multiDamageRange`, `range/fullDesc/moveDesc/recovery/recoil/kochance`) + `calc/src/desc.ts` (`RawDesc`, `display/buildDescription`, `getKOChance/computeKOChance/predictTotal/combine`, `getHazards/getEndOfTurn`, `getRecovery/getRecoil`, `squashMultihit`) (`calc/src/result.ts:9-118`, `calc/src/desc.ts:10-57,59-101,269-454,516-768`).
- Data layer: `calc/src/data/interface.ts` contracts (`Generation/Abilities/Items/Moves/Species/Types/Natures`) + per-kind tables (`abilities.ts`, `items.ts`, `moves.ts`, `species.ts`, `types.ts`, `natures.ts`) exposing `get(id)`/`[Symbol.iterator]` and indexed `*_BY_ID[gen]` maps (`calc/src/data/interface.ts:40-177`, `calc/src/data/abilities.ts:570-609`, `calc/src/data/types.ts:438-480`).
- UI shell: `src/index.html` (+ `randoms/honkalculate/oms/champions/shared.template.html`) static panels; `src/js/shared_controls.js` (DOM->model builders `createPokemon/createField/getMoveDetails`, stat/HP math, set/ability/item/forme autosync); per-mode `honkalculate_controls.js`, `index_randoms_controls.js`, `oms_controls.js`, `moveset_import.js`, `dark-theme-toggle.js`; vendored jQuery/select2 (`src/index.html:51-82,147-763`, `src/js/shared_controls.js:1066-1265,1267-1413`).
- Build/serve: `build` (compile `calc/`, `cpdir src+calc/dist->dist`, `makeCachebuster` template+hash), `bundler.js` (Babel+Terser), `server.js` (`GET /calculate` + `express.static('dist')`) (`build:36-74`, `server.js:1-30`).
- Data tooling: `import/src/set-import.ts` (fetch `data.pkmn.cc` dex sets + usage stats, validate with `@pkmn/sim TeamValidator`, emit `SETDEX_*`); `smoketest/src/smoketest.ts` (drive random `@pkmn/sim` battles through `@pkmn/client Battle`, compare sim damage to `calc.calculate`) (`import/src/set-import.ts:205-303,305-343,465-603`, `smoketest/src/smoketest.ts:39-121,152-200,232-301`).
- Tests: `calc/src/test/calc.test.ts` (per-gen range/desc goldens), `helper.ts` (`inGen/inGens/tests` + `toMatch` matcher), plus `data/pokemon/move/stats/utils` unit tests (`calc/src/test/calc.test.ts:6-57`, `calc/src/test/helper.ts:84-201`).

## Data flow
- Flow 1: `calculate(gen, attacker, defender, move, field)` end-to-end (`calc/src/calc.ts:27-41`, `calc/src/mechanics/gen789.ts:52-131`):
  1. Clone all inputs (engines mutate boosts/stats/types/items in place) then per-engine preamble: `checkAirLock/TeraformZero/Forecast/Item/RawStatChanges/SeedBoost/DauntlessShield/Embody`, `computeFinalStats(def,spd,spe)`, `checkIntimidate/Download/IntrepidSword/WindRider`, `computeFinalStats(atk,spa)`, `checkInfiltrator`.
  2. Early exits building `RawDesc`: Status moves, Protect (`breaksProtect` incl. Z/Max/Unseen Fist), `Pain Split`, immunities/absorptions (`Sap Sipper/Flash Fire/Dry Skin/Lightning Rod/Levitate/Bulletproof/Soundproof/...`), type-effectiveness `0` (`calc/src/mechanics/gen789.ts:127-287`).
  3. Staged pipeline: `calculateBasePower*` (weight/speed/HP-dependent BP + `BPMods` chain) -> `calculateAttack*`/`calculateDefense*` (boosts, Unaware/crit rules, weather/ability/item `At/DfMods`) -> `getBaseDamage` (level/BP/At/Df + spread/parental/weather/crit) -> 16-roll `getFinalDamage` loop (random factor, STAB, burn, protect, `chainMods(finalMods)`) -> multihit/`timesUsed` matrix via `checkMultihitBoost` re-entry (`calc/src/mechanics/gen789.ts:785-900,1317-1718`, `calc/src/mechanics/util.ts:471-559`).
  4. `new Result(gen,attacker,defender,move,field,damage,desc)`; text/KO lazily derived in `desc.ts` (`calc/src/result.ts:36-82`, `calc/src/desc.ts:59-101,269-454`).
- Flow 2: UI interaction -> calc -> render (`src/js/shared_controls.js:1066-1265,1267-1355`):
  - jQuery `calc-trigger`/`terrain-trigger` change handlers and `calcHP/calcStats` live stat preview (`calc.calcStat`) feed `createPokemon(pokeInfo)` (reads base/IV/EV/DV/boosts/nature/ability/item/status/Tera/Dynamax + `overrides:{baseStats,types}`), `getMoveDetails` (reads BP/type/cat/crit/Z/hits/times + `overrides`), `createField()` (reads format/weather/terrain/ruins/sides).
  - Calls global `calc.calculate(GENERATION, ...)` / `calc.Stats.getHiddenPower`; writes `#mainResult`, per-move `resultDamage*`, damage-values detail; autosync helpers (`autosetWeather/autosetTerrain/autosetQP`, forme/Tera/Stellar, Hidden Power IV fixup) keep DOM consistent (`src/js/shared_controls.js:314-447,473-561,882-904`).
- Flow 3: data resolution (`calc/src/data/index.ts:10-36`, `calc/src/data/species.ts:25-35,1562`, `calc/src/util.ts:1-10`):
  - `Generations.get(n)` constructs a `Generation` holding `Abilities/Items/Moves/Species/Types/Natures` for that `n`; each collection is a prebuilt `*_BY_ID[gen]` map built from layered literals (`RBY + GSC_PATCH + ADV_PATCH...`) via `extend(true,...)`.
  - Entities resolve via `toID(name)` lookups (`gen.species.get`, `gen.moves.get`) and deep-merge `options.overrides`; `Move` constructor further rewrites Z/Max (`getZMoveName/getMaxMoveName`), multihit defaults, category fallback, typeless/crit/happiness special cases (`calc/src/pokemon.ts:53-101`, `calc/src/move.ts:56-174`).
  - `adaptable.ts` lets hosts inject any `interface.Generation`; `index.ts` is just the bundled-`Generations` convenience (`calc/src/adaptable.ts:1-12`, `calc/src/index.ts:60-121`).
- Flow 4: build/serve + data refresh (`build:36-74`, `server.js:11-30`, `import/src/set-import.ts:586-603`):
  - `node build [view]` compiles `calc/`, copies `src/` + `calc/dist` to `dist/`, injects `shared.template.html` wrapper and MD5 cache-busters; `server.js` serves `dist/` and exposes `GET /calculate` constructing `calc.Pokemon/Move/Field` from JSON body.
  - Offline: `import/dist/set-import.js <outDir>` fetches dex sets + usage, validates, writes `src/js/data/sets/gen{1-9},champions.js` (minified on copy); UI loads those as `SETDEX_*` globals.

## Patterns & rationale
- Pattern: Facade over Adapter (`index` over `adaptable` + `data/interface`) — why: browser bundle convenience without locking library to shipped data; trade-off: dual `Pokemon/Move` class pairs and global-loader comment block must be kept in sync (`calc/src/index.ts:1-58,76-121`).
- Pattern: Generation Strategy (one engine per era behind `MECHANICS[gen]` table) — why: damage rules diverge irreconcilably (RBY stat-exp/DV vs ADV nature vs DPP split vs BW+ modifiers vs SV Tera); trade-off: large duplicated pipelines (~1k+ lines each) but isolates era quirks and keeps goldens readable (`calc/src/calc.ts:14-25`, `calc/src/mechanics/champions.ts:40-81` vs `calc/src/mechanics/gen789.ts:52-120`).
- Pattern: Fixed-point modifier chains (`chainMods(mods, lower, upper)` with 4096 denominator, `pokeRound`, `OF16/OF32`) — why: faithfully emulates cartridge rounding/overflow truncation; trade-off: magic constants everywhere, order-sensitive push sites (`calc/src/mechanics/util.ts:471-489,704-717`).
- Pattern: Mutable scratch copies + `RawDesc` accumulator — why: mirrors game step order (pre-checks mutate weather/types/boosts/items) while preserving a human-readable audit trail for `display()`; trade-off: engines rely on `calc.ts` cloning; calling engine directly without clone leaks mutations (`calc/src/calc.ts:34-40`, `calc/src/desc.ts:971-1106`).
- Pattern: Table-driven deltas (`RBY` literal + `GSC_PATCH...` + `extend`) — why: compact representation of 9 gens + Champions without a database; trade-off: huge checked-in literals, `toID` normalization edge cases (`flabebe`, `???`), deep-merge semantics to learn (`calc/src/data/moves.ts:46-341`, `calc/src/util.ts:106-157`).
- Pattern: Lazy presentation (`Result` holds raw rolls; `fullDesc/moveDesc/recovery/recoil/kochance` compute on demand; `combine/reduce` approximates >3-hit distributions) — why: keeps hot calc path numeric while supporting %/48th notations, hazards/EoT-aware KO chance; trade-off: `desc.ts` (~1144 lines) is the second-largest logic surface after engines (`calc/src/result.ts:36-118`, `calc/src/desc.ts:456-509,770-861`).
- Non-pattern (deliberate): no state management / no bundler-required UI — why: static-hostable on GitHub Pages, view-source friendly, per-page scripts (`index/randoms/honkalculate/oms/champions`) share one `shared_controls.js`; trade-off: globals, jQuery string selectors, `gen-specific gN` CSS gating, duplicated per-mode control files (`src/index.html:56-82`, `src/js/shared_controls.js:94-146`).
- Pattern: Validation-gated codegen + differential testing — why: `import` drops invalid Showdown sets before they ship; `smoketest` + `calc.test.ts` goldens catch formula drift; trade-off: external network + `@pkmn/*` version coupling (`import/src/set-import.ts:305-343`, `smoketest/src/smoketest.ts:39-70,123-200`).

## Coupling & extension points
- `gen.num` is the dominant coupling axis: stat formulas, boost tables, speed caps/paralysis divisors, multihit combination (`squashMultihit`), data-table index, UI `LEGACY_STATS[gen]`/DV-vs-IV switching all key off it — add/adjust a generation by adding tables + engine entry + `MECHANICS` slot + UI gating (`calc/src/stats.ts:14-15,118-131`, `calc/src/desc.ts:863-969`, `src/js/shared_controls.js:38-70`).
- In-place mutation contract: engines mutate `Pokemon/Move/Field` (ability nulling for Mold Breaker/Neutralizing Gas, type rewrites for Weather Ball/Terrain Pulse/-ate, screen clearing, multihit boost stepping); always `clone()` before `calculate*` or reuse (`calc/src/mechanics/gen789.ts:185-238`, `calc/src/mechanics/util.ts:327-469`).
- Shared kernel fragility: `mechanics/util.ts`, `stats.ts`, `items.ts`, `desc.ts` are imported by every engine — change rounding (`pokeRound/chainMods`), `getModifiedStat` tables, or STAB logic with cross-gen golden runs (`calc/src/test/helper.ts:84-131`).
- Data extension: new ability/item/move/species = add literal entries per affected gen + branch in relevant `calculate*Mods` + `RawDesc` field if user-visible + set/UI data if selectable; forme logic lives in both `Pokemon.getForme` and UI `showFormes`/Tera handlers (`calc/src/pokemon.ts:191-217`, `src/js/shared_controls.js:870-880,910-959`).
- UI extension: new mode = new HTML page + `*_controls.js` reusing `shared_controls.js` builders; new field option = HTML control + `createField()` mapping + `State.Side/Field` + engine consumption + `RawDesc` text (`src/js/shared_controls.js:1267-1355`, `calc/src/state.ts:39-84`).
- Library reuse: prefer `adaptable.ts` + `data/interface.ts` with your own `Generation` for custom data; prefer `index.ts` for stock Showdown data; `toID` + `extend/assignWithout` are the only sanctioned data-plumbing utils (`calc/src/adaptable.ts:1-12`, `calc/src/util.ts:1-28`).
- Service reuse: `server.js` is example-only (reads `req.body` on GET, no validation); productionize by validating `State.*` payloads and pinning `calc/dist` version.

## Feature/System Inventory (no deep-dives — input for feature analysis)
- Core damage engine dispatcher — gen-indexed `calculate()` entry + input cloning + per-era delegation — `calc/src/calc.ts`
- RBY/RBY-GSC engine (gen 1-2 DV/StatEXP rules) — `calc/src/mechanics/gen12.ts`
- ADV engine (gen 3 abilities/natures/doubles spread) — `calc/src/mechanics/gen3.ts`
- DPP engine (gen 4 physical/special split) — `calc/src/mechanics/gen4.ts`
- BW/XY engine (gen 5-6 modern modifier chains) — `calc/src/mechanics/gen56.ts`
- SM/SS/SV engine (gen 7-9 Z/Max/Tera/ruin/QP/Stellar) — `calc/src/mechanics/gen789.ts`
- Champions engine (gen-0 simplified ruleset + custom abilities) — `calc/src/mechanics/champions.ts`
- Shared battle math kernel (stat mods, speed, effectiveness, STAB, rounding/overflow) — `calc/src/mechanics/util.ts`
- Stat calculator (ADV/RBY/Champions formulas, Hidden Power, DV/IV) — `calc/src/stats.ts`
- Item/move-property helpers (plates/incense, berries, Fling, Natural Gift, Drives) — `calc/src/items.ts`
- Domain models (Pokemon/Move/Field/Side construction, forme, Z/Max, clone) — `calc/src/pokemon.ts`, `calc/src/move.ts`, `calc/src/field.ts`
- Plain input DTOs for all entities/field sides — `calc/src/state.ts`
- Pluggable data contracts (Generation/Abilities/Items/Moves/Species/Types/Natures) — `calc/src/data/interface.ts`
- Bundled generation data facade (per-gen collections constructor) — `calc/src/data/index.ts`
- Species table across 9 gens + Champions (types/base stats/weights/abilities/formes) — `calc/src/data/species.ts`
- Move table with per-gen patches (BP/type/category/flags/multihit/Z/Max metadata) — `calc/src/data/moves.ts`
- Ability allow-list per generation — `calc/src/data/abilities.ts`
- Item data + Mega-Stone/Natural Gift metadata — `calc/src/data/items.ts`
- Type chart per generation + effectiveness lookup — `calc/src/data/types.ts`
- Nature plus/minus table — `calc/src/data/natures.ts`
- Result container (damage rolls, range, desc/recovery/recoil/KO accessors) — `calc/src/result.ts`
- Description/KO engine (text builder, hazards/EoT, KO probability, multihit squash) — `calc/src/desc.ts`
- Stock-library facade (gen-number shorthand classes/functions) — `calc/src/index.ts`
- Data-agnostic entry (custom Generation injection) — `calc/src/adaptable.ts`
- ID normalization + deep-merge + error helpers — `calc/src/util.ts`
- Main 1v1 calculator UI (gen/mode/result/move/panel controls) — `src/index.html`, `src/js/shared_controls.js`
- Random Battles mode UI (randdex pools, role/move/tera selection) — `src/randoms.html`, `src/js/index_randoms_controls.js`
- Honkalculate 1v1 variant UI — `src/honkalculate.html`, `src/js/honkalculate_controls.js`
- Other Metagames (OMs) UI — `src/oms.html`, `src/js/oms_controls.js`
- Champions mode UI — `src/champions.html`
- Shared HTML wrapper template for all pages — `src/shared.template.html`
- Moveset import/export widget (localStorage + text parsing) — `src/js/moveset_import.js`
- Dark-theme toggle — `src/js/dark-theme-toggle.js`
- Curated competitive sets per gen + Champions — `src/js/data/sets/gen1.js`–`gen9.js`, `champions.js`
- Tier/format lists per gen — `src/js/data/tiers/gen1.json`–`gen9.json`
- Static-site builder (compile, copy, template hash-busting) — `build`
- Browser bundle helper (Babel + Terser) — `bundler.js`
- Optional Express API + static server — `server.js`
- Set/usage importer from Showdown data (dex sets + stats -> SETDEX) — `import/src/set-import.ts`
- Sim differential smoke-tester (random battles vs calc rolls) — `smoketest/src/smoketest.ts`
- Jest golden/unit suites (calc ranges/descs, data, pokemon, move, stats, utils) — `calc/src/test/calc.test.ts`, `helper.ts`, `data.test.ts`, `pokemon.test.ts`, `move.test.ts`, `stats.test.ts`, `utils.test.ts`
- CI build/test + Pages publish and scheduled data update — `.github/workflows/test.yml`, `update.yml`
- Container build (node build + server) — `Dockerfile`, `.dockerignore`
