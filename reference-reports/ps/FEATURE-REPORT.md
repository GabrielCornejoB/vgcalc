> **Reference note (moved 2026-10-07):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/ps`. Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Feature Report — ps

> Sources: ARCHITECTURE-REPORT.md (inventory), EXPLORER-REPORT.md, direct reads. ARCHITECTURE-REPORT.md was present and used as the feature-discovery base; all sections below were verified with fresh repo reads.

## Feature inventory
| Feature | One-line purpose | Key files |
|---|---|---|
| Shared branded types | Compile-time-only vocab (`ID`, `PokemonSet`, stat/tier brands) | `types/index.d.ts` |
| Unified static dex | Gen-layered species/moves/items/abilities lookup w/ `forGen`/`mod` | `dex/index.ts` |
| Competitive data façade | Existence-filtered per-gen `Abilities/Items/Moves/Species/Types/...` API | `data/index.ts` |
| Stat calculator | Gen-aware stat formula, DV/IV conversion, HP DV | `data/index.ts:714-779` |
| Hidden-Power calculator | IV→type/power math w/ gen≤2 DV path | `dex/index.ts:1330-1368` |
| Learnset legality queries | Evo-chain `learnable/canLearn/isLegal` + Pentagon/Plus/Galar/Paldea shorthands | `data/index.ts:519-690` |
| Type-effectiveness system | `damageTaken` 0/1/2/3 → `effectiveness` map + `canDamage/totalEffectiveness` | `data/index.ts:386-509`, `dex/index.ts:1288-1328` |
| Battle protocol spec | Typed `ArgType/KWArgType` tuples for ~100 `|move|/\|-damage\|/...` + room/global/tournament | `protocol/src/index.ts` |
| Battle protocol parser | `parse/parseLine/parseBattleLine/parseDetails/parseHealth/parseRequest` + legacy `upgradeArgs` | `protocol/src/index.ts:1954-2430` |
| Protocol verifier | Per-message boolean predicate suite + `verifyKWArg(s)` | `protocol/src/verifier/index.ts` |
| Client battle engine | `Battle/Side/Pokemon/Field` mirror updated from protocol lines + `request` reconcile | `client/src/battle.ts`, `client/src/handler.ts`, `client/src/side.ts`, `client/src/field.ts` |
| Illusion / forme / Tera tracking | Zoroark/Transform/forme-change/terastallize identity reconciliation | `client/src/battle.ts:263-344`, `client/src/pokemon.ts:90-91,201,615-624` |
| PP / Pressure / move-slot tracking | `rememberMove/useMove/mergePP` w/ Z/Max/Pressure accounting | `client/src/pokemon.ts:412-568` |
| Damage-range estimator | Pixel (`/48`) vs percentage HP range math + formatting | `client/src/pokemon.ts:652-718`, `client/src/battle.ts:435-459` |
| Weather / terrain / side-condition tracking | `Field` weather/terrain/pseudo-weather + `Side.sideConditions` duration tables | `client/src/field.ts`, `client/src/side.ts:27-35` |
| Log formatter | Template-driven text/HTML battle-log renderer w/ `Tracker` snapshot | `view/src/formatter.ts`, `view/src/data/text.json` |
| Choice builder | Request-aware `move/switch/team/shift/pass` choice-string state machine | `view/src/choices.ts` |
| Set serializer | Packed/export/import/canonicalize for single sets | `sets/src/sets.ts` |
| Team serializer | Multi-team pack/unpack/import/export with `===` headers | `sets/src/teams.ts` |
| Random-team generators | Per-gen (`gen1`–`gen9`) `getTeam` behind format-mod registry | `randoms/src/index.ts`, `randoms/src/gen1.ts`–`gen9.ts` |
| Format mods | Per-mod overlays applied via `Dex.mod(id, modData)` | `mods/src/*/index.ts`, `mods/src/index.ts`, `dex/index.ts:1276-1286` |
| Vendored simulator | Generated Showdown battle engine + `TeamValidator`/tools | `sim/` (`sim/package.json`) |
| Async stream primitives | `Read/Write/Object` streams bridging Node and battle I/O | `streams/index.ts` |
| Login actions | Pure `register/login/upkeep/rename/logout` HTTP-action builders | `login/index.ts` |
| Sprite resolver | Species/item/icon → image URL logic, gen-gated graphics sets | `img/src/index.ts`, `img/src/adaptable.ts` |
| Integration demo + harness | Parcel browser UI + cross-package tests | `integration/src/ui/index.ts`, `integration/index.html` |
| Monorepo build/publish tooling | `tsc` dual-format + `tsup`/`esbuild` IIFE + submodule import/generation | `build`, `import`, `publish`, `package.json` |

## Feature details
### Shared branded types
- What it does: Defines the compile-time contract every package imports — no runtime code.
- Where it lives (`path` refs): `types/index.d.ts:1-88`.
- How it works: `ID` is a triple-union brand (`types/index.d.ts:2`); `As<T>` branding (`types/index.d.ts:4`) underlies `Weather/FieldCondition/SideCondition` (`types/index.d.ts:6-8`); enums-as-unions for `GenerationNum` (`types/index.d.ts:10`), `StatID/StatsTable/BoostID/BoostsTable` (`types/index.d.ts:14-18`), `GameType` (`types/index.d.ts:40-41`), tiers (`types/index.d.ts:61-68`), and the full `PokemonSet` interface (`types/index.d.ts:70-88`).
- Inputs/outputs & data deps: None — pure `.d.ts`.
- Entry-points (UI/CLI/API): Imported as `@pkmn/types` by all packages.

### Unified static dex
- What it does: Bundles `data/*.json` and exposes versioned lookups for every data kind with gen inheritance.
- Where it lives (`path` refs): `dex/index.ts:1-18` (JSON imports), `dex/index.ts:1228-1442` (`ModdedDex`), `dex/index.ts:88-120` (`DexConditions` cache pattern).
- How it works: `new ModdedDex(modid, modData)` (`dex/index.ts:1258-1263`); `mod(id[, modData])` memoizes per-mod dexes (`dex/index.ts:1276-1281`); `forGen(n)` validates 1–9 and delegates to `mod('genN')` (`dex/index.ts:1283-1286`); `loadData` seeds each table from `DATA[type][gen]` then walks the parent chain (`forGen(gen+1)` or `Scripts.inherit`), merging missing keys / `{inherit:true}` partials and deep-cloning only `Species` (`dex/index.ts:1382-1442`); per-kind tables (`DexAbilities/DexMoves/DexSpecies/...`) expose `get/getByID` with memo caches and normalization (Hidden Power slicing, gen≤3 physical/special recategorization); constructors compute derived fields (`Move` Z/Max tables, `Species` `bst`/gender/gen inference). Null-object `EMPTY_CONDITION` (`dex/index.ts:86`) avoids crashes on unknown ids.
- Inputs/outputs & data deps: Input: `modid`/`gen` + optional `ModData`; data source: `dex/data/*.json`. Output: typed `Ability/Item/Move/Species/Type/Condition` objects.
- Entry-points (UI/CLI/API): `Dex.forGen(n)`, `Dex.mod(id, data)`, `dex.abilities/items/moves/species/types/...get(name)`.

### Competitive data façade
- What it does: Filters raw dex through a legality predicate and reshapes species/types for competitive use.
- Where it lives (`path` refs): `data/index.ts:30-35` (`DEFAULT_EXISTS`), `data/index.ts:57-130` (`Generations`/`Generation`), `data/index.ts:132-...` per-kind facades, `data/index.ts:295-346` (`Specie`), `data/index.ts:398-509` (`Types`/`Type`).
- How it works: `Generations(dex, exists)` (`data/index.ts:57-68`) caches per-gen; `get(g)` parses gen from string/number and wraps `dex.forGen(gen)` (`data/index.ts:70-76`); `Generation` ctor builds `abilities/items/moves/species/types/natures/learnsets/conditions/stats` sub-APIs (`data/index.ts:104-117`); each `get` re-checks `exists` (drops `!exists`, `isNonstandard`, `noability`, `Illegal/Unreleased` per `data/index.ts:30-35`); `Specie` rewrites gender/abilities per gen (gen<2 strips gender, gen<3 single ability, gen 3–4 filters by ability gen, H only ≥5, S only ≥7) and filters evos/formes through `exists` (`data/index.ts:295-329`); `toID` normalizer (`data/index.ts:51-55`).
- Inputs/outputs & data deps: Input: raw `Dex` + `ExistsFn`. Output: `Generation` object.
- Entry-points (UI/CLI/API): `new Generations(Dex).get(9)`, `Generation.get(dex, gen)`, `gen.species/moves/items/abilities/types.get(name)`.

### Stat calculator
- What it does: Implements the canonical stat formula plus gen<3 DV path, nature modifiers, and HP-DV helpers.
- Where it lives (`path` refs): `data/index.ts:714-779` (`Stats` class).
- How it works: `calc(stat, base, iv, ev, level, nature)` (`data/index.ts:721-737`) — defaults `ev` to 252 pre-gen-3 else 0; pre-gen-3 converts IV→DV (`toDV(iv)*2`) and drops nature; HP branch special-cases `base===1` (Shedinja) else `tr(tr(2*base+iv+tr(ev/4)+100)*level/100+10)`; other stats apply `+5` then nature `tr(tr(val*110,16)/100)` / `tr(tr(val*90,16)/100)` with 16-bit truncation `tr(num,bits)` (`data/index.ts:37`); `getHPDV` combines atk/def/spe/spa DV parity bits (`data/index.ts:757-764`); `toDV/toIV` (`data/index.ts:772-778`); `display` maps `spa→spc` in gen 1 (`data/index.ts:743-748`).
- Inputs/outputs & data deps: Inputs: base/IV/EV/level/nature + `dex.gen`. Output: integer stat.
- Entry-points (UI/CLI/API): `gen.stats.calc(...)`, used by `sets` EV back-calculation (`canonicalize`).

### Hidden-Power calculator
- What it does: Derives Hidden Power type/power from IVs with separate gen≤2 and gen≥3 paths.
- Where it lives (`path` refs): `dex/index.ts:1330-1368` (`ModdedDex.getHiddenPower`); delegated by `data/index.ts:445-447` (`Types.getHiddenPower`).
- How it works: Gen≤2 path converts IVs→DVs (`tr(iv/2)`), type = `HP_TYPES[4*(atkDV%4)+(defDV%4)]`, power from bit-shifted DV high bits (`dex/index.ts:1336-1351`); gen≥3 path accumulates `hpTypeX/hpPowerX` over `hp/atk/def/spe/spa/spd` parity bits, type = `HP_TYPES[tr(x*15/63)]`, power = `tr(p*40/63)+30` pre-gen-6 else flat 60 (`dex/index.ts:1352-1367`).
- Inputs/outputs & data deps: Input: `StatsTable` IVs + `gen`. Output: `{type, power}`.
- Entry-points (UI/CLI/API): `dex.getHiddenPower(ivs)`, `gen.types.getHiddenPower(ivs)`; consumed by `sets` HP-type inference.

### Learnset legality queries
- What it does: Answers "what can this species learn / can it learn X / is this source legal" by walking the evo chain.
- Where it lives (`path` refs): `data/index.ts:519-690`.
- How it works: `all(species)` async generator yields the species' learnset then walks `battleOnly→changesFrom→prevo` with Lycanroc-Dusk/Rockruff/Gastrodon-East/Pumpkaboo-Super special cases (`data/index.ts:547-576`); `learnable(name, restriction)` merges per-evo `moveid→MoveSource[]`, filters through `isLegal` + `+source[0] <= gen.num`, dedupes by 2-char source prefix (`data/index.ts:580-630`), caches per species when unrestricted (`data/index.ts:585-587,628`); `canLearn` short-circuits on first legal learnset (`data/index.ts:634-648`); `isLegal(move, sources, gen|restriction)` (`data/index.ts:652-689`) handles `Pentagon` (has gen-6 source), `Plus`/`Galar` (gen-7/8 non-VC-only), `Paldea` (has gen-9), gen-3/4 HM cross-gen propagation (`GEN3_HMS/GEN4_HMS`), else `minGen..9 includes gen.num`. Explicitly teambuilder-only — real legality needs `sim.TeamValidator` (`data/index.ts:578-579,632-633,650-651`).
- Inputs/outputs & data deps: Inputs: species/move names, optional `Restriction`; async dex learnsets. Outputs: `{moveid: sources[]}` map / boolean / boolean|undefined.
- Entry-points (UI/CLI/API): `gen.learnsets.learnable/canLearn/isLegal/get/all`.

### Type-effectiveness system
- What it does: Converts Showdown's `damageTaken` encoding into usable effectiveness multipliers and immunity checks.
- Where it lives (`path` refs): `data/index.ts:386-509`, `dex/index.ts:1288-1328`.
- How it works: `ModdedDex.getImmunity` checks `damageTaken[source]===3` per target type (`dex/index.ts:1288-1304`); `getEffectiveness` sums `+1/-1/0` per type from the same encoding (`dex/index.ts:1306-1328`); `Types.canDamage/totalEffectiveness` (`data/index.ts:449-458`) wrap those and map the summed encoding through `EFFECTIVENESS = {-3:0.125,...3:8}` (`data/index.ts:386-394`); `Type` ctor precomputes a dealt-damage `effectiveness` table via `DAMAGE_TAKEN[data.damageTaken[name]]` (`data/index.ts:466-491`) and infers physical/special category for gens ≤3 via `SPECIAL` list (`data/index.ts:464,483-484`); `???` handled by synthetic `unknown` type for gens 2–4 (`data/index.ts:405-428`).
- Inputs/outputs & data deps: Inputs: source type + target (name or `{types}`/`getTypes`). Output: boolean / 0/0.125/.../8.
- Entry-points (UI/CLI/API): `gen.types.canDamage/totalEffectiveness`, `type.canDamage/totalEffectiveness`, `dex.getImmunity/getEffectiveness`.

### Battle protocol spec
- What it does: Type-level specification of every server↔client message shape.
- Where it lives (`path` refs): `protocol/src/index.ts:22-~1950` (namespace `Protocol`).
- How it works: Branded `PokemonIdent/Details/HPStatus/SearchID` (`protocol/src/index.ts:67-94`); `Request` discriminated union (`move/switch/team/wait`) with `active[]/forceSwitch/side.pokemon` slots; `BattleInit/Major/Minor/Global/Tournament Args` as positional tuple types (e.g. `|move|`, `|-damage|`, `|switch|`); `KWArgs` (`[from]/[of]/[still]/...`) typed per message; `Handler` interface maps each `|cmd|` to a method signature — the contract `client.Handler` and `view.Handler` implement.
- Inputs/outputs & data deps: None at runtime — types only.
- Entry-points (UI/CLI/API): `import {Protocol, ArgType, Args, KWArgs} from '@pkmn/protocol'`.

### Battle protocol parser
- What it does: Turns raw server lines into typed `(args, kwArgs)` plus structured detail/health/request objects.
- Where it lives (`path` refs): `protocol/src/index.ts:1954-2430`.
- How it works: `parse(data)` splits `\n`, handles leading `>roomid`, yields `{roomid,args,kwArgs}` per line (`protocol/src/index.ts:1954-1966`); `key(args)` maps to `|cmd|` (or `|tournament|x|`) and returns undefined for unknown (`protocol/src/index.ts:1968-1973`); `parseLine` splits `|cmd|...` with 3-part (`c/chat/uhtml`) and 4-part (`c:/pm`) special cases (`protocol/src/index.ts:1975-2001`); `parseBattleLine` additionally strips trailing `[kw]value` kwargs (`protocol/src/index.ts:2003-2025`); `parsePokemonIdent` splits `p1a: Name` into player/position/name (`protocol/src/index.ts:2027-2043`); `parseDetails` extracts species/level/gender/shiny/`tera:TYPE` (`protocol/src/index.ts:2045-2079`); `parseHealth` parses `CUR/MAX[r/y/g] STATUS` incl. faint and `hpcolor` (`protocol/src/index.ts:2081-...`); `parseRequest` normalizes request JSON (fills `requestType`, normalizes `canZMove/maxMoves`, calls `parseDetails+parseHealth` per mon); `upgradeArgs/upgradeBattleArgs` normalize legacy forms (`-activate→-block/-start`, `Snow→Snowscape`, `-restoreboost→-clearnegativeboost`, `-nothing→Splash`).
- Inputs/outputs & data deps: Input: raw string lines / request JSON. Output: typed args + kwArgs. No data deps.
- Entry-points (UI/CLI/API): `Protocol.parse/parseLine/parseBattleLine/key/parseDetails/parseHealth/parseRequest`.

### Protocol verifier
- What it does: Validates that a server line is well-formed (and gen-legal) without mutating state.
- Where it lives (`path` refs): `protocol/src/verifier/index.ts:1-60` (tables), `:217-253` (predicate core), `:653-697` (gen-gated KW checks).
- How it works: Parallel `Handler<boolean>` — one predicate per protocol message returning true iff args/kwargs match expected arities, enum membership (`QUERYTYPES/GAMETYPES/PLAYERS/STATUSES/BOOSTS/REASONS` in `verifier/index.ts:29-37`), and `BOOL_KWARGS/ABILITY_KWARGS/MOVE_KWARGS/...` classification (`verifier/index.ts:38-47`); `verifyKWArg(s)` checks trailing `[kw]` names with gen-gated weather/condition tables (`WEATHER/FIELD_CONDITIONS/SIDE_CONDITIONS` in `verifier/index.ts:50-59`).
- Inputs/outputs & data deps: Inputs: parsed args + optional gen-gated `Generation` API. Output: boolean.
- Entry-points (UI/CLI/API): `import {...} from '@pkmn/protocol/verifier'`; used by tests/harnesses, not the hot battle path.

### Client battle engine
- What it does: Event-sourced mirror of the server battle — mutates `Battle/Side/Pokemon/Field` per protocol line.
- Where it lives (`path` refs): `client/src/battle.ts:25-199+`, `client/src/handler.ts:15-...`, `client/src/side.ts:9-350`, `client/src/field.ts:42-124`.
- How it works: `Battle.add(line|args)` parses via `Protocol.parseBattleLine`, keys on `|cmd|`, dispatches to same-named `Handler` method (`client/src/battle.ts:96-104`); `update(request)` reconciles `request.side.pokemon` into `side.team/active` by `searchid`/`checkDetails`, refreshes `moveSlots/maxMoves/zMoves/Dynamax/Tera/trapped` on `move` requests (`client/src/battle.ts:106-193`); `Handler` (~50 methods: `|switch|/|drag|/|replace|` → `getSwitchedPokemon` + `side.switchIn/dragIn/replace`; `|-damage|/|-heal|` → health parse; `|-boost|` family → `boosts{}`; `|-weather|/|-fieldstart|/|-sidestart|` → field/side tables; `|-item|/|-enditem|/|-ability|` bookkeeping); `Side` owns `team/active/sideConditions{min,maxDuration}` with `switchIn/dragIn/replace/switchOut/faint` lifecycle and Team-Preview/Illusion dedupe; `Field.setWeather/setTerrain/addPseudoWeather` tracks durations incl. gen≤5/extreme-weather special cases (`client/src/field.ts:82-110`); `Battle.get()` returns `NULL` (`client/src/battle.ts:16,197-199`) when dex is stale.
- Inputs/outputs & data deps: Inputs: protocol lines + `Generations` data + optional `PokemonSet[]`. Output: mutated battle graph readable by UI/formatter.
- Entry-points (UI/CLI/API): `new Battle(gens, player?, sets?)`, `battle.add(line)`, `battle.update(request)`; DI seams: `Battle` ctor `field/side` factories (`client/src/battle.ts:59-65`), `Side` ctor `provider` (`client/src/side.ts:42-47`).

### Illusion / forme / Tera tracking
- What it does: Keeps displayed identity correct under Zoroark Illusion, Transform, forme changes, and Terastallization.
- Where it lives (`path` refs): `client/src/battle.ts:263-344` (`getSwitchedPokemon`), `client/src/pokemon.ts:90-91,201,615-624`, `client/src/handler.ts:668-738`.
- How it works: Identity key is `searchid = ident|details` (`protocol/src/index.ts:83`, `battle.ts:268`); `getSwitchedPokemon` first exact-matches `searchid`, falls back to `checkDetails` Illusion-aware search; `Pokemon.illusion/revealedDetails` (`pokemon.ts:90-91`) stores the hidden mon, `toString` renders `illusion (displayed)` (`pokemon.ts:201`); `copyTypesFrom` propagates Transform types via `typechange/typeadd` volatiles (`pokemon.ts:615-624`); `details` parsing captures `tera:TYPE` suffix; `getPokemon`/`pokemonAt`/`getSwitchedOutPokemon` expose pre-mutation snapshots for the formatter.
- Inputs/outputs & data deps: Inputs: `|switch|/|replace|/|detailschange|/|-terastallize|` lines. Output: correct `speciesForme/types/illusion` on `Pokemon`.
- Entry-points (UI/CLI/API): Automatic inside `battle.add`; observable via `pokemon.illusion/speciesForme/terastallized`.

### PP / Pressure / move-slot tracking
- What it does: Tracks known moves and remaining PP from observed usage, including Pressure doubling and Z/Max quirks.
- Where it lives (`path` refs): `client/src/pokemon.ts:412-568`.
- How it works: `rememberMove(name, ppUsed)` merges into `moveSlots[]` via `mergePP`, propagates through Transform with `virtual` flag and recursion guard (`pokemon.ts:447-465`); `useMove(move, target, from)` clears move statuses, resolves Sleep Talk / Z-move name rewriting (with documented `zPower=true // BUG`), computes extra PP from Pressure by scanning foe targets by `moveTarget` class (singles/FFA/allies-in-gen≤4 rules) and `effectiveAbility()==='Pressure'` (+1, or `[n,n+1]` tuple when gen-3 ambiguity) (`pokemon.ts:467-547`); `getPressurePP` (`pokemon.ts:548-568`); `cantUseMove` records 0-PP observations (`pokemon.ts:570-590`).
- Inputs/outputs & data deps: Inputs: `|move|/|cant|` lines + gen data. Output: `pokemon.moveSlots[{name,ppUsed,pp/maxpp,target,disabled,virtual}]`.
- Entry-points (UI/CLI/API): Automatic via handler; inspectable as `pokemon.moveSlots`.

### Damage-range estimator
- What it does: Converts observed HP-bar deltas into min–max damage fractions for log tooltips.
- Where it lives (`path` refs): `client/src/pokemon.ts:652-718`, `client/src/battle.ts:435-459` (`damagePercentage` Tracker impl).
- How it works: `getDamageRange([delta, denom, oldnum, oldcolor], hpcolor)` (`pokemon.ts:653-674`) — non-`/48` denominators return exact ratio; `/48` pixel mode diffs `getPixelRange(before)` vs `getPixelRange(after)` with swap/zero-delta handling; `getPixelRange(pixels, color)` (`pokemon.ts:676-698`) maps 0/1/9/24/48 edge cases with `epsilon=0.5/714` and yellow/red/green threshold logic; `getFormattedRange(range, precision, sep)` (`pokemon.ts:700-718`) renders `x%` or `lo–hi%`.
- Inputs/outputs & data deps: Inputs: parsed HP deltas + colors. Output: `[min,max]` fraction + formatted string.
- Entry-points (UI/CLI/API): `Pokemon.getDamageRange/getPixelRange/getFormattedRange`; `battle.damagePercentage(ident, hpstring)` (used by `LogFormatter` via `Tracker`).

### Weather / terrain / side-condition tracking
- What it does: Maintains field-level conditions with duration windows.
- Where it lives (`path` refs): `client/src/field.ts:42-124`, `client/src/side.ts:27-35`.
- How it works: `Field` holds `weather/weatherState/terrain/terrainState/pseudoWeather` (`field.ts:42-49`); `setWeather` maps ids (`sandstorm→Sand`, `snowscape→Snow`, ...) with `WEATHERS` (`field.ts:12-21`), extreme weather (`desolateland/primordialsea/deltastream`) gets 0 duration, else gen-gated 5/8 windows (`field.ts:82-110`); `setTerrain` similar with +2 persist bonus (`field.ts:68-80`); upkeep ticks decrement `maxDuration/minDuration` (`field.ts:82-89`); `Side.sideConditions` stores `{name,level,minDuration,maxDuration}` per condition id.
- Inputs/outputs & data deps: Inputs: `|-weather|/|-fieldstart|/|-sidestart|/|upkeep|` lines. Output: readable `battle.field` / `side.sideConditions`.
- Entry-points (UI/CLI/API): Automatic via handler; read via `battle.field.weather/terrain`, `side.sideConditions`.

### Log formatter
- What it does: Renders parsed protocol args into human text/HTML using data-driven templates.
- Where it lives (`path` refs): `view/src/formatter.ts:85-...`, `view/src/formatter.ts:229-363` (`template/lineSection/sectionBreak/formatText/formatHTML`), `view/src/data/text.json`.
- How it works: `template(type, ...namespaces)` (`formatter.ts:229-...`) looks up `text.json` with per-effect + `GenN` overrides; `lineSection` classifies `break/preMajor/major/postMajor` (`formatter.ts:270-...`); `sectionBreak` inserts `\n` on transitions (`formatter.ts:311-...`); `formatText(args,kwArgs)` = sectionBreak + inner `Handler` method filling `[POKEMON]/[MOVE]/[EFFECT]/[STAT]` placeholders via `pokemon()/trainer()/effect()` helpers; `formatHTML` wraps `formatText` lines. Requires a `Tracker` snapshot (`formatter.ts:45-65`: `pokemonAt/damagePercentage/currentWeather/getSwitchedOutPokemon/getPokemonTypeList/getPokemonSpeciesForme`) — `Battle` implements it (`battle.ts:430-468`) — and **messages must hit the formatter before `client`'s `Handler`** (`formatter.ts:35-65`).
- Inputs/outputs & data deps: Inputs: `(args, kwArgs)` + tracker. Data: `text.json` templates. Output: string (text or HTML).
- Entry-points (UI/CLI/API): `new LogFormatter(perspective?, tracker?)`, `.formatText(args,kwArgs)`, `.formatHTML(args,kwArgs)`.

### Choice builder
- What it does: Builds legal server choice strings (`move N .../switch N`) from a `|request|`.
- Where it lives (`path` refs): `view/src/choices.ts:32-351`.
- How it works: State machine over `Protocol.Request` (`move/switch/team/wait`); ctor `fillPasses()` pre-fills `pass` for inactive/forced slots (`choices.ts:53-73`); `parseChoice` accepts move index|name + `mega/zmove/max/terastallize/targetLoc` suffixes and switch by index|name with fuzzy match; dedupes `alreadySwitchingIn/Mega/Z/Max/Tera` (`choices.ts:47-51`); buffers partial `current` until target known for multi-target moves; `stringChoice/toString` emits comma-joined protocol strings. `Data` interface (`choices.ts:28-30`) allows name resolution without importing dex.
- Inputs/outputs & data deps: Input: `Protocol.Request` (+ optional species data). Output: choice string the server consumes.
- Entry-points (UI/CLI/API): `new ChoiceBuilder(request)`, `.addChoice(str, data?)`, `.toString()`, `.isDone()/isEmpty()`.

### Set serializer
- What it does: Converts single Pokémon sets between in-memory, packed (`|`-delimited `/utm`), and human Showdown text.
- Where it lives (`path` refs): `sets/src/sets.ts:87-470` (`packSet/unpackSet/exportSet/importSet/canonicalize`), `sets/src/sets.ts:20-39` (`Data` interface).
- How it works: `packSet` (`sets.ts:92-...`) emits `name|species|item|ability|moves|nature|evs|gender|ivs|shiny|level|happiness,hpType,pokeball,...`; `unpackSet`/`_unpack` reverses incl. ability-index (`0/1/H/S`), EVs/IVs, `S` shiny, trailing misc fields; `exportSet` (`sets.ts:203-...`) prints human text with gen-gated fields and Hidden Power IV inference; `importSet` (`sets.ts:338-...`) parses it back; `canonicalize` (`sets.ts:361-...`) stable-sorts + back-calculates EVs via `Stats.calc` and rewrites HP-type IVs; `fromJSON` documented unvalidated. `Data` seam avoids hard dependency on `dex`.
- Inputs/outputs & data deps: Inputs: `Partial<PokemonSet>` / strings. Data: optional `Data` for name resolution.
- Entry-points (UI/CLI/API): `Sets.packSet/unpackSet/exportSet/importSet/canonicalize/fromJSON/toJSON`, `Set.import/export/pack/unpack`.

### Team serializer
- What it does: Same conversions at team granularity, plus multi-team files.
- Where it lives (`path` refs): `sets/src/teams.ts:8-244` (`Team`, `Teams`).
- How it works: `packTeam/unpackTeam` join/split `]`-joined packed sets (`teams.ts:93-125`); `importTeam/exportTeam` handle human text; `importTeams/exportTeams` handle `=== [format] folder/name ===` multi-team files plus single-packed-line builder format; `Team` ctor derives gen from `format` string (`gen6` default for legacy formats) via `data.forGen` (`teams.ts:22-29`); `fromJSON` unvalidated (`teams.ts:69-76`); `canonicalize` sorts lead + rest by ID.
- Inputs/outputs & data deps: Same as sets, at `PokemonSet[]` level.
- Entry-points (UI/CLI/API): `Teams.packTeam/unpackTeam/importTeam/exportTeam/importTeams/exportTeams`, `Team.pack/unpack/import/export/fromString/toJSON`.

### Random-team generators
- What it does: Generates legal Random Battle teams per generation.
- Where it lives (`path` refs): `randoms/src/index.ts:21-50`, `randoms/src/gen1.ts`–`gen9.ts`, `randoms/src/utils.ts`.
- How it works: `GENERATORS` maps `format.mod` → per-gen class (`RandomGen1Teams…RandomTeams`) (`index.ts:21-31`); `TeamGenerators.getTeamGenerator(format, seed)` resolves `Dex.formats.get`, requires `format.team==='random'`, instantiates with `Dex.mod(format.mod)` (`index.ts:39-50`); each class exposes `getTeam(options?)/setSeed(prng?)`; heavy per-gen logic (sets, items, levels, team composition) lives in `genN.ts`.
- Inputs/outputs & data deps: Inputs: format name + seed. Deps: `@pkmn/sim` (`Dex/Format/PRNG`). Output: `PokemonSet[]`.
- Entry-points (UI/CLI/API): `TeamGenerators.getTeamGenerator('gen9randombattle', seed).getTeam()`; also plugged into sim via `Teams.setGeneratorFactory(TeamGenerators)` (see integration).

### Format mods
- What it does: Provides per-format data overlays (old gens, DLC, community formats).
- Where it lives (`path` refs): `mods/src/gen1jpn/index.ts`, `gen1stadium/`, `gen2stadium2/`, `gen3rs/`, `gen3frlg/`, `gen4pt/`, `gen5bw1/`, `gen6xy/`, `gen7sm/`, `gen7letsgo/`, `gen8dlc1/`, `gen9dlc1/`, `gen9predlc/`, `champions/`, `championsregma/` (glob-verified); wrapper `mods/src/index.ts:27-171` (`ModdedDex` delegate); consumed via `dex/index.ts:1276-1286`.
- How it works: Each mod dir exports `pokedex/moves/items/abilities/scripts/formats/learnsets` partials typed as `ModData`; `Dex.mod(id, modData)` constructs a layered `ModdedDex` whose `Scripts.inherit` points at the parent gen; `mods/src/index.ts` itself is only a typed delegate wrapper, not data.
- Inputs/outputs & data deps: Inputs: mod id + overlay JSON/TS. Output: gen-correct `ModdedDex`.
- Entry-points (UI/CLI/API): `Dex.mod('gen9dlc1', modData)`, `import ... from '@pkmn/mods/gen9dlc1'` etc.

### Vendored simulator
- What it does: Full authoritative battle engine + team validator (generated, not hand-written here).
- Where it lives (`path` refs): `sim/` (generated `sim/`, `data/`, `lib/`, `config/`, `test/`); `sim/package.json:1-51`.
- How it works: Extraction of `pokemon-showdown` produced by `./import` + `./build`; exposes `BattleStreams`, `Dex`, `PRNG`, `RandomPlayerAI`, `TeamValidator`, `Teams` (consumed in `integration/src/ui/index.ts:10,27-28`); `pretest` rebuilds via `../build` (`sim/package.json:48`); tests run under mocha (`sim/package.json:45`). No hand-written sim logic lives here — behavior bugs must be fixed upstream.
- Inputs/outputs & data deps: Inputs: formats/teams/seeds. Output: simulated battle streams, validation errors.
- Entry-points (UI/CLI/API): `import {BattleStreams, TeamValidator, ...} from '@pkmn/sim'`; `sim/tools` subpath.

### Async stream primitives
- What it does: Small async stream substrate bridging Node streams and battle I/O.
- Where it lives (`path` refs): `streams/index.ts:20-874`.
- How it works: `ReadStream` (buffered `push/pushEnd/loadIntoBuffer/read`, Node-stream adapters, `BUF_SIZE=65536*4` in `streams/index.ts:15,20-...`); `WriteStream`/`ObjectReadStream` siblings with `pipeTo`; used by `sim` (`BattleStreams`) to shuttle protocol lines between engine and AIs/clients.
- Inputs/outputs & data deps: Inputs: strings/buffers/Node streams. Output: async-iterable battle I/O.
- Entry-points (UI/CLI/API): `import {ReadStream, WriteStream, ...} from '@pkmn/streams'`.

### Login actions
- What it does: Builds the HTTP requests needed to authenticate against Showdown's login server — without performing I/O.
- Where it lives (`path` refs): `login/index.ts:43-199`.
- How it works: Pure builders returning `{method,url,headers,data,onResponse}` (`login/index.ts:22-29`): `register` (`:44-63`), `login` (delegates to `rename` when passwordless, `:66-84`), `upkeep` (`:87-99`), `rename`→`getassertion` (`:102-112`), `logout`→`'¦/logout'` (`:115-123`); `action()` posts to `${url}api/${api}` (`:125-140`); `onResponse` parses assertion JSON (`safeJSON`, `:153-157`) and emits `¦/trn user,0,assertion` via `next()` (`:165-167`) after `process()` validation (MitM/empty/Google-auth guards, `:176-198`). Caller performs the HTTP fetch. Default endpoint `https://play.pokemonshowdown.com/` (`:1`), default captcha `Pikachu` (`:2`).
- Inputs/outputs & data deps: Inputs: `LoginDetails/RegistrationDetails` (username/challstr/password/serverid/url). Output: `Action` + eventually `¦/trn …` string.
- Entry-points (UI/CLI/API): `import {login, register, upkeep, rename, logout, Actions} from '@pkmn/login'`; demo `login/index.html`.

### Sprite resolver
- What it does: Resolves species/item/icon → hosted image URLs with per-gen graphics sets.
- Where it lives (`path` refs): `img/src/index.ts:1-7`, `img/src/adaptable.ts:1-...`, `img/src/data/`.
- How it works: `Sprites`/`Icons` classes instantiated over `Data` (`index.ts:4-5`); `GENS` maps graphics-gen labels to numeric gens (`gen1rg/gen1rb/gen1→1 … gen5/gen5ani→5, ani→6`) (`adaptable.ts:15-38`); `SOURCES` maps UI labels (`Red/Blue`, `Black/White (Animated)`, `Modern (Animated)`) to gens (`adaptable.ts:60-82`); `ANIMATED`/`FRAME2` track animated and second-frame variants (`adaptable.ts:42-58`); `NONDP` handles mid-gen-4 additions missing from `gen4dp` (`adaptable.ts:93-95`); base URL builder defaults to `https://play.pokemonshowdown.com` (`adaptable.ts:8-13`).
- Inputs/outputs & data deps: Inputs: species/item + gen/graphics-gen/animated options. Output: `{gen,w,h,url,pixelated?}` (`adaptable.ts:84-90`).
- Entry-points (UI/CLI/API): `import {Sprites, Icons} from '@pkmn/img'`; demo `img/index.html`.

### Integration demo + harness
- What it does: Proves all packages work together in a real (browser) battle.
- Where it lives (`path` refs): `integration/src/ui/index.ts:1-50+`, `integration/index.html`, `integration/test/`.
- How it works: Picks a random format (`gen1randombattle…gen9randomdoublesbattle`), wires `Dex.forFormat` + `TeamGenerators` as sim generator factory + `Generations(Dex)` + `Teams.importTeam` + `TeamValidator.validateTeam` (`ui/index.ts:17-36`); maps numeric gens to sprite sets (`SPRITES`, `ui/index.ts:38-49`); runs `BattleStreams` sim battles while mirroring into `@pkmn/client` `Battle` and rendering via `LogFormatter` + `Sprites/Icons`; run by `test:integration` (`package.json:44-48`), served by Parcel (`integration: start/build`).
- Inputs/outputs & data deps: Inputs: `integration/src/ui/teams/a.txt|b.txt`. Output: rendered browser battle + test results.
- Entry-points (UI/CLI/API): `npm run test:integration` (root), `integration/start` (parcel dev), `integration/index.html`.

### Monorepo build/publish tooling
- What it does: Compiles/bundles every `@pkmn/*` package and regenerates extracted code from Showdown submodules.
- Where it lives (`path` refs): `build:1-179`, `import`, `publish`, root `package.json:40-49`.
- How it works: `build` runs per-package `tsc` dual CJS/ESM (via generated `tsconfig.cjs/esm.json`, `mkconfig` in `build:33-41`) plus `tsup` bundle and optional `esbuild` IIFE appending to `window.pkmn.*` (`preamble` in `build:15`, `rewrite` in `build:18-29`); `dex` additionally bundles `learnsets.json` separately (`learnsets` const in `build:16`); `./import` syncs `vendor/pokemon-showdown*` submodules (currently empty — regeneration blocked without checkout) and regenerates `dex/data/*.json` + `sim/` + `streams/`; `./publish` releases each `@pkmn/*`.
- Inputs/outputs & data deps: Inputs: TS sources + `vendor/` submodules. Output: `build/cjs|esm` + browser IIFEs per package.
- Entry-points (UI/CLI/API): Root `npm run compile/build/test`, per-package `compile/build/test/lint`, `./build`, `./import`, `./publish`.

## Cross-feature notes
- Shared logic: `toID` is reimplemented in `dex/index.ts:14-18`, `data/index.ts:51-55`, `sets/src/sets.ts:41-44`, `view/src/common.ts:3-7`, `login/index.ts:159-163` — intentional zero-dependency duplication; `tr()` truncation helper duplicated in `data/index.ts:37` and inside `dex/index.ts:1331-1334`. `PokemonSet` flows `sets → sim/randoms → client(Battle ctor) → view(choices)`; `Protocol.Request` flows `protocol(parse) → client(update) → view(ChoiceBuilder)`; `Tracker` decouples `view(formatter)` from `client` but imposes formatter-before-handler ordering.
- Gaps vs other repos: This repo (`pkmn/ps`) has no damage calculator, no usage-stat/coverage tools, no team-builder UI — it is the protocol/data/client substrate those VGC tools build on. `Learnsets.isLegal` is explicitly teambuilder-only; authoritative legality, RNG, and mechanics live in generated `sim` (fix upstream, not here). `vendor/` emptiness means `sim`/dex JSON cannot be refreshed in this checkout.
- Reimplementation hints: For a calculator port, reuse don't fork: `@pkmn/types` brands + `Protocol.parse/key` for wire compat; `data.Generations` + `Dex.forGen/mod` for versioned lookups (never read `dex/data/*.json` directly); `Stats.calc` + `getHiddenPower` for stat/HP math; `Types.totalEffectiveness/canDamage` for chart math; `Sets/Teams.import/export/pack/unpack` for team I/O; `Battle.add/update` + `Pokemon.getDamageRange` for log-driven HP-range display; `ChoiceBuilder` for move/switch input; `Sprites/Icons` for artwork; `login` Actions for auth glue. Custom `Pokemon` subclasses plug in via `Side` provider; custom state plugs into `LogFormatter` via `Tracker`.
