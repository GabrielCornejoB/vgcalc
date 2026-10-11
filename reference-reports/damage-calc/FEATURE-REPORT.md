> **Reference note (moved 2026-10-07):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/damage-calc`. Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Feature Report — damage-calc

> Sources: ARCHITECTURE-REPORT.md (inventory), EXPLORER-REPORT.md, direct reads. ARCHITECTURE-REPORT.md was present and used as the feature-discovery baseline; all 40+ inventory items below were verified by direct code reads. No fallback inventory was needed.

## Feature inventory
| Feature | One-line purpose | Key files |
|---|---|---|
| Core damage engine dispatcher | Gen-indexed `calculate()` + input cloning + per-era delegation | `calc/src/calc.ts` |
| RBY/GSC engine (gen 1–2) | DV/StatEXP, 217–255 rolls, 255-cap overflow rules | `calc/src/mechanics/gen12.ts` |
| ADV engine (gen 3) | Abilities/natures, doubles spread, modern base formula start | `calc/src/mechanics/gen3.ts` |
| DPP engine (gen 4) | Physical/Special split refinement of ADV pipeline | `calc/src/mechanics/gen4.ts` |
| BW/XY engine (gen 5–6) | Modern 4096-modifier chains, weather/gem/berry logic | `calc/src/mechanics/gen56.ts` |
| SM/SS/SV engine (gen 7–9) | Z/Max/Tera/ruin/QP/Stellar/Parental-Bond multihit matrix | `calc/src/mechanics/gen789.ts` |
| Champions engine (gen 0) | Simplified SP-based ruleset + custom abilities (Mega Sol, Dragonize, Aura Guard…) | `calc/src/mechanics/champions.ts` |
| Shared battle math kernel | Stat mods, speed, effectiveness, STAB, rounding/overflow | `calc/src/mechanics/util.ts` |
| Stat calculator | ADV/RBY/Champions formulas, Hidden Power, DV/IV/StatEXP conversion | `calc/src/stats.ts` |
| Item/move-property helpers | Plates/incense boost type, berry resist, Fling, Natural Gift, Drives, Memories | `calc/src/items.ts` |
| Domain models (Pokemon/Move/Field/Side) | Mutable working state: construction, forme, Z/Max, clone | `calc/src/pokemon.ts`, `calc/src/move.ts`, `calc/src/field.ts` |
| Plain input DTOs | `State.*` interfaces for serializable inputs | `calc/src/state.ts` |
| Pluggable data contracts | `Generation/Abilities/Items/Moves/Species/Types/Natures` interfaces | `calc/src/data/interface.ts` |
| Bundled generation data facade | `Generations.get(n)` per-gen collections constructor | `calc/src/data/index.ts` |
| Species table (9 gens + Champions) | Types/base stats/weights/abilities/formes, patch-layered | `calc/src/data/species.ts` |
| Move table with per-gen patches | BP/type/category/flags/multihit/Z/Max metadata, layered deltas | `calc/src/data/moves.ts` |
| Ability allow-list per gen | Existence table per generation | `calc/src/data/abilities.ts` |
| Item data + Mega/Natural Gift metadata | Mega-stone map, berries, Natural Gift payloads | `calc/src/data/items.ts` |
| Type chart per gen + effectiveness | Per-gen `effectiveness` maps | `calc/src/data/types.ts` |
| Nature plus/minus table | 25-nature plus/minus pairs | `calc/src/data/natures.ts` |
| Result container | Holds rolls + lazy `range/desc/recovery/recoil/kochance` | `calc/src/result.ts` |
| Description/KO engine | Text builder, hazards/EoT, KO probability, multihit squash | `calc/src/desc.ts` |
| Stock-library facade | Gen-number shorthand `Pokemon/Move/calculate/calcStat` | `calc/src/index.ts` |
| Data-agnostic entry | Custom `Generation` injection exports | `calc/src/adaptable.ts` |
| ID + deep-merge utils | `toID`, `extend`, `assignWithout`, `error` | `calc/src/util.ts` |
| Main 1v1 calculator UI | Gen/mode/result/move/panel controls, DOM→model→calc→render | `src/index.html`, `src/js/shared_controls.js` |
| Random Battles mode UI | Randdex pools, role/move/tera selection, best-move highlight | `src/randoms.html`, `src/js/index_randoms_controls.js` |
| Honkalculate 1v1 variant UI | One-vs-all / all-vs-one DataTable over tier-filtered sets | `src/honkalculate.html`, `src/js/honkalculate_controls.js` |
| Other Metagames (OMs) UI | Mix-and-Mega / Tier Shift / Scalemons stat rewriting | `src/oms.html`, `src/js/oms_controls.js` |
| Champions mode UI | Gen-0 SP-based page reusing shared controls | `src/champions.html` |
| Shared HTML wrapper template | `[content]` wrapper + cache-busted asset injection | `src/shared.template.html` |
| Moveset import/export widget | Showdown-text ↔ SETDEX ↔ localStorage custom sets | `src/js/moveset_import.js` |
| Dark-theme toggle | `prefers-color-scheme` + localStorage persisted theme | `src/js/dark-theme-toggle.js` |
| Curated competitive sets | `SETDEX_*` per gen + Champions globals | `src/js/data/sets/gen1.js`–`gen9.js`, `champions.js` |
| Tier/format lists | Per-gen tier JSONs for honkalculate/OMs | `src/js/data/tiers/gen1.json`–`gen9.json` |
| Static-site builder | Compile calc, copy src+calc/dist→dist, template hash-busting | `build` |
| Browser bundle helper | Babel + Terser bundling of compiled calc | `bundler.js` |
| Optional Express API + static server | `GET /calculate` + `express.static('dist')` | `server.js` |
| Set/usage importer | Showdown dex sets + usage stats → validated `SETDEX_*` | `import/src/set-import.ts` |
| Sim differential smoke-tester | Random `@pkmn/sim` battles vs `calc.calculate` rolls | `smoketest/src/smoketest.ts` |
| Jest golden/unit suites | Per-gen range/desc goldens + data/pokemon/move/stats/utils tests | `calc/src/test/calc.test.ts`, `helper.ts`, `data.test.ts`, `pokemon.test.ts`, `move.test.ts`, `stats.test.ts`, `utils.test.ts` |
| CI build/test + Pages + scheduled data update | Test+publish and every-3-days set refresh | `.github/workflows/test.yml`, `update.yml` |
| Container build | `node build` + `server.js` image | `Dockerfile`, `.dockerignore` |

## Feature details
### Core damage engine dispatcher
- What it does: Single pure entry `calculate(gen, attacker, defender, move, field?)` that clones all inputs and delegates to the per-era engine via a positional `MECHANICS[gen.num]` table (index 0 = Champions, 1–2 = RBY/GSC, 3 = ADV, 4 = DPP, 5–6 = BW/XY, 7–9 = SM/SS/SV).
- Where it lives (`path` refs): `calc/src/calc.ts:14-41`
- How it works: `MECHANICS` array at `calc/src/calc.ts:14-25`; `calculate()` at `calc/src/calc.ts:27-41` calls `attacker.clone(), defender.clone(), move.clone(), field.clone()|new Field()` because engines mutate boosts/stats/types/items in place. Callers never branch on gen.
- Inputs/outputs & data deps: Inputs are `Generation` + `Pokemon/Move/Field` domain objects; output is `Result` (rolls + RawDesc). Depends on all six engine modules.
- Entry-points (UI/CLI/API): Library API (`@smogon/calc`); UI calls `calc.calculate(GENERATION, ...)` in `shared_controls.js` / mode controls; `server.js:20-26` calls it from `GET /calculate`; smoketest calls it per sim event (`smoketest/src/smoketest.ts:172-181`).

### RBY/GSC engine (gen 1–2)
- What it does: Faithful gen 1–2 damage: StatEXP/DV stats (via `computeFinalStats`), type-precedence swap in GSC, crit-ignores-mods with doubled level (RBY) / doubled damage (GSC), Reflect/LightScreen ×2, Light Ball/Thick Club/Metal Powder, 255-overflow truncation, 217–255 random rolls, Explosion halving, variable-BP Flail/Reversal/Present.
- Where it lives (`path` refs): `calc/src/mechanics/gen12.ts:10-305`
- How it works: Preamble `computeFinalStats(...'atk','def','spa','spd','spe')` (`gen12.ts:17`); early exits for Status/Protect/Pain Split/immunity-0 (`gen12.ts:27-42,91-93`); fixed-damage handling differs by gen (`gen12.ts:44-50,95-101`); GSC type-order swap (`gen12.ts:75-82`); crit/mod logic (`gen12.ts:131-164`); 255-cap `at/df` truncation (`gen12.ts:172-175`); Present glitch lookup (`gen12.ts:179-189`); base `⌊⌊⌊(2·lv/5+2)·at·BP/df⌋/50⌋`, min 997+2 (`gen12.ts:196-223`); weather ×1.5/÷2, STAB ×1.5, per-type floor in RBY (`gen12.ts:225-246`); 39-roll loop 217–255 (`gen12.ts:254-266`); Triple Kick re-entry + multihit matrix (`gen12.ts:268-302`).
- Inputs/outputs & data deps: Same `calculate*` signature; uses `stats.ts` (via Pokemon), `items.ts:getItemBoostType` (`gen12.ts:2,212-221`), `mechanics/util.ts:computeFinalStats/getMoveEffectiveness/handleFixedDamageMoves`.
- Entry-points (UI/CLI/API): Reached via dispatcher for `gen.num ∈ {1,2}`; goldens in `calc/src/test/calc.test.ts`.

### ADV engine (gen 3)
- What it does: Gen 3 rules: natures, abilities, EVs/IVs, doubles spread reduction, choice/items/weather/ability BP-attack-defense-final stage mods of that era.
- Where it lives (`path` refs): `calc/src/mechanics/gen3.ts:20-~700` (exports `calculateADV`; dispatcher slot `calc/src/calc.ts:18`)
- How it works: Same staged shape as later engines (`calculateBasePowerADV/BPMods/Attack/Defense/FinalMods` family in `gen3.ts`) but without physical/special split (type-based category), without Z/Max/Tera/QP/ruin/Stellar, with gen-3-specific ability/item/weather lists. Verified by export presence; detailed formulas mirror DPP/BW structure with era-gated branches.
- Inputs/outputs & data deps: Same signature; `mechanics/util.ts` kernel + `data/*` gen-3 tables.
- Entry-points (UI/CLI/API): Dispatcher for `gen.num === 3`; UI gen selector; test goldens.

### DPP engine (gen 4)
- What it does: Gen 4 rules: physical/special split, Platinum-era items/abilities, Iron Ball Klutz exception, gen-4 crash-damage text.
- Where it lives (`path` refs): `calc/src/mechanics/gen4.ts:24-~700` (`calculateDPP`, `calculateBasePowerDPP`, `calculateBPModsDPP`, `calculateAttackDPP`, `calculateDefenseDPP`)
- How it works: Staged pipeline `calculateDPP → BasePower → Attack → Defense → BaseDamage → FinalMods → 16-roll loop` (same shape as `gen789.ts:576-685` but era-scoped). Notable gen-4 hooks: `checkItem` early-returns for Iron Ball in gen 4 (`mechanics/util.ts:204`), struggle-recoil rounding note (`desc.ts:259-260`), gen-4 crash-damage branch (`desc.ts:239-243`).
- Inputs/outputs & data deps: Same signature; kernel + gen-4 data tables.
- Entry-points (UI/CLI/API): Dispatcher for `gen.num === 4`; UI; goldens.

### BW/XY engine (gen 5–6)
- What it does: Gen 5–6 modern engine: 4096-denominator `chainMods` everywhere, gems/berries/Aura/terrain-seed era behavior, crit 1.5×-era, spread reduction, Parental Bond child re-entry.
- Where it lives (`path` refs): `calc/src/mechanics/gen56.ts:41-~1100` (`calculateBWXY`, `calculateBasePowerBWXY:451`, `calculateBPModsBWXY:613`, `calculateAttackBWXY:810`, `calculateAtModsBWXY:851`, `calculateDefenseBWXY:933`, `calculateDfModsBWXY:982`)
- How it works: Same Initiate→BP→At→Df→Base→Final→rolls→multihit-matrix shape as gen789 but with gen 5/6 ability/item/move lists (e.g. terrain multiplier `gen.num > 7 ? 5325 : 6144` shows the era split at `gen789.ts:1150`). Dispatcher maps both 5 and 6 to `calculateBWXY` (`calc/src/calc.ts:19-21`).
- Inputs/outputs & data deps: Same signature; kernel + gen 5/6 tables.
- Entry-points (UI/CLI/API): Dispatcher for `gen.num ∈ {5,6}`; UI; goldens.

### SM/SS/SV engine (gen 7–9)
- What it does: Flagship modern engine: Z/Max resolution, Tera (+Stellar first-use floor 60), ruin abilities, Quark Drive/Protosynthesis, Embody Aspect, Supreme Overlord, -ate/Normalize/Dragonize, Weather/Terrain Ball/Pulse, Revelation Dance, Ivy Cudgel/Raging Bull/Aura Wheel, multihit + `timesUsed` matrices, Parental Bond child recursion, spread/parental/weather/crit base-damage order.
- Where it lives (`path` refs): `calc/src/mechanics/gen789.ts:52-783` (`calculateSMSSSV`), `785-1038` (`calculateBasePowerSMSSSV`), `1040-1315` (`calculateBPModsSMSSSV`), `1317-1365` (`calculateAttackSMSSSV`), `1367-~1523` (`calculateAtModsSMSSSV`), `1524-~1717` (`calculateDefense/DfMods`), `1718-~1900` (`calculateFinalModsSMSSSV`)
- How it works: Preamble mutates/caches: Air Lock/Teraform Zero/Forecast/Item/RawStatChanges/SeedBoost/Dauntless/Embody → `computeFinalStats(def,spd,spe)` → Intimidate/Download/Intrepid/WindRider → `computeFinalStats(atk,spa)` → Infiltrator (`gen789.ts:61-102`). Builds `RawDesc` (`gen789.ts:104-120`). Early exits: Status, Protect-vs-breaksProtect (`gen789.ts:144-152`), Pain Split (`155-159`), Mold Breaker/Neutralizing Gas nulling (`185-232`), crit (`236-238`), dynamic typing (`241-388`), effectiveness incl. Tera/Stellar (`390-470`), Wonder Guard/absorb immunities (`484-507`). Variable-BP switch (`799-999`) + Tera-Stellar 60-floor (`1025-1036`); BP/At/Df/Base/Final staged mods; `getStabMod/getStellarStabMod` (`635-636`); burn (`638-643`); `chainMods(finalMods,41,131072)` (`666`); spread + Parental Bond child (`668-678`); 16-roll `getFinalDamage` loop (`681-685`); `timesUsed/hits` matrix with `checkMultihitBoost` re-entry per hit/turn (`687-777`).
- Inputs/outputs & data deps: Same signature; heaviest consumer of `items.ts` (all helpers imported `gen789.ts:3-11`) and `util.ts` kernel (imported `gen789.ts:17-50`).
- Entry-points (UI/CLI/API): Dispatcher for `gen.num ∈ {7,8,9}`; main UI path; `GET /calculate`; smoketest; goldens.

### Champions engine (gen-0 simplified ruleset)
- What it does: Pokémon Champions mode: SP-point stats (no IV/EV), no Z/Max/Dynamax/Tera/ruin, trimmed ability/move lists, custom abilities (Mega Sol sun-override, Dragonize -ate, Aura Guard, Eelevate, Fire Mane), simplified speed/paralysis and burn/EoT divisors.
- Where it lives (`path` refs): `calc/src/mechanics/champions.ts:40-520` (`calculateChampions`), `522-654` (`calculateBasePowerChampions`), `656-838` (`calculateBPModsChampions`), `840-956` (`calculateAttack/AtModsChampions`), `958-1043` (`calculateDefense/DfModsChampions`), `1045-1089` (`calculateBaseDamageChampions`), `1091-1190` (`calculateFinalModsChampions`)
- How it works: Parallel-but-slimmer copy of SMSSSV: no TeraformZero/Download/Intrepid/SeedBoost import set (`champions.ts:14-38` vs `gen789.ts:17-50`); no breaksProtect-via-Z/Dynamax (`champions.ts:95-96`); `Mega Sol` replaces sun checks in Weather Ball/Solar Beam/sand-SpD (`champions.ts:143-151,708-711,992-995,1069-1081`); -ate set is Aerilate/Dragonize/Pixilate/Refrigerate only (`champions.ts:213-230`); simplified BP cases (no Pursuit-switch, no Collision Course, terrain mult fixed 5325 at `721`); simplified At mods (no Slow Start/Defeatist/Gorilla/Hustle nuances at `891-956`); Df mods only Marvel Scale/Grass Pelt/Fur Coat (`1016-1043`).
- Inputs/outputs & data deps: Same signature with `gen.num === 0`; uses `stats.ts:calcStatChampions` via `Pokemon` constructor and `util.ts` subset.
- Entry-points (UI/CLI/API): Dispatcher slot 0 (`calc/src/calc.ts:15`); `src/champions.html` page; `SETDEX_CHAMPIONS`; `sps` inputs in UI.

### Shared battle math kernel
- What it does: Cross-engine primitives: boost tables, final-stat/speed computation, pre-turn ability/field checks, multihit boost stepping, 4096-chain rounding, base/final damage, STAB (incl. Stellar), Shell Side Arm category, weight, stat text, fixed damage, pokeRound/OF16/OF32.
- Where it lives (`path` refs): `calc/src/mechanics/util.ts:30-35` (`isGrounded`), `37-70` (`getModifiedStat` with gen1/2 table vs modern `modernGenBoostTable`), `72-90` (`computeFinalStats`), `92-137` (`getFinalSpeed`: Tailwind/Unburden/Chlorophyll/Sand Rush/Swift Swim/Slush Rush/Surge Surfer/Quick Feet/Slow Start/QP/Choice Scarf/Iron Ball/Quick Powder + paralyis ÷2 gen≥7/0 else ÷4), `139-166` (`getMoveEffectiveness` incl. Freeze-Dry/Flying Press/Ring Target), `168-325` (`checkAirLock/TeraformZero/Forecast/Item/RawStatChanges/Intimidate/Download/IntrepidSword/DauntlessShield/WindRider/Embody/Infiltrator/SeedBoost`), `328-469` (`checkMultihitBoost`: Gooey/Power-Up Punch/berries/Stamina/Weak Armor/dropsStats/Mummy), `471-479` (`chainMods`), `481-489` (`getBaseDamage`), `501-539` (`getQPBoostedStat/isQPActive`), `541-559` (`getFinalDamage`: 85–100 roll → STAB → effectiveness → burn → protect → finalMod, OF16), `576-588` (`getShellSideArmCategory`), `590-607` (`getWeight`), `609-641` (`getStabMod/getStellarStabMod`), `643-691` (`countBoosts/getStatDescriptionText/handleFixedDamageMoves/pokeRound/OF16/OF32`)
- How it works: Pure helpers mutating passed-in `Pokemon/Field/RawDesc`; fixed-point convention is 4096-denominator with `pokeRound` (Round DOWN on .5) and OF16/OF32 overflow emulation.
- Inputs/outputs & data deps: Operates on domain objects + `Generation`; no I/O.
- Entry-points (UI/CLI/API): Imported by every engine (`gen789.ts:17-50`, `champions.ts:14-38`, `gen12.ts:8`).

### Stat calculator
- What it does: Converts base+IV/EV/DV/StatEXP/SP+level+nature → final stats; Hidden Power type/power both directions; DV↔IV and StatEXP↔EV conversion.
- Where it lives (`path` refs): `calc/src/stats.ts:14-15` (`STATS` per-gen table), `84-116` (`getHPDV/IVToDV/DVToIV/EVToStatEXP/StatEXPToEv/DVsToIVs`), `118-131` (`calcStat` dispatcher: 0→Champions, <3→RBY, else ADV), `133-159` (`calcStatChampions`: `base+sp+75` HP, `⌊n·(base+sp+20)⌋` others), `161-191` (`calcStatADV`), `193-205` (`calcStatRBY/RBYFromDV`), `207-251` (`getHiddenPowerIVs/getHiddenPower` with gen≤2 DV path vs gen≥3 IV-bit path, 60-BP cap gen≥6)
- How it works: `Pokemon` constructor calls `calcStat` per stat (`pokemon.ts:179-189`); UI live preview calls `calc.calcStat` per keystroke (`shared_controls.js:1389-1413`).
- Inputs/outputs & data deps: Inputs `(gen, stat, base, iv, ev, level, nature)`; depends on `Natures.get(toID(nature))`.
- Entry-points (UI/CLI/API): `calc.Stats` / `A.Stats` exports; `index.ts:123-141` (`calcStat` wrapper); UI `calcHP/calcStat`.

### Item/move-property helpers
- What it does: Maps items/berries/drives/memories to battle effects without touching data tables.
- Where it lives (`path` refs): `calc/src/items.ts:4-9` (`SEED_BOOSTED_STAT`), `11-76` (`getItemBoostType`: plates/fangs/incenses), `78-119` (`getBerryResistType`), `121-397` (`FLING_*` sets + `getFlingPower`), `399-402` (`getNaturalGift` from `gen.items`), `404-424` (`getTechnoBlast/getMultiAttack`)
- How it works: Pure string→type/power lookups; `getFlingPower(item, gen)` has gen-gated Big Nugget and Plate-fallback branches (`items.ts:378-397`).
- Inputs/outputs & data deps: Input item name (+gen); `getNaturalGift` reads `gen.items.get(toID(item)).naturalGift`.
- Entry-points (UI/CLI/API): Called from engines (notably `gen789.ts:252-266,901-904` for Fling/Techno/Multi-Attack/Natural Gift/Judgment).

### Domain models (Pokemon/Move/Field/Side)
- What it does: Mutable working-state classes the engines clone and mutate: stats/IV/EV/boosts/nature/ability/item/status/Tera/Dynamax/overrides (Pokemon); Z/Max resolution, multihit, `overrides`, typeless/crit/happiness (Move); weather/terrain/rooms/gravity/auras/ruins + both Sides (Field/Side).
- Where it lives (`path` refs): `calc/src/pokemon.ts:10-101` (fields+constructor: DV-derived HP IV gen1/2 at `77-86`, `rawStats/stats` init `88-94`, curHP clamp `96-97`), `103-119` (`maxHP/curHP` Dynamax scaling), `121-153` (`hasAbility/hasItem/hasStatus/hasType/hasOriginalType/named`), `155-177` (`clone`), `191-217` (`getForme`: Mega/Primal/Relic/Dragon-Ascent logic); `calc/src/move.ts:47-174` (constructor: Max rewrite `61-88`, Z rewrite `89-98`, multihit/Skill Link `99-115`, typeless Struggle/Future Sight/Doom Desire `127-130`, pre-split category `131-134`, `dropsStats` `136-139`, crit incl. gen-1 slash set `149-151`, Return/Frustration 102 `168-173`), `184-198` (`clone`), `201-295` (`getZMoveName/getMaxMoveName` tables); `calc/src/field.ts:4-71` (`Field`: ctor/clone/swap/hasWeather/hasTerrain), `74-132` (`Side`: spikes/SR/screens/seed/charge/tailwind/helping-hand/flower-gift/friend-guard/aurora/battery/power-spot/switching + clone)
- How it works: Constructors resolve `gen.species/moves.get(toID(name))` deep-merged with `options.overrides` via `extend(true,...)`; `Move` rewrites `data` to the resolved Z/Max move before copying fields.
- Inputs/outputs & data deps: Inputs `(gen, name, Partial<State.*>)`; depend on `Generations` data + `util.ts`.
- Entry-points (UI/CLI/API): Constructed by UI builders (`createPokemon/getMoveDetails/createField` in `shared_controls.js:1066-1355`), `server.js:21-25`, smoketest (`smoketest.ts:39-70,109-121`), tests via `helper.ts:24-38`.

### Plain input DTOs
- What it does: Serializable option shapes (`State.Pokemon/Move/Field/Side`) decoupled from class behavior — the contract for JSON APIs, custom sets, and tests.
- Where it lives (`path` refs): `calc/src/state.ts:4-25` (Pokemon: level/ability/abilityOn/Dynamax/alliesFainted/boostedStat/item/gender/nature/ivs/evs/boosts/curHP/status/teraType/toxicCounter/moves/overrides), `27-37` (Move: useZ/useMax/isCrit/Stellar/hits/timesUsed/metronome/overrides), `39-55` (Field), `57-83` (Side)
- How it works: `Partial<State.*>` accepted by every domain constructor; `server.js` passes `req.body.*Options` straight through as these DTOs.
- Inputs/outputs & data deps: Type-only; references `data/interface.ts` names.
- Entry-points (UI/CLI/API): `GET /calculate` body; `adaptable.ts` custom-data callers.

### Pluggable data contracts
- What it does: Defines the `Generation` seam (`Abilities/Items/Moves/Species/Types/Natures` collections + `Ability/Item/Move/Specie/Type/Nature` records) so hosts can inject non-Showdown data.
- Where it lives (`path` refs): `calc/src/data/interface.ts:1-38` (ID/GenerationNum/StatID/Types/Weather/Terrain/NatureName/TypeName/MoveCategory/Target), `40-52` (Generations/Generation), `54-177` (per-kind `get(id)` + iterator contracts and record fields incl. `Move.multihit/multiaccuracy/Z/Max/overrides`, `Item.megaStone/naturalGift`, `Specie.baseStats/weightkg/abilities/otherFormes`)
- How it works: Engines only touch `gen.*.get(toID(...))` and record fields, never concrete table classes — `adaptable.ts:1-12` passes any `interface.Generation` through.
- Inputs/outputs & data deps: None (types only).
- Entry-points (UI/CLI/API): Implement this interface to use `adaptable.ts` with custom data.

### Bundled generation data facade
- What it does: Builds a `Generation` value object holding the six per-gen collections for `gen.num`.
- Where it lives (`path` refs): `calc/src/data/index.ts:10-36` (`Generations.get(num) → new Generation(num)` wiring `Abilities/Items/Moves/Species/Types/Natures`)
- How it works: Constructor news up each collection with `num`; each collection internally indexes its prebuilt `*_BY_ID[gen]` map (e.g. `abilities.ts:570-609`, `types.ts:438-480` pattern).
- Inputs/outputs & data deps: Input `GenerationNum` (0–9); depends on all six table modules.
- Entry-points (UI/CLI/API): `calc.Generations.get(n)` used by UI global, `index.ts:68`, `server.js:12`, tests.

### Species table across 9 gens + Champions
- What it does: Every species' types/baseStats/weightkg/gender/abilities/Gmax/forme graph per generation, expressed as base literal + cumulative patches.
- Where it lives (`path` refs): `calc/src/data/species.ts:25-35,1562` (layering `extend(true,...)` pattern per arch report; `Species.get/iterator` class)
- How it works: `RBY` literal + `GSC_PATCH/ADV_PATCH/.../SV_PATCH + CHAMPIONS_PATCH` merged in order so later gens inherit; `toID` lookup; `Pokemon` merges `options.overrides` on top.
- Inputs/outputs & data deps: Static literals; no runtime I/O.
- Entry-points (UI/CLI/API): `gen.species.get(...)` in `pokemon.ts:53`, `getForme:197`, UI `pokedex` global.

### Move table with per-gen patches
- What it does: Every move's BP/type/category/flags/multihit/target/recoil/drain/Z/Max per generation.
- Where it lives (`path` refs): `calc/src/data/moves.ts:218-341` (GSC_PATCH→GSC), `343-571` (ADV_PATCH→ADV), `573-1177` (DPP_PATCH→DPP), `1179-1681` (BW), `1685-1978` (XY), `1980-2839` (SM), `2841-4139` (SS), `4160-4936` (SV incl. `ZA_PATCH:4923`), `5453-5498` (`CHAMPIONS_PATCH`, `MOVES = [CHAMPIONS,RBY,...,SV]`)
- How it works: Same `extend(true, {}, PREV, PATCH)` chain; `Move` constructor layers `options.overrides` over the resolved row, then rewrites for Z/Max.
- Inputs/outputs & data deps: Static literals.
- Entry-points (UI/CLI/API): `gen.moves.get(...)` in `move.ts:58,71,92`, `getMaxMoveName:268`; UI `moves` global.

### Ability allow-list per generation
- What it does: Which abilities exist in each gen (existence gate for legality/selection).
- Where it lives (`path` refs): `calc/src/data/abilities.ts:570-609` (`Abilities` class + `ABILITIES_BY_ID`-style per-gen maps per arch report)
- How it works: Per-gen string lists; `Abilities.get(num)` scoping; engines test `hasAbility(...)` against the mon's ability string (empty string = nulled by Mold Breaker/Neutralizing Gas).
- Inputs/outputs & data deps: Static lists.
- Entry-points (UI/CLI/API): `gen.abilities` (rarely read directly; UI ability dropdowns derive from species + this list).

### Item data + Mega-Stone/Natural Gift metadata
- What it does: Item existence + `megaStone` evolution map + `naturalGift` type/power + berry flags.
- Where it lives (`path` refs): `calc/src/data/items.ts` (exports `ITEMS, MEGA_STONES` via `index.ts:151`; `Item.megaStone/naturalGift/isBerry` fields `interface.ts:76-81`)
- How it works: `gen.items.get(toID(item))` lookups: Knock Off resistance via `item.megaStone[defender.name]` (`gen789.ts:1078-1082`, `champions.ts:677-681`); `getNaturalGift` reads `.naturalGift` (`items.ts:399-402`); OMs `generateMegaDelta` walks `ITEMS_BY_ID[gen][item].megaStone` (`oms_controls.js:46-87`).
- Inputs/outputs & data deps: Static literals.
- Entry-points (UI/CLI/API): `gen.items.get(...)`; UI item dropdowns.

### Type chart per generation + effectiveness lookup
- What it does: Per-gen effectiveness matrix (`0/0.5/1/2`) backing all effectiveness math.
- Where it lives (`path` refs): `calc/src/data/types.ts:438-480` (`Types` class + `TYPE_CHART` export via `index.ts:155`)
- How it works: `gen.types.get(toID(move.type)).effectiveness[defType]` read in `getMoveEffectiveness` (`util.ts:156-164`) and hazard calc (`desc.ts:523-541`); Tera defender swaps the lookup target (`gen789.ts:416-425`); gen-1 floors per type separately (`gen12.ts:241-246`).
- Inputs/outputs & data deps: Static literals.
- Entry-points (UI/CLI/API): `gen.types.get(...)`; UI type selectors.

### Nature plus/minus table
- What it does: 25 natures → (plus, minus) stat pair; neutral natures have plus===minus.
- Where it lives (`path` refs): `calc/src/data/natures.ts` (exports `NATURES`; consumed `stats.ts:143-148,176-179`, `util.ts:681-689`)
- How it works: `natures.get(toID(nature))` → `[plus, minus]` → 1.1/0.9/1.0 multiplier in `calcStatADV/Champions` and desc text (`getStatDescriptionText` appends `+`/`-`).
- Inputs/outputs & data deps: Static table.
- Entry-points (UI/CLI/API): `gen.natures.get(...)`; UI nature dropdown.

### Result container
- What it does: Owns raw rolls (`number | number[] | [n,n] | number[][]`) plus back-refs (gen/attacker/defender/move/field/desc) and lazy presentation accessors.
- Where it lives (`path` refs): `calc/src/result.ts:7-82` (`Damage` union, `Result` ctor, `desc()`, `range()`, `fullDesc/moveDesc/recovery/recoil/kochance`), `84-118` (`damageRange` sums multihit pairs; `multiDamageRange` handles fixed/standard/multihit)
- How it works: Engines `new Result(...)` with damage 0 then fill `result.damage` (single `number[]` of 16, `[parent,child]` for Parental Bond, `number[][]` matrix for multihit/timesUsed, scalar for fixed/Pain Split). Accessors delegate to `desc.ts` (`result.ts:1,45-81`).
- Inputs/outputs & data deps: Depends on `desc.ts` functions.
- Entry-points (UI/CLI/API): Returned by `calculate()`; UI calls `.range()/.fullDesc()/.moveDesc()/.kochance()`; `server.js:20-27` serializes it as JSON.

### Description/KO engine
- What it does: Turns rolls + RawDesc + field state into human text (`Lv/EVs/item/ability/weather/terrain/screens/crit`), %/48th damage strings, Shell Bell/drain/Pain Split recovery, recoil/crash/Struggle/Mind Blown text, hazard + end-of-turn aware KO chance, multihit distribution squash/combine.
- Where it lives (`path` refs): `calc/src/desc.ts:10-57` (`RawDesc`), `59-81` (`display`: `min-max (%-%) -- KO text`), `83-101` (`displayMove`), `103-176` (`getRecovery`: Shell Bell/G-Max Finale/Pain Split/drain+Big Root), `179-267` (`getRecoil`: recoil%/crash-by-gen/Struggle/Mind Blown), `269-454` (`getKOChance`: OHKO fast path, hazard+EoT layering, 2–9-turn search, timesUsed branch), `456-509` (`combine`: fixed/16-roll/parental/multihit `combineTwo` + `reduce` approximation >3 hits), `516-569` (`getHazards`: SR/Steelsurge/Spikes, Boots/Magic Guard/Mountaineer gates), `571-768` (`getEndOfTurn`: weather abilities, Leftovers/Black Sludge/Sticky Barb, Leech Seed, status incl. toxic-counter, Bad Dreams, trapping, Salt Cure, pledges, G-Max persistent effects, Grassy Terrain, Heal Block/Psychic Noise), `770-861` (`computeKOChance/predictTotal`), `863-969` (`squashMultihit`: 16-roll and 39-roll tables for 2/3/4/5/10 hits + Parental Bond averaging), `971-1106` (`buildDescription`), `1108-1144` (`getDescriptionLevels/serializeText/toDisplay`)
- How it works: Pure functions over `(gen, attacker, defender, move, field, damage)`; `toDisplay('%')` = `⌊a·1000/b⌋/10`, `'px'` = 48ths. KO chance enumerates roll distributions with EoT/toxic adjustments rather than simulating.
- Inputs/outputs & data deps: Depends on `isGrounded` (`desc.ts:8`) and `result.ts:damageRange/multiDamageRange` (`desc.ts:5`).
- Entry-points (UI/CLI/API): Via `Result` accessors; UI `#mainResult` and per-move labels.

### Stock-library facade
- What it does: Browser/bundler-friendly shorthand: numeric-gen `calculate/Pokemon/Move/calcStat` plus re-exports of data/Stats/utils, with a global-loader ordering comment and an `Acalculate` alias kludge for `<script>`-tag use.
- Where it lives (`path` refs): `calc/src/index.ts:1-58` (loader comment + globals note), `60-74` (`calculate(gen:number|Generation,...)` → `Generations.get(n)`), `76-121` (`Move/Pokemon` subclasses accepting numeric gen + `getForme`), `123-141` (`calcStat`), `143-156` (re-exports: `Field/Side/Result/Generations/toID/State/ABILITIES/ITEMS/MEGA_STONES/MOVES/SPECIES/NATURES/TYPE_CHART/STATS/Stats`)
- How it works: Wraps `adaptable.ts` exports; `const Acalculate = exports.calculate` (`index.ts:58`) avoids self-recursion when files are concatenated as globals.
- Inputs/outputs & data deps: Depends on `data/index.ts:Generations` + `adaptable.ts`.
- Entry-points (UI/CLI/API): The documented `@smogon/calc` entry (`calc/src/index.ts` per `EXPLORER-REPORT.md:15`); UI global `calc`.

### Data-agnostic entry
- What it does: Same `calculate/Pokemon/Move/Field/Side/Result/Stats` surface but requiring an explicit `Generation` object — for hosts with custom data layers.
- Where it lives (`path` refs): `calc/src/adaptable.ts:1-12`
- How it works: Re-exports `calc/pokemon/move/field/result/stats` directly with no `Generations` import; comment (`adaptable.ts:1-5`) explains the two-import pattern.
- Inputs/outputs & data deps: Caller-supplied `interface.Generation`.
- Entry-points (UI/CLI/API): `import ... from '@smogon/calc/adaptable'` alternative entry.

### ID normalization + deep-merge + error helpers
- What it does: `toID` (lowercase-strip + `flabébé`→`flabebe` special-case), jQuery-style `extend(deep,...)` for patch layering, `assignWithout` for spc-exclusion, `error(err,msg)` throw-or-log.
- Where it lives (`path` refs): `calc/src/util.ts:4-10` (`toID`), `12-18` (`error`), `20-28` (`assignWithout`), `106-157` (`extend` deep-merge)
- How it works: Every data lookup normalizes via `toID`; every table layer and `overrides` merge uses `extend(true, ...)`; gen-1/2 `spc` IV/EV aliasing uses `assignWithout(cur, current, SPC)` (`pokemon.ts:227`).
- Inputs/outputs & data deps: None.
- Entry-points (UI/CLI/API): Re-exported as `calc.toID` (`index.ts:147`); used by UI (`oms_controls.js` `toID(item)`).

### Main 1v1 calculator UI
- What it does: Two-pane 1v1 page: gen/format pickers, per-side set/ability/item/level/nature/IV/EV/DV/boost/status/Tera/Dynamax/forme controls, 4 moves each with BP/type/cat/crit/Z/hits/turns/metronome, full field panel (weather/terrain/rooms/gravity/ruins/hazards/screens/seed/charge/tailwind/helping-hand/etc.), live stat/HP preview, `#mainResult` + per-move `%` lines + damage-value details, autosync (weather/terrain/QP ability toggles, forme/Tera/Stellar, Hidden Power IV fixup, gender/Rivalry).
- Where it lives (`path` refs): `src/index.html:51-82,147-763` (panels + per-page script includes); `src/js/shared_controls.js:38-70` (`LEGACY_STATS`/bounds/validation), `94-146` (format/defaultLevel handlers), `118-241` (live `calcHP/calcStats` + HP bar), `256-447` (ability/item/weather/terrain/QP autosync), `473-561` (move-selector: BP/type/cat/hits/Z/crit autosync + Hidden Power IV rewrite), `563-604` (item: Metronome/Orb-status sync), `612-835` (set-selector: base/type/EV/IV/ability/item/moves/forme/level-by-format), `870-1014` (formes/Tera/Ogerpon/Terapagos), `1016-1064` (Hidden Power correction), `1066-1265` (`createPokemon/getMoveDetails`: DOM→`calc.Pokemon/Move` incl. Dynamax/overrideMove/teraType/boosts/curHP/status/toxicCounter + BP/type/category overrides), `1267-1355` (`createField`: DOM→`calc.Field/Side`), `1357-1413` (`calcHP/totalEVs/calcStat` preview), `1415-1421` (`GENERATION` alias map)
- How it works: jQuery `calc-trigger` handlers scrape DOM into models, call `calc.calculate`, write text results; no framework/store/router — globals `calc/pokedex/moves/GENERATION`.
- Inputs/outputs & data deps: Reads `SETDEX_*/randdex/pokedex/moves`; writes result text.
- Entry-points (UI/CLI/API): `src/index.html` (default page); built to `dist/index.html` by `build:72`.

### Random Battles mode UI
- What it does: Same 1v1 shell but backed by randdex pools: shows ability/item/move/role/tera-type pools, filters to damaging moves, picks up-to-4 representative moves, computes all 8 moves both directions, highlights the fastest side's best move (or locked move), click-to-copy main result, `?import=` base64 team preload.
- Where it lives (`path` refs): `src/randoms.html`; `src/js/index_randoms_controls.js:1-9` (calc handler + autoset wiring), `11-21` (result locations), `23-95` (`performCalculations`: `createPokemon` both sides, `field.clone().swap()`, `calculateAllMoves`, `maxDamages` sort, fastest-side best-move select), `97-123` (result-move detail render), `125-160` (damage-hits display), `173-196` (`calculateAllMoves`: 8× `calc.calculate`), `198-260` (mode-switch links, notation switch, import param, click-to-copy)
- How it works: Pools come from the set-selector handler's `randset` branch (`shared_controls.js:646-697,726-757`); randoms `createPokemon(string)` overload builds from pooled sets (`shared_controls.js:1067-1129`).
- Inputs/outputs & data deps: `randdex` (GEN*RANDOMBATTLE) + shared builders.
- Entry-points (UI/CLI/API): `src/randoms.html` → `dist/randoms.html` (`build:73`); `?mode=randoms` routing.

### Honkalculate 1v1 variant UI
- What it does: One-vs-all / all-vs-one batch mode: one custom attacker/defender vs every set in the selected tier(s); DataTable with best-move %/px + KO text + type/ability/item columns, sortable/filterable, tier-gated level/format auto-adjust (LC 5, VGC 50, Doubles forcing).
- Where it lives (`path` refs): `src/honkalculate.html`; `src/js/honkalculate_controls.js:65-128` (`performCalculations`: tier filter → `createField` → `createPokemon` per set → `calculateMovesOfAttacker` → best-of-4 rollup), `137-143` (`calculateMovesOfAttacker`: 4× `calc.calculate`), `130-180` (tier/level/format sync), `182-238` (DataTable construction + Honkalculate button), `240-342` (mode links, tier UX, init)
- How it works: `getSetOptions()` enumerates setdex entries for checked tiers; `field.swap()` flips perspective for all-vs-one (`honkalculate_controls.js:82`).
- Inputs/outputs & data deps: `SETDEX_*` + `tiers/gen*.json` selection + shared builders.
- Entry-points (UI/CLI/API): `src/honkalculate.html` → `dist/honkalculate.html` (`build:71`); `?mode=one-vs-all|all-vs-one`.

### Other Metagames (OMs) UI
- What it does: Adds three OM toggles on top of the 1v1 shell: Mix-and-Mega (item→mega-stat/ability/type delta from `megaStone` data), Tier Shift (tier→flat BST addon via `tiers/*.json`), Scalemons (BST-proportional stat scaling); recomputes displayed base stats live and re-runs calcs.
- Where it lives (`path` refs): `src/oms.html`; `src/js/oms_controls.js:7-44` (`CONVERSION_G9/PREG9`, `specialCases`, `multMegas`), `46-101` (`generateMegaDelta` from `ITEMS_BY_ID[*].megaStone` + `calcMult` BST scaler), `103-225` (`autoUpdateStats`: MnM→Scale→TierShift priority + type/ability/weight patching), `227-337` (`performCalculationsOM`: alias name for display, per-item desc injection), `339-374` (item/gen triggers + `performCalculationsOM` rebinding)
- How it works: Rewrites the `.base` inputs (not the library) so the normal `createPokemon` path picks up adjusted stats; MnM name alias preserves display name while calc uses the big forme.
- Inputs/outputs & data deps: `ITEMS_BY_ID`, `pokedex`, `tiers/gen*.json` fetched synchronously on gen change (`oms_controls.js:353-367`).
- Entry-points (UI/CLI/API): `src/oms.html` → `dist/oms.html` (`build:74`).

### Champions mode UI
- What it does: Gen-0 page: same shell with `sps` (SP points 0–32) instead of EVs, level fixed 50 in preview, Champions setdex, Stellar/Tera-less ability/item lists.
- Where it lives (`path` refs): `src/champions.html`; `src/js/shared_controls.js:705-722` (`sps` fill from set `evs`), `1082-1093` (`sps` read in `createPokemon`), `1376-1386` (`totalEVs` uses `.sps` when `gen===0`), `1395-1399` (`calcStat` uses `.sps`, level 50 when `gen===0`), plus `checkExceptionsExport`/`addToDex` Champions branches (`moveset_import.js:243-244,296-297`)
- How it works: `gen===0` switches every `LEGACY_STATS[gen]`/stat path to the Champions SP formula (`stats.ts:128`); UI `bounds.sps [0,32]` (`shared_controls.js:79`).
- Inputs/outputs & data deps: `SETDEX_CHAMPIONS` (`src/js/data/sets/champions.js`).
- Entry-points (UI/CLI/API): `src/champions.html` → `dist/champions.html` (`build:70`).

### Shared HTML wrapper template
- What it does: Single chrome (header/footer/theme toggle/nav) wrapped around each page's `<!-- START WRAPPER -->…<!-- END WRAPPER -->` content at build time, with MD5 cache-busters on asset URLs.
- Where it lives (`path` refs): `src/shared.template.html`; `build:11-34` (`makeCachebuster`: extract wrapper, `template.replace('[content]', ...)`, regex `(src|href)` → `?md5[:8]`, write `dist/*.html`)
- How it works: `build:70-74` applies it to all five pages (champions/honkalculate/index/randoms/oms).
- Inputs/outputs & data deps: Reads `src/*.html` + `dist/*` asset bytes for hashing.
- Entry-points (UI/CLI/API): Build-time only.

### Moveset import/export widget
- What it does: Exports either side to Showdown text (name/gender/item/ability/level/Tera/Gmax/EVs/nature/sub-31 IVs/moves) and imports pasted text into `localStorage.customsets`, merged into every `SETDEX_*` + randoms pools; supports forme-exception normalization both directions.
- Where it lives (`path` refs): `src/js/moveset_import.js:1-10` (`placeBsBtn` Import button), `12-77` (`ExportPokemon` + `#exportL/#exportR`), `111-210` (`findSpecies/getGender/getItem/getStats/getMoves` parsers), `231-319` (`addToDex/updateDex`: fan-out to all `SETDEX_*`, `localStorage.customsets`), `321-354` (`addSets`: split rows → species → stats → moves → `addToDex`), `356-569` (`checkExceptionsImport/Export`: Alcremie/Aegislash/Flabébé/Vivillon/Unown/etc.), `571-600` (clear-sets + boot hydration)
- How it works: Lookup validation against `calc.SPECIES[9]/ITEMS[9]/ABILITIES[9]/TYPE_CHART[9]` so typos are dropped; Aegislash-Blade fans out to Shield/Both; `isCustomSet` flag marks imports.
- Inputs/outputs & data deps: Showdown text ↔ `localStorage.customsets` ↔ in-memory `SETDEX_*`.
- Entry-points (UI/CLI/API): Import modal buttons on every page; `?import=` base64 boot path in randoms controls.

### Dark-theme toggle
- What it does: Three-state theme (stored `true`/`false` → light/dark, unset → OS `prefers-color-scheme`) toggling the disabled flag on the dark stylesheet, persisted in localStorage.
- Where it lives (`path` refs): `src/js/dark-theme-toggle.js:16-26` (init + `updateTheme`), `32-40` (`toggleTheme` + change listeners incl. OS-change live update); styles `src/css/dark-theme.css`
- How it works: No jQuery; `darkStyles.disabled = !isDark`; `localStorage.darkTheme` string compare (`dark-theme-toggle.js:22`).
- Inputs/outputs & data deps: `localStorage`, `matchMedia('(prefers-color-scheme: dark)')`.
- Entry-points (UI/CLI/API): `#dark-theme-toggle` control in shared template.

### Curated competitive sets per gen + Champions
- What it does: Prebuilt `SETDEX_RBY/GSC/ADV/DPP/BW/XY/SM/SS/SV + SETDEX_CHAMPIONS` globals (`{species: {setName: {level/ability/item/nature/isGmax/teraType/ivs/evs/sps/moves}}}`) powering set dropdowns and honkalculate enumeration.
- Where it lives (`path` refs): `src/js/data/sets/gen1.js`–`gen9.js`, `champions.js` (emitted as `var SETDEX_* = {...}` by `import/src/set-import.ts:599-601`); minified on copy by `build:53-55`
- How it works: Regenerated from Showdown data (see importer); short-stat keys (`hp/at/df/sa/sd/sp`); level-100/natural-25-nature elision in `psetToCalcSet` (`set-import.ts:288-303`).
- Inputs/outputs & data deps: Generated artifact; consumed by `shared_controls.js:662,698-757`.
- Entry-points (UI/CLI/API): Loaded as globals by each HTML page.

### Tier/format lists per gen
- What it does: `{pokemon: tier}` maps per gen used for honkalculate tier checkboxes and OMs Tier Shift addons.
- Where it lives (`path` refs): `src/js/data/tiers/gen1.json`–`gen9.json`; consumed `honkalculate_controls.js` (checkbox list) and `oms_controls.js:188-196,353-367` (sync AJAX fetch of `js/data/tiers/gen${gen}.json`)
- How it works: Static JSON; `CONVERSION_G9/PREG9` maps tier → flat stat addon in OMs.
- Inputs/outputs & data deps: Static data.
- Entry-points (UI/CLI/API): UI-only.

### Static-site builder
- What it does: `node build [view]`: optionally `npm --prefix calc/ run compile`, copies `src/`→`dist/` (stripping `exports.* = void 0`, minifying sets), copies `calc/dist`→`dist/calc`, then wraps+hashes all five pages.
- Where it lives (`path` refs): `build:36-62` (`cpdir` with `.js` rewrite + sets minify), `64-74` (compile gate + 4 `cpdir/makeCachebuster` calls)
- How it works: `view` arg skips calc compile for UI-only iteration; hashes are content MD5 `[:8]` with random fallback.
- Inputs/outputs & data deps: `src/`, `calc/dist`, `src/shared.template.html` → `dist/`.
- Entry-points (UI/CLI/API): CLI (`node build`); CI publish (`test.yml:52`); `publish` script pushes `dist/` to `gh-pages`.

### Browser bundle helper
- What it does: Transpiles concatenated compiled calc JS (TS → broadly-compatible JS) and minifies to `production.min.js` for `<script>`-tag pages.
- Where it lives (`path` refs): `bundler.js:8-30` (Babel preset: TS/isTSX + class/properties/optional-chaining/spread/etc.), `32-57` (`Bundler.read` from `dist/` + `bundle(bundled, output)`)
- How it works: `babel.transformSync` then `terser.minify`; `read()` slices compiled-file headers/footers via `(b,e)` line args.
- Inputs/outputs & data deps: `calc/dist/*.js` → `dist/calc/production.min.js`-style output.
- Entry-points (UI/CLI/API): Called from `calc/` build flow (`EXPLORER-REPORT.md:9`).

### Optional Express API + static server
- What it does: Serves the built site and a JSON calc endpoint defaulting to gen 9.
- Where it lives (`path` refs): `server.js:1-9` (listen 3000 + `express.json()`), `11-28` (`GET /calculate`: required `attackingPokemon/defendingPokemon`, optional `*Options/moveName/field/gen`, `new calc.Pokemon/Move/Field` → `calc.calculate` → `res.json(result)`), `30` (`express.static('dist')`)
- How it works: Example-only: reads `req.body` on a GET (unusual), no validation beyond presence checks; serializes the full `Result` (rolls + desc + refs).
- Inputs/outputs & data deps: `calc` package (`require("calc")` at `server.js:2`).
- Entry-points (UI/CLI/API): HTTP `GET /calculate` + static `dist/`; `Dockerfile` runs it.

### Set/usage importer from Showdown data
- What it does: Regenerates all `SETDEX_*`: fetches dex sets + ladder usage from `data.pkmn.cc`, converts to calc sets (short-stat keys, Gmax suffix, Hidden Power IV fixup), validates each through `@pkmn/sim` TeamValidator, copies sets across similar formes (Mega/Primal/Zen/Meloetta/Aegislash/Keldeo/Minior/Palafin/Terapagos…), falls back to usage-spread sets under threshold for uncovered species/formats.
- Where it lives (`path` refs): `import/src/set-import.ts:449-463` (`fetchDexSets/fetchStats` from `data.pkmn.cc/sets|stats`), `465-564` (`importGen`: dex loop `473-515` + usage loop `516-562`), `205-243` (`dexToPset` + HP fixup), `244-286` (`usageToPset` via `top()`/`fromSpread`), `288-303` (`psetToCalcSet`: level-100/nature elision, `sps` for champions), `305-343` (`validatePSet` + shiny/0-EV/Arceus exceptions), `345-447` (`similarFormes`), `586-603` (CLI: `node dist/set-import.js <outDir>`, writes `gen{i}.js/champions.js`)
- How it works: Tier-ranked format ordering (`TIER_RANKINGS:75-100`, `getTierRanking:125-129`); usage threshold by battle count (`getUsageThreshold:147-153`); `statsIgnore` prevents usage-sets from duplicating dex-sets (`statsIgnore` at `472,497-498,536-537`).
- Inputs/outputs & data deps: Network (`data.pkmn.cc`) + `@pkmn/data/dex/sim/smogon`; outputs `src/js/data/sets/*.js`.
- Entry-points (UI/CLI/API): CLI; scheduled CI (`update.yml:25`).

### Sim differential smoke-tester
- What it does: Cross-checks calc against the real sim: generates random teams, drives them with `RandomPlayerAI`, mirrors each damaging `|-damage|` event into `calc.calculate`, and fails when sim damage isn't in the calc roll set (skipping Status/≤1-BP moves and Sash/Focus-Band saves).
- Where it lives (`path` refs): `smoketest/src/smoketest.ts:39-70` (`createCalcPoke`: boosts/curHP/Dynamax/fa​ints/status/Tera + types override), `71-108` (`createCalcSide/createCalcField`: hazards/screens/rooms/weather/auras/ruins), `109-121` (`createCalcMove`: single-hit override), `125-201` (`PreHandler`: `|move|/-crit/-enditem/-damage|` capture + roll-membership check), `232-290` (`run(gen,type)`: random-team battle loop + `./logs/<ts>.txt` artifact), `292-301` (10 random gen/format runs)
- How it works: `@pkmn/client` Battle mirror fed from `@pkmn/sim` `BattleStreams` protocol lines (`smoketest.ts:258-281`); known workaround for the calc's `overrides.types` 2-length merge quirk documented inline (`smoketest.ts:59-66`).
- Inputs/outputs & data deps: `@pkmn/sim/randoms/client/data/dex/protocol`; log files.
- Entry-points (UI/CLI/API): CLI (`node dist/smoketest.js [replay-file]` for single-replay mode at `212-230`, else random loop).

### Jest golden/unit suites
- What it does: Pins ranges/descriptions per gen plus unit-covers data/models/stats/utils: `calc.test.ts` (per-gen range/desc goldens via custom `toMatch`), `data.test.ts`, `pokemon.test.ts`, `move.test.ts`, `stats.test.ts`, `utils.test.ts`, with `helper.ts` (`inGen/inGens/tests` + gen-aware `toMatch` matcher that inherits expectations from lower gens).
- Where it lives (`path` refs): `calc/src/test/calc.test.ts:6-57` (golden style), `calc/src/test/helper.ts:8-60` (`calc/move/pokemon/field/side` factories), `62-82` (`inGens` range expansion), `84-131` (`tests()` fan-out `from..to`), `133-201` (`toMatch`: cumulative-breakdown lookup `Number(g) > gen → break`, then range/desc/result asserts)
- How it works: `npm run test` (jest+ts-jest per `calc/jest.config.js`); `toMatch`'s cumulative semantics mean a diff entry applies to its gen and all higher gens until overridden.
- Inputs/outputs & data deps: Library only.
- Entry-points (UI/CLI/API): `npm test` locally; CI `test.yml:32`.

### CI build/test + Pages publish and scheduled data update
- What it does: `test.yml`: on push/PR (or after a successful set update) installs, runs `npm run test`, then (push to master) rebuilds `dist/` and deploys to GitHub Pages. `update.yml`: every 3 days (`27 23 */3 * *`) reinstalls root+import, runs the set importer into `src/js/data/sets/`, commits/pushes if changed (which retriggers `test.yml` via `workflow_run`).
- Where it lives (`path` refs): `.github/workflows/test.yml:4-14` (triggers), `17-34` (build+test on node 18), `36-61` (Pages deploy: `node build` at `52`, artifact `dist` at `56`), `.github/workflows/update.yml:3-6` (schedule), `12-33` (fetch→import→commit/push)
- How it works: `workflow_run[Update Sets]` chaining guarantees freshly imported sets are tested before publish.
- Inputs/outputs & data deps: npm + network.
- Entry-points (UI/CLI/API): GitHub Actions only.

### Container build
- What it does: Ships the static site + API as a container (`node:12` base per explorer report, `node build`, then `server.js`).
- Where it lives (`path` refs): `Dockerfile`, `.dockerignore` (per `EXPLORER-REPORT.md:35`)
- How it works: Build bakes `dist/`; runtime serves `:3000`.
- Inputs/outputs & data deps: Same as `build` + `server.js`.
- Entry-points (UI/CLI/API): `docker build/run` → port 3000.

## Cross-feature notes
- Shared logic: `mechanics/util.ts` + `stats.ts` + `items.ts` + `desc.ts` + `data/interface.ts` are imported by every engine — the highest-blast-radius files. Rounding (`chainMods/pokeRound/OF16/OF32`), `getModifiedStat` tables, STAB, and speed must be changed with full cross-gen golden runs (`helper.ts:84-201` fan-out). Engines deliberately duplicate their `BasePower/BPMods/Attack/AtMods/Defense/DfMods/FinalMods` stage functions per era instead of sharing, so a fix must be ported to each affected `calculate*` family (gen789 vs champions vs gen56 vs gen4 vs gen3 vs gen12).
- `gen.num` is the dominant coupling axis: stat formula (`stats.ts:118-131`), boost tables (`util.ts:37-70`), speed/paralysis (`util.ts:130-135`), multihit squash tables (`desc.ts:863-969`), data-table index (`MOVES = [CHAMPIONS,RBY,…,SV]` at `moves.ts:5498`), and UI `LEGACY_STATS[gen]`/DV-vs-IV/`sps` switching (`shared_controls.js:38-40,1376-1413`) all key off it. Adding a generation = new tables + engine entry + `MECHANICS` slot + UI gating.
- In-place mutation contract: engines mutate `Pokemon/Move/Field` (Mold Breaker/Neutralizing Gas nulling, -ate/Weather Ball/Terrain Pulse rewrites, screen clearing, multihit boost stepping) — `calc.ts:34-40` cloning is load-bearing. Never call an engine directly with live UI objects; always go through `calculate()`. The smoketest's `types: [t, '']` workaround (`smoketest.ts:59-66`) papers over an `extend`-merge edge worth fixing in `pokemon.ts:53`.
- UI/API parity gaps: `server.js:11-28` accepts only a single `moveName` (no `overrides/hits/timesUsed/isCrit`) and reads GET bodies (many clients drop them) — a production API should accept full `State.Move` + POST + validation. OMs stat rewriting lives entirely in DOM base-stat inputs (`oms_controls.js:103-199`), so API/CLI callers can't reproduce MnM/Scale/TS without reimplementing that layer in `calc/`.
- Reimplementation hints (for a VGC calculator repo): reuse `adaptable.ts + data/interface.ts` with your own `Generation` rather than forking tables; copy the `gen789.ts` stage skeleton (BasePower→At→Df→Base→Final→16-roll→matrix) and delete unneeded branches; keep `desc.ts`'s `getKOChance/getHazards/getEndOfTurn/combine/squashMultihit` verbatim if you want %/hazard-aware KO text; keep `stats.ts` + `util.ts` rounding verbatim for cartridge fidelity; drive curated sets through the `set-import.ts` validate→`psetToCalcSet`→`similarFormes` pipeline instead of hand-authoring them.
