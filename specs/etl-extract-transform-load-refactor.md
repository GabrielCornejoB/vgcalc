# Spec — ETL extract / transform / load refactor (all strategies)

Status: **READY TO IMPLEMENT** (2026-10-20). All decisions in §6 are closed.
The ability strategy pilot (§7) was **successful** and approved by the user;
the rest of the plan (§4) can start.

## Handoff (read this first)

Branch `feature/etl`. Nothing in §4 has been started yet; the only code change
so far is the ability pilot (§7), committed together with this spec.

- **Reference implementation:** `scripts/etl/internal/etl/ability-strategy.go`
  and `scripts/etl/internal/entity/ability.go`. Copy their shape for every
  other strategy (§2.1 conventions).
- **Next step:** §4 step 1 (rename read-side helpers), then follow §4 in order.
  Ability is already done, so skip it in step 3.
- **Before step 1**, take the table snapshots described in §5.
- **Conventions** (from `AGENTS.md` and the previous spec): conventional
  commits (`refactor(etl): …`), one commit per §4 item, commit trailer as in
  earlier commits, never push unless asked, say "calc (calc is short for
  calculatXXX btw)" instead of the full word. Follow
  `.agents/skills/golang-guidelines/SKILL.md` (120-char lines, index rules).
- Pure refactor: if any diff in §5 is not empty, stop and report it instead
  of "fixing" the data.

Files in scope: `scripts/etl/main.go`, `scripts/etl/internal/csv/`, every file
in `scripts/etl/internal/etl/`, and a new `scripts/etl/internal/entity/`.

## 1. Problem

The previous refactor (`etl-readability-refactor.md`) left every strategy as
`loadInputs → build* → apply*×N → upsert*`. The `apply*` passes fill rows
**column-wise**: each one loops over a secondary CSV and writes into
`rows[indexMap[id]]`. Consequences:

- To know what one `Pokemon` row contains you read 4 functions and follow an
  index map (`pokemonsMap`, `megasMap`, `movesPerEffect`).
- Every `apply*` mixes three jobs: locate the row, translate pokeapi id →
  vgcalc id, set the field.
- `pokemon-strategy.go` and `mega-evolution-strategy.go` load the same 5 CSVs
  and duplicate `apply*Stats` / `apply*Types` almost line by line.

## 2. Agreed design

| # | Decision |
|---|---|
| A1 | Every strategy's `Run()` is three steps: **extract → transform → load**. |
| A2 | **Extract** reads CSVs + DB id maps and returns lookups grouped by pokeapi id, with FKs already resolved to vgcalc ids (see D1 for the main-row case). |
| A3 | **Transform** builds each entity **row-wise** with a struct literal, then calls `Validate()`. No fluent builder, no index maps, no half-built rows. |
| A4 | **Load** is today's `upsert*` + `inTx`, renamed `load*`. Logic unchanged. |
| A5 | Entities live in a **separate package** `internal/entity` with exported types: types + `Validate()` only. No SQL, no CSV structs. |
| A6 | Pokemon and mega_evolution share one extract function, `extractPokemonSource`, in `internal/etl/pokemon-source.go`. Each strategy calls it and adds its own extras. Files are read twice at runtime (harmless); strategies stay independent; `EtlStrategy` and `main.go` are unchanged. |
| A7 | **Scope: all 9 strategies**, including the small ones, for consistency. |

### 2.1 Shape of every `Run()`

```go
func (s PokemonStrategy) Run() error {
	src, err := s.extract()
	if err != nil {
		return err
	}
	pokemon, err := transformPokemon(src)
	if err != nil {
		return err
	}
	return loadPokemon(s.DB, pokemon)
}
```

Conventions:
- `s.extract()` is a method (needs paths + DB). It returns an unexported
  `<entity>Source` struct.
- `transform<Entity>(src)` is a pure function returning `([]entity.X, error)`.
  It always returns an error, even when today's code cannot fail, because
  `Validate()` can.
- `load<Entity>(db, rows)` holds the query and the `inTx` loop.
- No doc comments on `<entity>Source` structs or `Validate()` methods; comment
  only non-obvious rules (skip rules, magic ids).

### 2.2 Naming collision with "load"

`load` now means "write to the DB". Two existing names use it for reading:

| Today | Renamed to |
|---|---|
| `csv.LoadFile` | `csv.ReadFile` |
| `loadPokeapiIDMap` | `queryPokeapiIDMap` |
| `loadTypes` (type-strategy, actually a transform) | `transformTypes` |

(Confirmed by D3.)

### 2.3 `internal/entity` package

```
internal/entity/
  shared.go          // Stats, TypeSlots, AbilitySlots
  type.go            // Type
  type-efficacy.go   // TypeEfficacy
  ability.go         // Ability
  nature.go          // Nature
  item.go            // Item
  move.go            // Move, MoveFlags
  pokemon.go         // Pokemon
  mega-evolution.go  // MegaEvolution
  pokemon-move.go    // PokemonMove
```

Shared value types (they replace the duplicated `apply*Stats/Types`):

```go
type Stats struct{ HP, Atk, Def, SpA, SpD, Spe int }
type TypeSlots struct {
	Primary   int  // vgcalc type id
	Secondary *int
}
type AbilitySlots struct {
	Slot1  int // vgcalc ability id
	Slot2  *int
	Hidden *int
}
```

`Pokemon` and `MegaEvolution` embed `Stats` and `TypeSlots`; `Pokemon` embeds
`AbilitySlots`. The column set and the `load*` SQL are unchanged; only the
field access changes (`p.Stats.HP`).

`Validate()` is exported, because the package boundary requires it. It returns a
`fmt.Errorf` naming the entity's slug and the broken rule. Its scope is set by D2.

### 2.4 `pokemon-source.go` (shared extract)

Replaces `pokemon-rows.go`. It holds the shared `pokemon*.csv` row structs
and:

```go
type pokemonSource struct {
	pokemonRows []PokemonCSVRow
	megaIDs     map[int]struct{}          // is_mega=1 forms + primals
	stats       map[int]entity.Stats
	types       map[int]entity.TypeSlots
	abilities   map[int]entity.AbilitySlots
}

func extractPokemonSource(db *sql.DB, dataDir string) (pokemonSource, error)
```

The stat id → field switch and the slot → field switches live here, once.
Unknown stat_id, slot or unresolvable type/ability id → error (same messages
as today's `apply*`).

**Behavior note:** today `apply*` only resolves pokemon that are kept, so
skipped rows (megas in pokemon, eternamax) are never checked. The shared
extract resolves **every** row. All types and abilities are seeded from
their full CSVs, so no new error is expected, but the verification run must
confirm it.

## 3. Per-strategy plan

"Source" = what `extract()` returns. FK resolution in the main row is subject to D1.

| Strategy | Source (extract) | Transform (row-wise) | Validate |
|---|---|---|---|
| **type** | `types.csv` rows | Title-case display name | — / D2 |
| **type_efficacy** | `type_efficacy.csv` rows + types id map | Resolve attack/target type, `factor/100` → multiplier. Moves the FK lookup **out of the tx loop** (today it is inside `inTx`). | multiplier ∈ {0, 0.5, 1, 2} |
| **ability** | `abilities.csv` rows; `names[id]` (EN); `prose[id]` (EN) | Literal from row + lookups | — / D2 |
| **nature** | `natures.csv` rows; `names[id]` (EN) | `statIDToSlug` for plus/minus (unknown → error) | plus/minus ∈ {atk, def, spa, spd, spe} |
| **item** | `items.csv` rows; `names[id]`; `prose[id]` | Duplicate-slug filter (keep first) + literal | fling_power ≥ 0 |
| **move** | `moves.csv` rows; types id map; `names[moveID]`; `effects[effectID]`; `flags[moveID]` → `entity.MoveFlags`; `hits[moveID]` | Category/target enum mapping, type resolve, literal from row + lookups. `movesPerEffect` disappears: transform reads `effects[row.EffectID]`. | category set; power, pp ≥ 0; accuracy 0–100; min/max hits ≥ 1 and max ≥ min |
| **pokemon** | `extractPokemonSource` + `fixedGenderBySpecies` (pokemon_species) | Skip megas and weight ≤ 0; display name; `-male`/`-female` suffix rule; literal embedding `stats[id]`, `types[id]`, `abilities[id]` | primary type ≠ 0, ability1 ≠ 0, all stats ≥ 1, weight > 0, fixed_gender in set |
| **mega_evolution** | `extractPokemonSource` + `itemIDByMega[id]` (form_conditions → forms → resolved item id, megas/primals only) + `basePokemonIDBySpecies` (species → resolved vgcalc pokemon id) + `dragonAscentMoveID` | Select rows with trigger or rayquaza-mega; skip if no `abilities[id]` (speculative megas); error if species has no base; error if `abilities[id]` has Slot2/Hidden (single ability rule); literal | Same stat/type/weight checks + exactly one of item/move (mirrors DB CHECK) |
| **pokemon_moves** | `pokemon_moves.csv` rows; pokemon + moves id maps | Filter VG 32 + train; unknown pokemon → skip (megas); unknown move → error | both ids ≠ 0 |

Constants and enum maps (`damageClassIDToSlug`, `moveTargetIDToSlug`,
`statIDToSlug`, rayquaza/dragon-ascent ids, champions VG ids) stay in
`internal/etl`. They are source-mapping rules, not entity rules.

Row structs used by one strategy stay in that strategy's file (unchanged
convention). Shared pokemon*.csv row structs move to `pokemon-source.go`.

## 4. Order of work (one commit each)

1. `refactor(etl): rename read-side load helpers` (§2.2).
2. `refactor(etl): add entity package` with shared value types only.
3. One commit per strategy, simple first to settle the pattern:
   type → type_efficacy → nature → item → move → pokemon_moves
   (ability is already done by the pilot, §7).
   Each one moves its entity into `internal/entity`, adds `Validate()` and
   reshapes `Run()` to extract/transform/load.
4. `refactor(etl): add shared pokemon source` + pokemon strategy.
5. mega_evolution strategy; delete `pokemon-rows.go`.
6. Update this spec's status/handoff.

## 5. Verification

Pure refactor: **no change to rows written.**

1. After every commit: `cd scripts/etl && gofmt -l . && go vet ./...`
2. **Before commit 1**, snapshot every table from the current code:
   `\copy (SELECT * FROM vgcalc.<table> ORDER BY <natural key>) TO '<file>'`
   for all 9 tables (docker DB `apps-db-1`, `.env` with `DB_*` + `DATA_DIR`).
3. After step 5: reset DB, `go run .`, re-export with the same queries and
   `diff`. Surrogate `id`s can differ only if insert order changes; compare by
   natural key (`pokeapi_id`, or the composite PK for junction tables).
4. Expected counts: `pokemon` 1251, `mega_evolution` 90, `pokemon_moves`
   14,752, `moves` 937, `abilities` 374.
5. Per-strategy check (the technique used in the pilot): after refactoring a
   strategy, blank its table's non-key columns (or `TRUNCATE` the junction
   tables `type_efficacy` / `pokemon_moves`, which nothing references), rerun
   `go run .`, re-export and `diff` against the snapshot. This proves the new
   code wrote every value instead of leaving old rows in place. Example
   (pilot): `UPDATE vgcalc.abilities SET slug = 'blank-' || id,
   display_name = '', short_effect = NULL, effect = NULL`.

## 6. Decisions

Closed by the user: **D1 = (b)**, **D2 = (a)**, **D3 = rename**, **D4 = as
proposed**. Under D2 (a), a table with no CHECK constraints or FKs (e.g.
abilities) validates only its identity columns: `pokeapi_id > 0`, `slug != ""`.

| # | Question | Options | Recommendation |
|---|---|---|---|
| D1 | A2 says extract resolves FKs. That fits the **grouped lookups** (stats/types/abilities). But some FKs sit on the **main row being transformed**: move.type_id, type_efficacy attack/target, pokemon_moves pokemon/move, mega base/item/move. Where do those resolve? | **(a) Extract, strictly:** extract returns small "resolved input" records per main row, so transform never sees an id map. Uniform, but extract for type_efficacy / pokemon_moves becomes most of the work and duplicates the CSV row shape. **(b) Split rule:** lookups built from *secondary* CSVs are resolved in extract; FK columns of the *main* CSV are resolved in transform via id maps carried in the source struct. | **(b)**. One sentence explains it; filters like "skip unknown pokemon" stay next to the resolution that triggers them. |
| D2 | What does `Validate()` check? | **(a)** Mirror DB `CHECK`/`NOT NULL` constraints and required FKs only (zero id = missing). **(b)** (a) + domain rules not in the DB (e.g. `display_name != ""`). | **(a)**. Keeps it a pure refactor; (b) may reject rows that load today (e.g. abilities without an EN name) and is a separate decision. |
| D3 | Rename read-side `load*` helpers (§2.2)? | Rename / keep. | **Rename**; otherwise "load" means two things. |
| D4 | Commit order (§4): simple strategies first, shared pokemon source after? | As proposed / pilot mega first. | **As proposed**; the pattern settles on small files before the hard one. |

## 7. Pilot: ability strategy (done, successful, approved)

Scope was limited to what ability needs; no shared helpers or other entities.

- New `internal/entity/ability.go`: `Ability` + `Validate()`.
- `internal/etl/ability-strategy.go`: `Run()` = `s.extract()` →
  `transformAbilities(src)` → `loadAbilities(db, abilities)`.
  `abilitySource` holds the rows plus `names` / `prose` maps (English only).
  `buildAbilities`, `applyAbilityNames` and `applyAbilityEffects` are gone.
- No doc comments on `abilitySource` or `Validate()` (removed at the user's
  request; now a §2.1 convention).
- `csv.LoadFile` was **not** renamed yet (D3): every strategy uses it, so the
  rename stays as step 1 of §4.

Verification: `gofmt` / `go vet` clean. The abilities table was snapshotted
from the old code, then all its columns were blanked in the local DB
(`slug`, `display_name`, `short_effect`, `effect`) and the ETL was rerun.
The new output is byte-identical to the snapshot (374 rows).
