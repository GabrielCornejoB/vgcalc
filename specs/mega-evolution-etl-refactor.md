# Spec — mega_evolution ETL refactor + scope change

Status: approved (all decisions confirmed with user). Date: 2026-10-17.
Files touched: `scripts/etl/internal/etl/mega-evolution-strategy.go`,
`scripts/etl/internal/etl/pokemon-strategy.go`,
`scripts/etl/internal/etl/pokemon-rows.go`,
`docs/relational-model.md`.

## 1. Context

`mega-evolution-strategy.go` works but is structurally sloppy compared to
`pokemon-strategy.go`:

- `buildMegaEvolutions` takes **10 arguments** and is ~160 lines: filtering,
  base resolution, trigger resolution, stats pivot, type/ability lookup and
  assembly all in one body.
- `statsSeen` (a parallel `[6]bool` map) + a `for i := 0; i < 6; i++`
  validation loop exist only to pre-validate "all 6 stats present" before
  insert — the same anti-pattern as the already-removed `validatePokemon`.
- `sort.Slice` at the end is dead code: the append loop iterates
  `pokemonRows` (CSV order = ascending id), output is already sorted.
- `pokemonByID` maps all ~1300 pokemon rows but is used once (IsDefault
  comparison in base selection).
- Zero comments on the deliberate skip rules, so intended scope is
  indistinguishable from bugs.

Scope change decided with this refactor:

- **Rayquaza-mega is added** (move-triggered, no stone).
- **Primal Kyogre / Groudon are added** (orb items, not stones).
- Speculative megas without a confirmed ability stay **skipped**.

## 2. Verified data facts (PokeAPI CSVs)

- 97 `is_mega=1` forms. 96 have a `pokemon_form_conditions` row with
  `trigger_item_id` (stone megas). **Rayquaza-mega (form 10181, pokemon
  10079) is the only `is_mega=1` form with no condition row** — it is
  move-triggered and needs `dragon-ascent` (move 620).
- Primals: forms **10179** (kyogre-primal, pokemon 10077, item **468** Blue
  Orb) and **10180** (groudon-primal, pokemon 10078, item **467** Red Orb).
  `is_mega=0`, `is_battle_only=1`, `form_identifier="primal"`.
- All three rows are data-complete: 6 stats, slot-1 types, single slot-1
  ability (189 Primordial Sea / 190 Desolate Land / 191 Delta Stream —
  matches the relational model), positive weights (430.0 / 999.7 / 392.0 KG).
- "Battle-only + item trigger" is **not** a sufficient inclusion rule:
  Ultra Necrozma (item 969), crowned Zacian/Zamazenta (items 1161/1162) and
  Terapagos-Stellar are also battle-only item-triggered forms. The
  `form_identifier == "primal"` discriminator excludes them.
- 9 speculative megas have a trigger item but **no slot-1 ability** in
  `pokemon_abilities.csv` (zygarde, heatran, darkrai, magearna ×2, zeraora,
  tatsugiri ×3). No confirmed data exists, so they are skipped.
- The new Legends Z-A mega wave (pokemon 10278+) with confirmed abilities is
  included automatically by the generic rules — no per-row handling.
- `pokemon_form_conditions.csv` also has `base_form_id` and `trigger_move_id`
  columns: `base_form_id` is populated for only 8 of 96 mega rows (unusable
  for base resolution — species-sharing stays), `trigger_move_id` is never
  set on a mega row.
- Diancie-mega is a real Gen 6 mega with full data — already included; it is
  not part of the speculative skip list.

## 3. Decisions

| # | Decision | Rationale |
|---|---|---|
| D1 | Split `buildMegaEvolutions` into helper builders + apply passes, mirroring `pokemon-strategy.go` | 10 args → 7; both strategies read identically |
| D2 | Delete `statsSeen` bookkeeping + the stat validation loop | DB `CHECK (hp >= 1)` … + wrapped insert error cover it; same precedent as removing `validatePokemon` |
| D3 | Delete `sort.Slice` (+ `sort` import) | Output already in `pokemon.csv` id order from the append loop |
| D4 | Primals identified by `is_battle_only=1 AND form_identifier="primal"` (+ item trigger) | Data-driven; excludes ultra/crowned/gmax/tera which share the battle-only + item-trigger shape |
| D5 | Rayquaza-mega via explicit constants `rayquazaMegaPokeapiID = 10079`, `dragonAscentPokeapiID = 620` | Documented manual overlay (relational model marks `required_move_id` "Manual — no PokeAPI row"); fails loudly if data changes, unlike a "sole condition-less mega" property rule |
| D6 | Speculative megas skipped by the generic "no slot-1 ability → skip" rule, with a comment | No confirmed data in source; no per-row exceptions |
| D7 | Comments added on every skip/inclusion edge case | Genuine edge cases; the "short comment permitted" rule applies |
| D8 | `docs/relational-model.md` stone-less exceptions section reframed: rayquaza + primals **included** in seed; legality/bans are a `regulation_pokemon` curation concern, not a seed concern | User-approved scope change |
| D9 | Base resolution stays species-sharing with `is_default` preference | `base_form_id` unusable (8/96 populated) |
| D10 | `vgcalc.pokemon` excludes only mega + primal ids (not all `is_battle_only=1` forms) | User decision: minimal exclusion. Primals must not exist in both tables; gmax/ultra/crowned/etc. knowingly stay in `pokemon` for now |

## 4. Target structure

```
pokemon-rows.go (shared):
  PokemonFormCSVRow  + FormIdentifier `csv:"form_identifier"`
                    + IsBattleOnly   `csv:"is_battle_only"`

pokemon-strategy.go:
  megaEvolutionPokemonIDs(rows []PokemonFormCSVRow) map[int]struct{}
      // is_mega=1 OR (is_battle_only=1 AND form_identifier=="primal")
      // replaces megaPokemonIDs; used by both strategies

mega-evolution-strategy.go:
  const rayquazaMegaPokeapiID   = 10079
  const dragonAscentPokeapiID   = 620

  megaTriggers(formRows, conditionRows)   map[int]int   // mon id -> item pokeapi id
  slot1Abilities(abilityRows)             map[int]int   // mon id -> ability pokeapi id
  basePokemonBySpecies(pokemonRows, megaEvolutionIDs)   // species -> base mon pokeapi id

  buildMegaEvolutions(pokemonRows, triggers, slot1s, bases,
                      pokemonMap, itemsMap, movesMap)   // 7 args, skeleton only
  applyMegaStats(statRows, megas, megasMap)
  applyMegaTypes(typeRows, megas, megasMap, typesMap)
  applyMegaAbilities(abilityRows, megas, megasMap, abilitiesMap)
```

## 5. Step-by-step

1. **`pokemon-rows.go`** — add `FormIdentifier` and `IsBattleOnly` to
   `PokemonFormCSVRow` (pokemon strategy parses but ignores them; csvutil
   tolerates extra fields).
2. **`pokemon-strategy.go`** — rename `megaPokemonIDs` →
   `megaEvolutionPokemonIDs`, extend the predicate to
   `is_mega=1 OR (is_battle_only=1 AND form_identifier="primal")`, comment
   the primal case. `buildPokemon` exclusion now drops the 2 primal rows
   from `vgcalc.pokemon`. The `weight <= 0` skip stays (eternamax is
   `is_battle_only=1` but is caught by weight; rules overlap harmlessly).
3. **`mega-evolution-strategy.go`** — add the two rayquaza constants with a
   comment referencing the manual-overlay decision.
4. **`megaTriggers`** — for forms in the `megaEvolutionPokemonIDs` set, map
   pokemon id → `trigger_item_id` from condition rows. Rayquaza never
   appears (no row). Comment the ultra/crowned exclusion rationale.
5. **`slot1Abilities`** — pokemon id → slot-1 ability pokeapi id.
6. **`basePokemonBySpecies`** — species → base pokeapi id, candidates
   exclude transformation rows, prefer `is_default=1`.
7. **`buildMegaEvolutions`** — include a row iff (in `triggers`) OR
   (`row.ID == rayquazaMegaPokeapiID`); skip if no slot-1 ability
   (comment: speculative megas); resolve base + item FKs with descriptive
   errors; for rayquaza resolve `required_move_id` via
   `movesMap[dragonAscentPokeapiID]` (error if missing), leave
   `required_item_id` NULL — the exactly-one CHECK stays satisfied.
   Emits skeleton rows: PokeapiID, BasePokemonID, Slug, RequiredItemID,
   RequiredMoveID, Weight. No stats/types/abilities yet.
8. **`applyMegaStats` / `applyMegaTypes` / `applyMegaAbilities`** — mirror
   the pokemon equivalents (switch on stat_id 1–6; slot 1/2 types with
   pokeapi→surrogate resolution; slot-1 ability only). Unknown id/slot →
   descriptive inline error, matching `applyPokemonTypes` precedent.
9. **Delete** `statsSeen` + validation loop, `sort.Slice` + `sort` import,
   `pokemonByID`.
10. **`Run()`** — orchestrate: load CSVs → load 5 ID maps (types, abilities,
    items, pokemon, moves) → build the 3 helper maps → `buildMegaEvolutions`
    → 3 applies → single-transaction upsert (query unchanged).
11. **`docs/relational-model.md`** — stone-less exceptions section: rayquaza
    and primals included in seed; note that Champions-illegal marking is a
    regulations-layer concern.
12. **Verify** — `gofmt -l .` and `go vet ./...` only. No `go run`, no DB
    writes (testing happens separately after spec implementation is agreed).

## 6. Expected output after implementation (for the later DB test)

- `vgcalc.mega_evolution`: **90 rows** = 87 stone megas (96 stone-triggered
  minus 9 ability-less) + rayquaza-mega + 2 primals.
- `vgcalc.pokemon`: **1252 rows** (1254 minus the 2 primal rows).
- `vgcalc.pokemon_moves`: no code change. Primal/rayquaza/mega pokemon ids
  no longer resolve in `vgcalc.pokemon`, so their VG32 rows skip —
  transformations inherit the base mon's list, per the model.
- Rayquaza row: `required_item_id` NULL, `required_move_id` → dragon-ascent.
  Primal rows: `required_item_id` → Red/Blue Orb, `required_move_id` NULL.
