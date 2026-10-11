> **Reference note (moved 2026-10-07):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/engine`. Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Feature Report — engine

> Sources: ARCHITECTURE-REPORT.md (inventory), EXPLORER-REPORT.md, direct reads. Arch report was present; inventory in §"Feature/System Inventory" used as discovery list. All claims grounded in files read during this pass.

## Feature inventory
| Feature | One-line purpose | Key files |
|---|---|---|
| Gen1 battle simulation | Cartridge-accurate RBY link-battle state + mechanics | `src/lib/gen1/data.zig`, `src/lib/gen1/mechanics.zig`, `src/lib/gen1/helpers.zig` |
| Gen2 battle simulation | Parallel GSC engine with Field/Weather/Items/Gender | `src/lib/gen2/data.zig`, `src/lib/gen2/mechanics.zig`, `src/lib/gen2/helpers.zig` |
| Turn engine | turn-0 start → selectMove → turnOrder → doTurn → endTurn | `src/lib/gen1/mechanics.zig:52-78` |
| Choice enumeration | Legal move/switch listing incl. soft-lock semantics | `src/lib/gen1/mechanics.zig:3134-3213`, `src/lib/gen1/data.zig:115-117` |
| Chance tracking | RNG-event Actions/Durations + Rational probabilities | `src/lib/gen1/chance.zig`, `src/lib/gen2/chance.zig` |
| Damage calc overrides/summaries | Forced RNG outcomes + base/final/capped damage | `src/lib/gen1/calc.zig:37-134`, `src/lib/gen2/calc.zig` |
| Cartridge RNG (Gen12 LCG) | 9-byte link-battle PRNG | `src/lib/common/rng.zig:81-103` |
| Showdown RNG compat | Gen5/6-based PSRNG behind `-Dshowdown` | `src/lib/common/rng.zig:34-77` |
| Binary battle protocol logging | Native-endian FixedLog/Writer wire log | `src/lib/common/protocol.zig:312-881` |
| Showdown text-log translation | Binary → `@pkmn/protocol` text parser | `src/pkg/protocol.ts` |
| Zero-copy TS battle views | DataView Battle/Side/Pokemon lenses | `src/pkg/gen1.ts`, `src/pkg/index.ts` |
| ID Lookup translation | num↔Showdown-ID for species/moves/types/items | `src/pkg/data.ts:64-181` |
| Generated static data tables | Species/moves/types/items codegen from Dex | `src/tools/generate.ts`, `src/lib/gen1/data/*.zig` |
| Layout/offset metadata | SIZES/OFFSETS ABI JSON | `src/pkg/data/layout.json`, `src/pkg/data.ts:29-36` |
| Node-API native addon | ArrayBuffer update/choices binding | `src/lib/bindings/node.zig`, `src/lib/node.zig` |
| WASM binding | Freestanding wasm32 update/choices | `src/lib/bindings/wasm.zig`, `src/lib/wasm.zig` |
| C ABI binding | Static/dynamic libpkmn | `src/lib/bindings/c.zig`, `src/lib/c.zig`, `src/include/pkmn.h` |
| Dual addon loader | node/wasm autoload + showdown-variant slots | `src/pkg/addon.ts`, `src/pkg/addon/node.ts`, `src/pkg/addon/wasm.ts`, `src/pkg/addon/browser.ts` |
| Battle factory (create/restore) | ArrayBuffer init + JSON round-trip | `src/pkg/index.ts:342-365`, `src/pkg/gen1.ts:117-160` |
| Debug UI + pkmn-debug CLI | HTML battle/log inspector | `src/tools/debug.ts`, `src/tools/display/`, `src/bin/pkmn-debug` |
| Interactive demo app | Example/runner harness | `src/tools/demo.ts`, `src/tools/demo.zig` |
| Protocol dump tool | Raw log dumper (markdown/protocol/layout) | `src/tools/dump.zig` |
| LLVM analyze tool | Copyhound codegen analysis (copies/sizes) | `src/tools/analyze.zig` |
| Transition visualizer | RNG/state search visualizer | `src/tools/transitions.zig`, `src/lib/gen1/calc.zig:183-413` |
| Serde round-trip tool | Binary/JSON serialization checker | `src/tools/serde.ts`, `src/tools/serde.zig` |
| Release tooling | Cross-compile version/publish helper | `src/tools/release.ts` |
| VSCode extension | Editor integration (pkmn-debug preview) | `src/tools/vscode/extension.ts` |
| Zig toolchain installer | Pinned-Zig download + addon rebuild | `src/bin/install-pkmn-engine` |
| Showdown parity integration tests | `@pkmn/sim` cross-check | `src/test/showdown/`, `src/test/showdown.ts`, `src/test/integration.test.ts` |
| Fuzz harness | Randomized battle differential tester | `src/test/fuzz.zig`, `src/test/fuzz.ts` |
| Benchmark harness | Engine throughput measurements | `src/test/benchmark.zig`, `src/test/benchmark.ts` |
| C example | libpkmn consumer sample | `examples/c/example.c` |
| JS/web example | Vite/Parcel browser+node sample | `examples/js/example.ts` |
| Zig example | Native consumer sample | `examples/zig/example.zig` |

## Feature details
### Gen1 battle simulation
- What it does: Complete RBY link-battle simulation: parties, active volatiles/boosts/status, turn counter, last-damage/last-move desync detectors, PRNG.
- Where it lives (`path` refs): `src/lib/gen1/data.zig:57-119` (Battle extern struct), `src/lib/gen1/data.zig:126-227` (Side/Active/Pokemon/MoveSlot), `src/lib/gen1/data.zig:237-386` (Status/Volatiles bitfields), `src/lib/gen1/data.zig:418-472` (Stats.calc/Boosts), `src/lib/gen1/mechanics.zig` (~3300 lines full move logic), `src/lib/gen1/helpers.zig:35-72` (init/fixed/random constructors).
- How it works: `Battle(RNG)` is `extern struct {sides:[2]Side, turn:u16, last_damage:u16, last_moves:[2]MoveDetails, rng:RNG}` with size assertion `384` (`src/lib/gen1/data.zig:121-123`). `Side` is 184 bytes with `order:[6]u8` indirection instead of pointers; active Pokémon state (stats/species/types/boosts/volatiles/moves) is 32 bytes. `Stats.calc` (`src/lib/gen1/data.zig:427-434`) implements cartridge stat formula with DV/EV/level math. Mechanics implements every move effect, residual, switching, faint handling.
- Inputs/outputs & data deps: Inputs are two `Choice` bytes + `Options(log,chance,calc)`; output is `Result` byte. Depends on generated `data/{moves,species,types}.zig` tables and `common/rng`, `common/protocol`.
- Entry-points (UI/CLI/API): `Battle.update` / `Battle.choices` per gen; TS `Battle.create/update/choices` (`src/pkg/gen1.ts:51-57`); C `pkmn_gen1_battle_update/choices` (`src/include/pkmn.h:264-288`); fuzz/bench/showdown tests drive it.

### Gen2 battle simulation
- What it does: Parallel GSC engine mirroring gen1 structure but with its own layout and mechanics: weather/field, held items, genders, new volatiles/conditions, different stat/type system.
- Where it lives (`path` refs): `src/lib/gen2/data.zig:53-87` (Battle with `sides,rng,turn,field,padding`, size `512`), `src/lib/gen2/data.zig:93-103` (Field/Weather), `src/lib/gen2/data.zig:105-...` (Side with `Conditions{Spikes,Safeguard,LightScreen,Reflect}`), `src/lib/gen2/mechanics.zig`, `src/lib/gen2/helpers.zig`, `src/lib/gen2/chance.zig`, `src/lib/gen2/calc.zig`, `src/lib/gen2/generated.zig` (gen2-only codegen output), `src/lib/gen2/data/{items,moves,species,types}.zig`.
- How it works: Same `update/choices` shape as gen1 (`src/lib/gen2/data.zig:79-85`) but dispatches to gen2 `mechanics.update`. Shares only constants from gen1 (`src/lib/gen2/data.zig:7` imports gen1 data sparingly) per deliberate per-gen silo pattern. Extra enums (`Item`, `Gender`, `Weather`, TriAttack etc.) exported in `src/lib/pkmn.zig:155-200`.
- Inputs/outputs & data deps: Same Choice/Result ABI as gen1; extra data dep on `data/items.zig` and gen2 type chart (18 types incl. Dark/Steel/`???`).
- Entry-points: Zig `gen2.Battle.update/choices`; C header currently only exposes gen1 symbols (gen2 not yet in `pkmn.h` — only gen1 block at `src/include/pkmn.h:149-288`); TS factory only supports gen1 (`src/pkg/index.ts:347-350` `case 1`, throws otherwise).

### Turn engine (selectMove/turnOrder/doTurn/endTurn)
- What it does: Resolves one full turn from the two player choices, including turn-0 battle start, move selection validation, speed-order, sequential execution, and end-of-turn residuals.
- Where it lives (`path` refs): `src/lib/gen1/mechanics.zig:52-78` (`update`), `src/lib/gen1/mechanics.zig:80-96` (`start`), `src/lib/gen1/mechanics.zig:103-...` (`selectMove`), plus `turnOrder`, `doTurn`, `endTurn` later in same file.
- How it works: `update` asserts pass-choices only legal at turn 0 (`src/lib/gen1/mechanics.zig:53`); turn 0 calls `start` which finds first alive slot per side and `switchIn`s both. Otherwise calls `selectMove` per player (handles Recharging/Rage/Thrashing/Charging forced moves, PP, disable), computes `turnOrder` (speed + priority + RNG speed-tie via chance), runs `doTurn` in winner-first order (each half-turn can early-return a terminal `Result`), clears `Binding` when `attacks==0`, then `endTurn` (residual damage, sleep/confusion/disable/attacking/binding duration rolls, faint checks, turn increment + log).
- Inputs/outputs & data deps: In: `*Battle, c1:Choice, c2:Choice, options{log,chance,calc}`. Out: `!Result` (error only if log buffer overflows). Uses RNG, Chance, Calc, Log side-channels.
- Entry-points: Called by `Battle.update` (`src/lib/gen1/data.zig:101-103`) → FFI bindings → TS `battle.update(c1,c2)`.

### Choice enumeration
- What it does: Lists legal next choices for a player given the last `Result` request type (move vs switch vs pass), including cartridge quirks where no legal choice exists (soft-lock).
- Where it lives (`path` refs): `src/lib/gen1/data.zig:115-117` (doc contract), `src/lib/gen1/mechanics.zig:3134-3213+` (`pub fn choices`), `src/lib/gen1/mechanics.zig:2701-2704` (`isForced`: Recharging|Rage|Thrashing|Charging).
- How it works: `switch request`: `Pass→[pass]`; `Switch→` alive bench slots 2..6 via `side.order` (skips fainted/empty; emits single `pass` if none). `Move→` if `isForced` returns single `Move{data: showdown?1:0}` (cartridge gives no input; Showdown forces slot-1 move). Otherwise lists alive bench switches + move slots, with cartridge branch: if `Bide/Binding` or foe binding or self FRZ/SLP (non-showdown) returns single `Move{0}` without listing moves. Showdown branch disables/limits moves instead and always returns >0 choices; cartridge may return 0 when Transform+MirrorMove/Metronome PP error collides with Disable.
- Inputs/outputs & data deps: In: `*const Battle, player:Player, request:Choice.Type, out:[]Choice` (caller-allocated). Out: `u8` count; `out[0..n]` valid. No RNG/log interaction.
- Entry-points: Zig `battle.choices(...)`; Node `choices` binding (`src/lib/bindings/node.zig:92-117`); WASM `choices` (`src/lib/bindings/wasm.zig:34-45`); C `pkmn_gen1_battle_choices` (`src/include/pkmn.h:283-288`); TS `battle.choices(player,result)` (`src/pkg/gen1.ts:55-57`, `src/pkg/addon.ts:76-105` decodes bytes to `Choice[]`).

### Chance tracking
- What it does: Records which RNG events occurred during an `update` and accumulates their exact probability as a lazy `Rational`, enabling search/MCTS deduplication and full transition enumeration.
- Where it lives (`path` refs): `src/lib/gen1/chance.zig:30-115` (`Actions` + `matches/relax/matchesAny/eql`), `src/lib/gen1/chance.zig:174-271` (`Action` packed u64: damage/hit/crit/secondary/speed_tie/confused/paralyzed/duration/sleep/confusion/disable/attacking/binding/move_slot/multi_hit/pp/psywave/metronome), `src/lib/gen1/chance.zig:274-342` (`Durations/Duration`), `src/lib/gen1/chance.zig:350-782` (`Chance(Rational)` tracker + `commit/glitch/clearPending/switched/speedTie/hit/criticalHit/damage/.../sleep/confusion/disable/attacking/binding/psywave/metronome/pp`), `src/lib/common/rational.zig:14-123` (`Rational(T)` lazy-reduced accumulator).
- How it works: Mechanics calls `chance.hit/criticalHit/damage/...` at each RNG roll. Cartridge mode defers hit/crit/damage into `pending` and only `commit`s when relevance is known (immune/sub/miss make rolls no-ops) — see extended comment `src/lib/gen1/chance.zig:362-385`. Showdown mode updates probability eagerly. Duration observations (`sleep/confusion/disable/attacking/binding`) track continuing/ended with per-turn conditional probabilities (e.g. sleep `1/(8-n)`). `Actions.matches` compares shape (field set-ness) for frontier dedup; `relax` maps `overwritten` confusion → `continuing` under `-Doverwrite`.
- Inputs/outputs & data deps: In: live battle + RNG rolls. Out: `actions:Actions` (16B), `durations:Durations` (8B), `probability:Rational`. Gated by `-Dchance`; nulled via `chance.NULL` (`src/lib/gen1/chance.zig:1128-1246`).
- Entry-points: No direct TS API (chance only visible via C `pkmn_gen1_chance_options` at `src/include/pkmn.h:206-210,231-250` and Zig `transitions` tool); TS side does not surface chance.

### Damage calc overrides/summaries
- What it does: Dual-purpose calc side-channel: (a) force RNG outcomes by overriding Actions, (b) expose base/final/capped damage per player for external damage-calculator UIs and transition coalescing.
- Where it lives (`path` refs): `src/lib/gen1/calc.zig:37-84` (`Summaries/Summary.Damage{base,final,capped}`), `src/lib/gen1/calc.zig:88-134` (`Calc{overrides,summaries}` + `overridden/confusion/base/final/capped`), `src/lib/gen1/calc.zig:594-842` (`Rolls` range helpers incl. `damage`, `coalesce`, `metronome`, `psywave`, `moveSlot`, `multiHit`), `src/lib/gen1/calc.zig:183-413` (`transitions` exhaustive enumerator), `src/lib/gen1/calc.zig:415-497` (`update` verifier).
- How it works: `overridden(player, field)` returns `?Action-field-type` if the override slot is non-zero/non-None and `-Dcalc` on (`src/lib/gen1/calc.zig:94-107`); mechanics consults it before drawing RNG so tests/tools can force hit/miss/crit/damage/etc. `summaries.foe.damage.{base,final}` records raw damage; `capped` marks saturation (substitute break / faint — extra damage irrelevant unless recoil/residual could still matter). `Rolls.coalesce` closed-form (`src/lib/gen1/calc.zig:766-774`, credit Orion Taylor) maps a damage roll to max roll producing identical damage, letting `transitions` skip redundant re-simulations. `transitions` nests ~20 deep `for (Rolls.*)` loops to enumerate the full frontier and asserts probabilities sum to 1.
- Inputs/outputs & data deps: In: `Calc{overrides}` pre-update. Out: `summaries` post-update. Depends on `chance.Actions` shape.
- Entry-points: C `pkmn_gen1_calc_options` + getters (`src/include/pkmn.h:215-217,252-256`); Zig `transitions` binary; `calc.NULL` null object (`src/lib/gen1/calc.zig:136-165`).

### Cartridge RNG (Gen12 LCG)
- What it does: Bit-exact Game Boy link-battle RNG for gens 1–2.
- Where it lives (`path` refs): `src/lib/common/rng.zig:81-103` (`Gen12{seed:[9]u8,index:u8}`), `src/lib/common/rng.zig:8-32` (`PRNG(gen)` dispatch), `src/lib/gen1/helpers.zig:74-...` (`prng` seeding incl. link-battle constrained seed glitch).
- How it works: `next()` computes `5*seed[index]+1` (wrapping), stores back, advances `index=(index+1)%9`; 10 bytes total (`src/lib/common/rng.zig:85-87` size assert). `percent(p)` maps 0–100 to 0–255 threshold (`(p*0xFF)/100`). Rolls compared as `roll < threshold`. Test vector `src/lib/common/rng.zig:105-115` pins sequence for seed 1..9.
- Inputs/outputs & data deps: In: 9-byte seed + index. Out: `u8` per draw. TS `prng` getter rotates seed by index on read (`src/pkg/gen1.ts:96-104`) so logical state round-trips though raw bytes don't.
- Entry-points: Internal to mechanics; seeded via `Battle.init(seed,p1,p2)` and `encodePRNG` (`src/pkg/gen1.ts:744-757`).

### Showdown RNG compat
- What it does: Replicates Pokémon Showdown's RNG so `-Dshowdown` builds produce bit-identical battles to `@pkmn/sim`.
- Where it lives (`path` refs): `src/lib/common/rng.zig:34-77` (`PSRNG{src:Gen56}` + `range/chance/newSeed`), `src/lib/common/rng.zig:149-164` (`Gen56` LCRNG64 `0x5D588B656C078965*x+0x269EC3`), `src/lib/common/rng.zig:8-10` (`PRNG` returns PSRNG when `showdown`), `src/lib/common/rng.zig:181-231` (`FixedRNG` test injector).
- How it works: `PSRNG.next→Gen56.next` returns `seed>>32`; `range(T,from,to)` scales `u32` to range via `next()*(to-from)/2^32+from`; `chance` is `range(0,den)<num`. Seeds are 64-bit (vs 9+1 bytes cartridge) so battle layout diverges (padding 3 vs 5 in gen2 `src/lib/gen2/data.zig:54`). `FixedRNG(gen,len)` replays canned rolls and panics on exhaustion; used by unit/integration tests to force deterministic branches.
- Inputs/outputs & data deps: In: u64 seed. Out: u32 draws. Layout/ABI implications handled by `Lookup`/`LAYOUT` per-variant selection.
- Entry-points: Build flag `-Dshowdown` (artifact renamed `pkmn-showdown`, `build.zig:72`); TS `showdown:true` option + `addon.check` (`src/pkg/addon.ts:40-49`); `pkmn_psrng_*` C API (`src/include/pkmn.h:127-132`).

### Binary battle protocol logging
- What it does: Emits a compact native-endian binary event stream per `update` (moves, damage, status, boosts, activations, field/side/weather for gen2, win/tie/turn framing, plus `LastStill/LastMiss` modifiers).
- Where it lives (`path` refs): `src/lib/common/protocol.zig:16-110` (`ArgType` enum incl. gen1/gen2/gen3+ reservations), `src/lib/common/protocol.zig:112-311` (per-reason enums: Move/Cant/Heal/Damage/Status/CureStatus/Boost/Fail/Activate/Start/End/...), `src/lib/common/protocol.zig:312-822` (`Log(WriterT)` with ~30 methods), `src/lib/common/protocol.zig:828-861` (`FixedLog` + minimal `Writer{buffer,pos}` with `writeByte/writeAll/writeInt`), `src/lib/common/protocol.zig:863-881` (`NULL` null logger).
- How it works: Mechanics calls e.g. `log.move(.{src,mv,tgt})`, `log.damage`, `log.turn`, `log.win/tie`. Encoding is `ArgType byte + packed ID bytes + native-endian u16s (hp)` — see `|switch|` test `src/lib/common/protocol.zig:1375-1428` showing big vs little endian branches. `Writer` errors `NoSpaceLeft` if buffer < `MAX_LOGS` (180 gen1); bindings translate that into `Result.Error`. `format`/`expectLog` debug helpers pretty-print with optional red diff coloring.
- Inputs/outputs & data deps: In: battle state deltas. Out: bytes in caller-provided `LOGS_SIZE` buffer (256B rounded unless ReleaseSmall). Gated by `-Dlog`.
- Entry-points: Node/WASM shims construct `FixedLog` only when a log buffer arg is passed (`src/lib/bindings/node.zig:69-84`, `src/lib/bindings/wasm.zig:21-31`); TS allocates `log:DataView` iff `config.log` (`src/pkg/gen1.ts:42`) and exposes `battle.log`.

### Showdown text-log translation
- What it does: Decodes the binary wire log into `@pkmn/protocol` text lines (`|move|`, `|-damage|`, `|switch|`, etc.) for display and Showdown interop.
- Where it lives (`path` refs): `src/pkg/protocol.ts:50-100` (`Info/SideInfo/PokemonInfo` — nicknames/genders needed for rendering), `src/pkg/protocol.ts:105-153` (`Log.parse` generator), `src/pkg/protocol.ts:246-476` (`DECODERS` per ArgType), `src/pkg/protocol.ts:483-533` (`decodeIdent/decodeDetails/decodeHPStatus/decodeStatus/decodeTypes`).
- How it works: `Log.parse(data:DataView)` iterates bytes; `0x00` terminates; `LastMiss/LastStill` patch kwargs of the buffered `|move|` line; `Move` flushes the pending line group (move groups its sub-events). Each decoder reads IDs via `decodeIdentRaw` (`player=(byte>>3)==0?p1:p2, id=byte&7`), resolves species/move names through `Lookup` + `gen` Dex, formats HP as `hp/max` or `0 fnt`, status via bit test order slp→tox→par→frz→brn→psn (`src/pkg/protocol.ts:521-529`). Needs out-of-band `Info` because binary log carries only numbers (species num, level, hp) not nicknames.
- Inputs/outputs & data deps: In: `DataView` binary log + `Lookup` + `Info{name,team,gender}`. Out: `Iterable<ParsedLine{args,kwArgs}>`. Depends on `PROTOCOL` maps generated from Zig (`src/pkg/data/protocol.json`).
- Entry-points: `new Log(gen, lookup, info).parse(battle.log)` — used in JS example (`examples/js/example.ts:81-86`) and debug UI.

### Zero-copy TS battle views
- What it does: Provides ergonomic typed access to the raw Zig battle buffer without copying or string storage in the hot path.
- Where it lives (`path` refs): `src/pkg/gen1.ts:26-161` (`Battle`: DataView+Lookup+out/log buffers, `turn/lastDamage/prng/toJSON/create/restore`), `src/pkg/gen1.ts:163-742` (`Side/Pokemon/StoredPokemon/MoveSlot` lenses, volatiles/boosts/status bit-twiddling at `src/pkg/gen1.ts:352-508,633-676,775-790`), `src/pkg/common.ts:41-116` (`Choice/Result` encode/decode/parse/format), `src/pkg/index.ts:23-49` (`API` interface), `src/pkg/index.ts:58-262` (`Gen1.*` types).
- How it works: All getters do `data.getUint*/getInt*` at `OFFSETS` with `LE` flag; enums/IDs resolved via `Lookup` (Zig stores only `u8` nums). `update`/`choices` delegate to `addon` with raw `ArrayBuffer`s (`src/pkg/gen1.ts:51-57`). `toJSON` materializes `Data<Battle>` (sides/turn/lastDamage/prng) for restore/serde.
- Inputs/outputs & data deps: Depends on `LAYOUT.sizes/offsets` + `Lookup`; must match the Zig build variant (offsets differ showdown vs cartridge).
- Entry-points: `Battle.create/restore` factory, `battle.sides/side/foe/turn/lastDamage/prng/bytes/toJSON/update/choices`.

### ID Lookup translation
- What it does: Keeps strings out of the engine core by translating between engine `u8` nums and Showdown string IDs only on the TS side.
- Where it lives (`path` refs): `src/pkg/data.ts:64-181` (`Lookup` class), `src/pkg/data.ts:13-36` (`IDS/PROTOCOL/LAYOUT` loading + `p1/p2` offset derivation), `src/pkg/data.ts:7-11` (`LE` endian probe).
- How it works: Constructor builds `typesByNum/ByName` from `IDS` JSON, `species/moves` arrays indexed `num-1` by iterating the passed `Generation` (Dex), `itemsByNum/ByID` from `IDS[1].items` (with `flowermail→mail` alias at `src/pkg/data.ts:102-104`). `speciesByNum(num)=species[num-1]`, `speciesByID(id)=gen.species.get(id).num`, same for moves; items offset by +1 (0=None). Cached per gen in `LOOKUPS` (`src/pkg/data.ts:45,74-78`), supports only gens 1–2 (throws otherwise).
- Inputs/outputs & data deps: In: `@pkmn/data` Generation. Out: num↔ID/TypeName mappings. Source JSON: `src/pkg/data/ids.json`.
- Entry-points: `Lookup.get(gen)` used by `Battle.create/restore`, `Side/Pokemon` getters, `Log` parser.

### Generated static data tables
- What it does: Derives pruned, cache-friendly Zig data tables (moves/species/types/items + effect groupings) from the canonical `@pkmn/data`/`@pkmn/sim` Dex and pret disassembly mirrors, so the engine never hand-maintains stats.
- Where it lives (`path` refs): `src/tools/generate.ts` (entire 1300+ line codegen driver), Mustache templates in `src/lib/common/data/` (referenced `src/tools/generate.ts:18,199-207`), outputs `src/lib/gen1/data/*.zig`, `src/lib/gen2/data/*.zig`, `src/lib/gen2/generated.zig`, plus `src/pkg/data/{ids,layout,mapping}.json`.
- How it works: Fetches/caches `pret/pokered` + `pret/pokecrystal` `.asm` (`getOrUpdate`, `src/tools/generate.ts:262-294`), parses move effects with `HIGH_CRIT`/`NAMES`/`EFFECT_MAPPING` normalizations, groups effects into `onBegin/onEnd/isSpecial/isMulti` (`src/tools/generate.ts:135-160`), computes type charts from Dex effectiveness (`getTypeChart`), gender ratios, PP tables, Metronome exclusion lists, gen2 effect-command optimizer + `GROUPED`/`SNIPPETS`/`FNS` move-function synthesizer (`src/tools/generate.ts:384-772`). `npm run generate` regenerates everything; generated Zig carries `//! Code generated by tools/generate` header.
- Inputs/outputs & data deps: In: `@pkmn/data` Dex + pret asm via HTTPS (cached in `.cache/`). Out: Zig tables + TS JSON.
- Entry-points: `npm run generate` / `package.json:generate`; must run regenerate+rebuild together or ABI drifts.

### Layout/offset metadata
- What it does: Versioned ABI contract between Zig memory layout and TS DataView lenses.
- Where it lives (`path` refs): `src/pkg/data/layout.json` + `src/pkg/data.ts:29-36` (`LAYOUT` typed as `{sizes,offsets}[]`), `src/pkg/gen1.ts:13-17` (`SIZES/OFFSETS` + `spa/spd→spc` aliasing), `src/data/layout.json` shared copy.
- How it works: Codegen emits per-gen `sizes` (Battle/Side/... byte sizes) and `offsets` (field byte offsets incl. nested `Stats/Boosts/Volatiles/Battle`); runtime derives `Battle.p1/p2` (`sides` and `sides+sizeof(Side)` at `src/pkg/data.ts:33-36`). Every field access and `ArrayBuffer` allocation (`new ArrayBuffer(SIZES.Battle)`) flows from it. `LE` probe + native-endian protocol means buffers aren't portable across endianness.
- Inputs/outputs & data deps: Generated from Zig `extern struct` layouts; consumed by views, Lookup, protocol parser, debug/dump/serde tools, `*.bin` fixtures.
- Entry-points: Import-time constants; no function API.

### Node-API native addon
- What it does: Exposes `update`/`choices` to Node with zero-copy ArrayBuffer sharing and build-option introspection.
- Where it lives (`path` refs): `src/lib/bindings/node.zig:36-117` (`bind/update/choices`), `src/lib/node.zig` (entry shim), `src/pkg/addon/node.ts` (loader).
- How it works: `bind` publishes `{CHOICES_SIZE, LOGS_SIZE, update, choices}` per gen (`src/lib/bindings/node.zig:36-48`; array length 1 — gen1 only at `src/lib/bindings/node.zig:30-34`). `update` asserts `len==sizeof(Battle)`, bit-casts `Choice`/`Result` to `u8`, branches on log-arg type: `undefined/null→&gen.NULL`, else wraps the caller's buffer in `FixedLog{Writer}` with `chance.NULL/calc.NULL` (`src/lib/bindings/node.zig:69-84`). `choices` asserts sizes, reinterprets out-buffer as `[]Choice`, returns count (`src/lib/bindings/node.zig:92-117`). `options` object exposes `{showdown,log,chance,calc}` (`src/lib/bindings/node.zig:18-28`).
- Inputs/outputs & data deps: In: ArrayBuffers + u8s. Out: u8 result / u8 count. Requires exact buffer sizes or asserts trap.
- Entry-points: `require('build/lib/pkmn[.node]').engine`; `src/pkg/addon/node.ts:15-27` async `load`, `:58-64` sync `loadSync`.

### WASM binding
- What it does: Same `update`/`choices` surface compiled to freestanding `wasm32` for browsers/bundlers.
- Where it lives (`path` refs): `src/lib/bindings/wasm.zig:9-47` (`gen(num)` factory + `size` pow2 rounding), `src/lib/wasm.zig` (shim), `src/pkg/addon/wasm.ts`, `src/pkg/addon/browser.ts`.
- How it works: `callconv(.c)` exports taking `*Battle + Choice + ?[*]u8` log pointer; `CHOICES_SIZE/LOGS_SIZE` rounded up to pow2 via `ceilPowerOfTwo` (`src/lib/bindings/wasm.zig:49-51`). Null log pointer → `&g.NULL`, else wraps slice in FixedLog. `choices` takes raw `[*]u8` + asserts player/request ranges. TS `toBindings(wasm.exports)` adapts to the same `Binding` interface as Node.
- Inputs/outputs & data deps: Same as Node but pointer+len instead of ArrayBuffer handles; memory owned by the WASM instance.
- Entry-points: `initialize(showdown, 'wasm'|Module|Instance|Response)` (`src/pkg/addon.ts:27-38`); browser auto-path via `addon/browser.ts`.

### C ABI binding
- What it does: Stable `extern "C"` static/dynamic `libpkmn` for C/C++/other-language consumers, including full log/chance/calc options (unlike Node/WASM which pin chance/calc to NULL).
- Where it lives (`path` refs): `src/lib/bindings/c.zig` (OPTIONS consts + choice/result/psrng/rational/gen1 wrappers), `src/lib/c.zig:1-46` (re-export shims), `src/include/pkmn.h` (entire public header: `PKMN_OPAQUE`, `pkmn_options`, size consts, `pkmn_choice/result`, `pkmn_psrng/rational`, gen1 battle/options/chance/calc at `src/include/pkmn.h:149-288`).
- How it works: `PKMN_OPTIONS` exposes build flags; `PKMN_*_SIZE/MAX` consts mirror Zig (`384B battle`, `128B options`, `16B actions`, `8B durations/summaries`); `pkmn_gen1_battle_options_set(log,chance,calc)` initializes the opaque options struct (NULL resets); getters return pointers to probability/actions/durations/summaries; `pkmn_gen1_battle_update(battle,c1,c2,opts)` and `..._choices(battle,player,request,out,len)` mirror Zig semantics including soft-lock-0 and `Result.Error` on log overflow.
- Inputs/outputs & data deps: Plain C structs/uints; battle must be 384-byte aligned storage.
- Entry-points: `#include <pkmn.h>` + link `libpkmn`; sample in `examples/c/example.c`.

### Dual addon loader
- What it does: Resolves and caches the correct native artifact per `showdown` flag, supporting Node require, WASM instantiate (file/Module/Instance/streaming), and browser fetch paths.
- Where it lives (`path` refs): `src/pkg/addon.ts:16-110` (`ADDONS` 2-slot cache, `initialize/check/supports/update/choices/size`), `src/pkg/addon/node.ts:9-13` (`NODE/WASM` path constants under `build/lib/`), `src/pkg/addon/wasm.ts` (`toBindings` adapter), `src/pkg/addon/browser.ts` (fetch-based loader).
- How it works: `ADDONS[+showdown]` holds `Bindings<false|true>`; `addons(showdown)` lazy `loadSync`s on Node (`src/pkg/addon.ts:57-62`). `update` encodes `Choice→u8`, calls binding, decodes `Result` (`src/pkg/addon.ts:64-74`); `choices` maps `'pass/move/switch'→0/1/2` via `choice[0]` char test, calls binding, decodes `out[0..n]` (`src/pkg/addon.ts:76-105` — includes perf comment about unavoidable full decode). `check` throws rebuild hint (`npx install-pkmn-engine --options=...`) on variant mismatch; `supports` probes availability.
- Inputs/outputs & data deps: In: `showdown:boolean` + optional `Argument`. Out: cached `Bindings{options,bindings[]}`.
- Entry-points: `initialize(showdown, addon?)` (required in browser before use), implicit autoload on Node.

### Battle factory (create/restore)
- What it does: Constructs zero-initialized engine buffers from human-friendly teams, or rehydrates them from JSON snapshots (turn/lastDamage/prng/party order/volatiles/boosts/status).
- Where it lives (`path` refs): `src/pkg/index.ts:342-365` (`Battle.create/restore` switch — gen1 only), `src/pkg/index.ts:264-339` (`CreateOptions/RestoreOptions/PlayerOptions{name,team}`), `src/pkg/gen1.ts:117-160` (`create`: Side.init+encodePRNG; `restore`: Side.encode+last_moves packing+turn/last_damage/PRNG), `src/pkg/gen1.ts:744-757` (`encodePRNG`), `src/lib/gen1/helpers.zig` (Zig-side `Side.init/random`, `prng` seeding).
- How it works: `create` allocates `SIZES.Battle` ArrayBuffer, encodes each side's 6 `PokemonSet`s (species/level/stats/moves/PP/ivs→DVs) plus seed, returns `Battle` lens. `restore` additionally encodes `turn`, `lastDamage`, `last_moves` (packed nibble cartridge vs 2-byte showdown at `src/pkg/gen1.ts:143-148`), active volatiles/boosts/status, and PRNG bytes. `CreateOptions` requires `seed:number[]` (+`inert` skips native binding allocation for pure-view use); log:true requires player names.
- Inputs/outputs & data deps: In: `Generation` + options/JSON. Out: `Gen1.Battle` lens over ArrayBuffer. Depends on `Lookup` + `LAYOUT`.
- Entry-points: `Battle.create(gen, options)`, `Battle.restore(gen, json, options)` — the only supported construction path from JS.

### Debug UI + pkmn-debug CLI
- What it does: Renders a battle buffer + binary log as inspectable HTML (teams, state, Showdown-text log lines) for debugging desyncs and fixtures.
- Where it lives (`path` refs): `src/tools/debug.ts:38-...` (`render(gens,buffer,err?,seed?,json?)` — peeks showdown/gen header bytes), `src/tools/display/{index,util,dom,ui}.ts(x?)` (HTML/DOM builders, autocomplete, prune helpers), `src/bin/pkmn-debug` (npm bin → debug runner, `package.json:92-95`), `src/tools/vscode/extension.ts` (invokes `pkmn-debug` binary for editor preview).
- How it works: Reads raw frame file (header: showdown byte, gen byte, 6 reserved), builds `Lookup` + `Battle` view over `Data.view(buffer)`, renders sides/active/turn/PRNG plus parsed `Log` lines into a single HTML doc printed to stdout. `error()` helper strips ROOT/ZIG paths and trailing frames for readable CLI errors (`src/tools/debug.ts:31-36`).
- Inputs/outputs & data deps: In: binary frame buffer (+ optional seed/JSON mode). Out: HTML string on stdout.
- Entry-points: `npm run debug` / `pkmn-debug <file>` / VSCode custom editor.

### Interactive demo app
- What it does: End-to-end sample runner that builds the engine (`zig build -Dshowdown -Ddemo`), pulls Smogon usage/sets, prunes species/moves, and plays/simulates battles to showcase the public API.
- Where it lives (`path` refs): `src/tools/demo.ts` (TS driver: Smogon fetch, `overall()` usage ordering, `pruneSpecies/pruneMove`, battle loop), `src/tools/demo.zig` (Zig counterpart binary), `src/tools/display/` shared render helpers.
- How it works: Resolves `gen` from argv, rebuilds via `zig build`, fetches Smogon sets over a local-data shim (`URL=https://data.pkmn.cc/` mapped to `../smogon/data`, `src/tools/demo.ts:24-32`), picks first-set teams per species in usage order, then runs `Battle.create/update/choices` loop (mirrors `examples/js` play loop).
- Inputs/outputs & data deps: In: gen argv + Smogon local data. Out: demo battles/HTML. Depends on `@pkmn/smogon`, `@pkmn/dex`.
- Entry-points: `npm run demo` (`package.json` demo script); `build/tools/demo` binary.

### Protocol dump tool
- What it does: Prints machine-readable dumps of the protocol and memory layout for docs/tests/codegen verification.
- Where it lives (`path` refs): `src/tools/dump.zig:8-...` (`Tool{markdown,protocol,layout}` enum, `main` argv dispatch), registered as zig build step (`build.zig:252-262`).
- How it works: `markdown` iterates `protocol` struct decls to emit docs tables; `protocol` emits raw ArgType numeric mapping (feeds `src/pkg/data/protocol.json`); `layout` emits sizes/offsets (feeds `layout.json`). Stdout writer with 4KB buffer.
- Inputs/outputs & data deps: In: subcommand arg. Out: text/JSON on stdout consumed by TS package + docs (`docs/PROTOCOL.md`).
- Entry-points: `zig build dump -- <markdown|protocol|layout>`.

### LLVM analyze tool
- What it does: Codegen-level static analysis (copy counts, type sizes) forked from TigerBeetle's copyhound, used to keep the hot path allocation/copy-free.
- Where it lives (`path` refs): `src/tools/analyze.zig:1-30` (`Tool{copies:usize,sizes}` union, threshold default 8), `build.zig:424-457` `tool()` registration.
- How it works: `copies [threshold]` walks LLVM IR counting memcpy-like copies above threshold; `sizes` reports aggregate sizes. No battle semantics — operates on compiled artifacts.
- Inputs/outputs & data deps: In: build artifacts + optional threshold. Out: report on stderr/stdout.
- Entry-points: `zig build analyze -- <copies [n]|sizes>`.

### Transition visualizer
- What it does: Exhaustively enumerates and pretty-prints every reachable RNG/state transition from a position (with probabilities), used to verify chance/calc correctness.
- Where it lives (`path` refs): `src/tools/transitions.zig:15-...` (`main` asserts `calc&&chance`, parses gen+seed, builds fixed Zapdos-vs-Mew-style battle, calls `calc.transitions`), `src/lib/gen1/calc.zig:183-413` (`transitions` engine + `Stats{frontier,updates,seen,saved}` + `MAX_FRONTIER=90`), `src/lib/gen1/calc.zig:529-592` (TTY colored `debug/format` printers).
- How it works: Seeds a battle, runs one `update` to get the observed `Actions` shape, then nests `Rolls.*` range loops (speedTie/sleep/disable/attacking/confusion/confused/paralyzed/binding/hit/psywave/moveSlot/multiHit/secondary/crit/damage×2 + metronome) forcing each combination via `calc.overrides`, re-simulating and classifying outcomes as same-shape (coalesce damage ranges, accumulate `Rational(u256)` probability, assert sums to 1) vs new-frontier (append + dim print). `MAX_FRONTIER` guards explosion.
- Inputs/outputs & data deps: In: gen + seed. Out: colored transition tree + stats on stdout; errors on already-seen/mismatch/improper fraction.
- Entry-points: `zig build transitions -- <gen> [seed]`; also invoked from `calc.update` verifier when `transition=true` (`src/lib/gen1/calc.zig:453-459`).

### Serde round-trip tool
- What it does: Verifies binary↔JSON serialization fidelity by emitting raw battle bytes for TS-side comparison.
- Where it lives (`path` refs): `src/tools/serde.zig:7-30` (parses gen+seed, builds `gen1.Battle.random` or fixed `GEN1`, `writeStruct` native-endian to stdout), `src/tools/serde.ts:1-22` (spawns `zig build serde -- <gen>`, partitions stdout bytes into `0x..` lines for golden comparison).
- How it works: Zig side writes the `extern struct` verbatim (`writeStruct(battle, endian)`); TS side captures the buffer and pretty-prints hex rows. Round-trip via `Battle.toJSON/restore` must reproduce identical bytes (modulo documented PRNG-index rotation in `src/pkg/gen1.ts:96-104`).
- Inputs/outputs & data deps: In: gen (+optional seed). Out: raw bytes / hex listing.
- Entry-points: `zig build serde -- <gen> [seed]`; `npm run serde`-style via `serde.ts`.

### Release tooling
- What it does: Automates multi-target cross-compilation, versioning, and npm/Zig publishing.
- Where it lives (`path` refs): `src/tools/release.ts:10-...` (`minimist` argv `{prod,dryRun}`, `sh` logger/executor, `TARGETS[{label,triple,mcpu}]` starting `x86_64-windows-gnu/aarch64-windows-gnu` at `src/tools/release.ts:38-40`), `src/tools/generate.ts` DATA outputs consumed as release inputs.
- How it works: Logs each `(cd dir; cmd)` invocation, supports `--dryRun` print-only mode; iterates TARGETS calling `zig build -Dtarget=...` + npm pack/publish steps (full body beyond lines read follows same `sh` pattern).
- Inputs/outputs & data deps: In: CLI flags. Out: `build/` artifacts + published packages.
- Entry-points: `npm run release`.

### VSCode extension
- What it does: Opens `.pkmn`/frame files in a read-only custom editor showing the `pkmn-debug` HTML preview.
- Where it lives (`path` refs): `src/tools/vscode/extension.ts:7-40` (`PkmnDebugEditorProvider implements CustomReadonlyEditorProvider`, `preview(uri)` spawns `pkmn-debug [path]` and collects stdout/stderr, kills on cancellation).
- How it works: `openCustomDocument` awaits `preview()` → `{uri,preview,dispose}`; `resolveCustomEditor` sets `webview.options.enableScripts` + `html=preview`, surfacing errors via `showErrorMessage`. Separate subpackage with `@vscode/vsce` dev dep.
- Inputs/outputs & data deps: In: file URI. Out: webview HTML.
- Entry-points: `npm run extension` (vsce package); activation via `package.json` contributes in `src/tools/vscode/`.

### Zig toolchain installer
- What it does: Pinned, verified Zig download + `zig build` of the correct `-Dshowdown/-Dlog/...` addon so `require('pkmn.node')` resolves.
- Where it lives (`path` refs): `src/bin/install-pkmn-engine:13-18` (`ZIG_VERSION 0.16.0`, `INDEX=https://ziglang.org/download/index.json`), `:37-57` (`download` with SHA-256 check + stream pipeline), `:59+` (`unpack7` tar/7zip fallback), plus `sh`/`fetch` helpers and `postinstall` hook (`package.json:postinstall`).
- How it works: Queries ziglang index JSON, picks platform tarball, verifies SHA-256 while streaming to disk, unpacks (tar or downloaded 7zip), then `execFileSync('zig',['build',...])` with user `--options` string (e.g. `-Dshowdown -Dlog -Dchance -Dcalc` per `Makefile:11-12`) into `build/lib/pkmn[.node,.wasm]`. `addon/node.ts:error` message points here when artifacts are missing (`src/pkg/addon/node.ts:66-70`).
- Inputs/outputs & data deps: In: `--options` flags, `--silent`. Out: `build/lib/` artifacts + cached toolchain.
- Entry-points: `npx install-pkmn-engine [--options='...']` (also `postinstall`).

### Showdown parity integration tests
- What it does: Proves bit-parity with `@pkmn/sim` by replaying identical battles with controlled RNG rolls and diffing resulting states/logs.
- Where it lives (`path` refs): `src/test/showdown/gen1.test.ts:1-40` (gen1 roll constants: HIT/MISS/CRIT/NO_CRIT/MIN_DMG/MAX_DMG/TIE + SECONDARY/SLP/DISABLE/MIMIC/BIDE/PAR/FRZ/CFZ/THRASH keyed by `file:line:col` instrument points), `src/test/showdown/gen2.test.ts` (gen2 counterpart), `src/test/showdown.ts:1-50+` (`ROLLS`, `BattleQueue.insertChoice` monkey-patch, `createStartBattle/verify/formatFor` helpers), `src/test/integration.test.ts:1-13`, `src/test/fixtures/*.input.log/*.bin`.
- How it works: `FixedRNG`/instrumented `@pkmn/sim` PRNG returns MIN/MAX/ranged values at exact source locations (e.g. `data/mods/gen1/scripts.ts:436:42` hit) forcing both engines down the same branch; `verify` compares engine `Result` + battle bytes + parsed logs against sim output across thousands of generated matchups (moves/status/durations sampled per `Choices` helper).
- Inputs/outputs & data deps: In: generated teams/choices/seeds. Out: vitest pass/fail + fixture logs. Requires `-Dshowdown` twin build.
- Entry-points: `npm run test:integration` / `vitest src/test`.

### Fuzz harness
- What it does: Long-running randomized differential tester that plays random battles to completion, records per-turn frames (state/result/choices/logs), and re-verifies transitions/calc invariants, dumping replayable seeds on failure.
- Where it lives (`path` refs): `src/test/fuzz.zig:10-29` (`Frame{log,state,result,c1,c2,extra}`, `transitions=true` debug flag, `pkmn_options.internal=true`), `:32-...` (`main`: `gen duration seed?` args with `s/m/h/d` suffix parsing), `src/test/fuzz.ts` + `src/test/fuzz.test.ts` (TS driver/CI wrapper).
- How it works: Generates `Battle.random` teams via `helpers` (`cleric/block` options), picks random legal choices via `choices()`, applies `update`, appends `Frame`s to an arena list; on mismatch runs `calc.transitions` + `calc.update` override-replay checks (same as unit verifier) and persists `initial` bytes + frame log + seed for `pkmn-debug` replay. CI runs via `.github/workflows/fuzz.yml`.
- Inputs/outputs & data deps: In: `gen duration seed?`. Out: stdout progress + failure artifacts.
- Entry-points: `npm run fuzz` / `zig build fuzz -- <gen> <duration> [seed]`.

### Benchmark harness
- What it does: Measures raw engine throughput (updates/sec, turns/sec) for perf regression tracking.
- Where it lives (`path` refs): `src/test/benchmark.zig:6-50` (`main`: `gen battles[/warmup] seed?` args, `benchmark(io,gen,seed,battles,warmup)`), `src/test/benchmark.ts` (TS-side driver importing `Options` in `src/test/showdown/gen1.test.ts:4`), `src/test/benchmark.test.ts`, `src/test/stats.ts` (aggregation helpers).
- How it works: Seeds `PSRNG`, plays `battles` full random battles (optional `warmup/` prefix count excluded from timing), accumulates wall time + turn count with `choices` buffer sized `CHOICES_SIZE`; reports battles/sec and turns/sec. Both Zig-native and TS-addon paths measured (`benchmark.zig` vs `benchmark.ts`).
- Inputs/outputs & data deps: In: gen, N battles, seed. Out: timing summary on stdout.
- Entry-points: `npm run benchmark` / `zig build benchmark -- <gen> <battles> [seed]` / `npm run integration` aggregation.

### C example
- What it does: Minimal native consumer proving the C ABI: seed → random full-battle playout with log buffer.
- Where it lives (`path` refs): `examples/c/example.c:9-27` (`choose`: `pkmn_gen1_battle_choices` + `pkmn_psrng_next` uniform pick, asserts `n>0` with soft-lock comment), `:29-50+` (`main`: decimal seed argv, `pkmn_psrng_init`, `choices[PKMN_CHOICES_SIZE]` stack buffer, battle init + update loop — remainder follows same pattern as JS example).
- How it works: Statically allocates `pkmn_gen1_battle` (384B opaque) + choices/log buffers, loops `update` until terminal `Result`, prints turn count/outcome. Built via `examples/c/Makefile` (`make example`).
- Inputs/outputs & data deps: In: CLI seed. Out: stdout battle summary. Links `build/lib/libpkmn.*` + `pkmn.h`.
- Entry-points: `make example` → `examples/c/example`.

### JS/web example
- What it does: Canonical public-API usage sample working identically on Node and in browsers (Vite/Parcel), with logging + random playout.
- Where it lives (`path` refs): `examples/js/example.ts:10-32` (`Random` Mulberry32 — explicit note not to use `@pkmn/sim` PRNG), `:39-55` (packed `Team.unpack` Bulbasaur-line + Tauros-line teams), `:68-76` (options `{seed:[1,2,3,4], showdown:true, log:true}` + build-flag caveats), `:79-120` (`play()`: create→Log→random-choose→update/display loop→win/lose/tie message), `:122-130` (Node runs sync; browser awaits `initialize(showdown)` first).
- How it works: `Battle.create(gen,options)` → `new Log(gen,Lookup.get(gen),options)` → loop `update(c1,c2)` + `log.parse(battle.log!)` + `battle.choices(p,result)` random pick until `result.type` set. Documents that `log:true` needs both `-Dlog` build and `name` fields, and `showdown:true` needs `-Dshowdown` build.
- Inputs/outputs & data deps: In: none (hardcoded teams/seed). Out: console Showdown-protocol lines + `Battle won by ... after N turns`.
- Entry-points: `examples/js/index.html` (Vite) / Parcel target; `make example` js step.

### Zig example
- What it does: Native Zig consumer sample (no FFI) showing direct `pkmn.gen1` imports, battle init, and update loop.
- Where it lives (`path` refs): `examples/zig/example.zig` (mirrors C/JS playout using `helpers.Battle.init` + `FixedRNG`/`PRNG` + `choices/update`), wired via `Makefile: example` + `build.zig` example step.
- How it works: Imports `pkmn` module directly (`@import("pkmn")`), builds teams with `helpers.Pokemon` literals, drives the same `update→choices` hot loop as the engine's own tests, printing results via `std.debug.print`.
- Inputs/outputs & data deps: Same as C example but against Zig API (no ABI translation).
- Entry-points: `zig build` example step / `make example` zig step.

## Cross-feature notes
- Shared logic: `common/` kernel (`data.zig` Choice/Result/Player/ID packing `src/lib/common/data.zig:7-117`, `battle.zig` Options triple, `protocol.zig` ArgType+Log, `rng.zig` PRNG dispatch, `rational.zig` lazy Rational, `options.zig` build-flag resolution `src/lib/common/options.zig:19-34`) is imported by every gen — the blast radius for any ABI change (also breaks `LAYOUT` JSON, TS offsets, fixtures, C header sizes simultaneously).
- Compile-time injection + Null Object is the unifying pattern: `showdown/log/chance/calc` booleans select `PSRNG vs Gen12`, `FixedLog vs NULL`, `Chance vs chance.NULL`, `Calc vs calc.NULL`; MCTS playouts that skip side-channels pay zero bytes/branches, at the cost of a build matrix (`pkmn` × `pkmn-showdown` × flag sets) that `install-pkmn-engine` + `addon.check/supports` manage.
- Gaps vs other repos (VGC calculator context): no doubles/triples format logic, no team preview/validation, no damage-calc UI (only `base/final/capped` summaries as data feed), no AI/search/MCTS (only `Chance.matches/relax` + `Rolls`/`transitions` primitives to build search on top of), no streams/async (synchronous byte-buffer FFI only — driver must guarantee legal choices as the engine does no validation), gen2 TS/C surfaces incomplete (C header and `Battle.create` are gen1-only; gen2 usable from Zig/tools/tests only).
- Reimplementation hints: to port this engine's feature set, implement in order — (1) `Choice/Result/Player/ID` 1-byte packings + `update/choices` signatures, (2) `extern` POD battle layout with `order[]` indirection + Handles-not-pointers + integer-only move/species/type IDs, (3) Gen12 LCG + PSRNG dual RNG with `FixedRNG` injector, (4) `mechanics.update` turn pipeline with `selectMove/isForced/turnOrder/doTurn/endTurn`, (5) `choices` with forced-move repurposing + soft-lock-0 contract, (6) `FixedLog/Writer` binary protocol + `Log.parse` text decoder driven by out-of-band `Info`, (7) `Chance` pending-commit probability tracker over `Rational`, (8) `Calc` overrides + `base/final/capped` + `Rolls.coalesce` + `transitions` enumerator, (9) codegen (`generate.ts`) as source of truth for data tables + `LAYOUT/IDS/PROTOCOL` JSON, (10) three FFI shims (Node/WASM/C) over caller-allocated buffers + dual `showdown` slot loader + `create/restore` factory, (11) verification trio (Showdown parity with `file:line:col` roll keys, fuzz frame replay, bench throughput) before any perf work.
