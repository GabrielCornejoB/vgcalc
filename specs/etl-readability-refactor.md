# Spec — ETL readability refactor (pokemon_moves first, then all strategies)

Status: **Phase 1, S1, S2, S3, S4 implemented** (2026-10-18/20). S5, S7, S8, S9
pending approval. S6 revised: row structs used by a single strategy stay in
that strategy's file; `pokemon-rows.go` holds only shared ones. S10 dropped
(the golang-guidelines skill has no file naming rule).

## Handoff (read this first)

Branch `feature/etl`. Last pushed commit is `3ca058c`; later commits are local
until pushed. All 9 strategies are registered in `scripts/etl/main.go`.

**Done (one commit each)**
- Phase 1: `pokemon-moves-strategy.go` simplified (pure `buildPokemonMoves`,
  unknown move = error, no skip counters/dedupe).
- S1: pokemon, mega evolution, pokemon moves registered in `main.go`.
- S2: strategies return bare `err`; `main.go` prints `%T: err` once.
- S3: `tx.go` `inTx(db, fn)` replaces Begin/Rollback/Commit in all strategies.
- S4: `upsertMoves` / `upsertPokemon` / `upsertMegaEvolutions` extracted;
  `loadInputs()` + `pokemonInputs` / `megaEvolutionInputs` structs for the
  pokemon and mega strategies. `Run()` is now 22 lines (pokemon) and 29
  (mega); move is 46 (left as is: it is a linear build/apply pipeline).
- Row structs: single-use ones live in their own strategy file.

**How to verify a change** (no automated tests exist)
1. `cd scripts/etl && gofmt -l . && go vet ./...`
2. `go run .` (needs `scripts/etl/.env` with DB_* and `DATA_DIR`, and the
   docker DB from `apps/docker-compose.yaml` running as `apps-db-1`).
3. Expected counts: `pokemon` 1251, `mega_evolution` 90, `pokemon_moves`
   14,752, `moves` 937. A pure refactor must not change them.

**Remaining work, in suggested order**
- S5 stat-id constants (option A): named ids for the identical 1–6 switch in
  `applyPokemonStats` / `applyMegaStats`; put in `constants.go`.
- S7 naming of id maps / `slot1s`: only if done across all strategies.
- S8 (`isEnglish` helper) and S9 (fold mega slot-1 ability into build):
  recommended to skip.
- Not part of this refactor, deferred by the user: seed scope (filter
  `pokemon` by pokedex 36 or not), Showdown Champions-mod overlay for
  Champions-specific move changes (e.g. Dire Claw 30% vs 50%), stale rows on
  re-run (D4). Details in `docs/relational-model.md`.

**Conventions** (see `AGENTS.md`): conventional commits (`refactor(etl): …`),
commit trailer `Co-Authored-By` as in earlier commits, one commit per item,
never push without being asked, and say "calc (calc is short for
calculatXXX btw)" instead of the full word.

Files in scope: `scripts/etl/main.go` and every file in
`scripts/etl/internal/etl/`.

## 0. Ground rules

- Pure refactor: **no change to rows written** (except where a decision below
  says so). Verify with `gofmt -l .` and `go vet ./...`; DB comparison is a
  separate step run by the user.
- Match the existing pattern: `Run()` orchestrates, `build*` creates skeleton
  rows, `apply*` fills columns from secondary CSVs, pure functions take slices.
- Comments only where a rule is non-obvious (skip rules, magic ids).

---

## Phase 1 — `pokemon-moves-strategy.go`

### 1.1 Problems in the current file

1. `Run()` interleaves three jobs: resolving ids, skipping, and inserting.
   You have to read the loop body to learn which rows are dropped and why.
2. `skippedPokemon` / `skippedMove` are counted and never used (dead code,
   and a silent data-loss path).
3. A missing **move** is skipped silently, like a missing pokemon. They are
   not equivalent: pokemon are legitimately absent (mega/primal ids live in
   `mega_evolution`), but every move in `pokemon_moves.csv` should exist in
   `vgcalc.moves`, so a miss is a data bug.
4. `buildPokemonMovesPairs` has two filters and a dedupe, but the names
   (`pairs`, `seen`) don't say what is being deduplicated or why. The dedupe
   is also redundant for correctness (`ON CONFLICT DO NOTHING`); it only saves
   round trips.
5. Two loose `const` declarations with a prose comment above them; the
   comment explains the *data*, not the constants.
6. `pokemonMoveCSVRow` / `pokemonMovePair` sit after `Run()`, away from the
   other row structs (`pokemon-rows.go`).
7. Naming: `pokemonMovesChampionsVersionGroupID` is 35 chars; `pairs`
   vs `rows` vs `pair` is vague.

### 1.2 Target structure

```
const (
    // version_group_id 32 = "champions" (25 = "scarlet-violet").
    championsVersionGroupID = 32
    // pokemon_move_method_id 12 = "train", the only method in that group.
    trainMoveMethodID = 12
)

type PokemonMove struct { PokemonID, MoveID int }   // surrogate ids, ready to insert

pokemonMovesToLoad(rows []pokemonMoveCSVRow) []pokemonMoveCSVRow   // filter VG32 + train, dedupe
resolvePokemonMoves(rows, pokemonMap, moveMap) ([]PokemonMove, skipped int, err error)
Run(): load csv -> load 2 id maps -> resolve -> insert in one tx
```

- `Run()` stays short: load, resolve, insert. Insert loop only does
  `tx.Exec`, nothing else.
- `resolvePokemonMoves` is pure: unknown pokemon → counted in `skipped`
  (expected), unknown move → **error** (see D1).
- `Run()` prints/returns nothing new unless D2 is chosen.
- `pokemonMoveCSVRow` moves to `pokemon-rows.go` (with the other pokemon*.csv
  row structs).

### 1.3 Decisions needed

| # | Question | Recommendation |
|---|---|---|
| D1 | Unknown move id: error or skip? | **Error.** All moves are seeded from `moves.csv`; a miss means the move seed is out of sync. |
| D2 | Report the skipped-pokemon count? | **Yes**, one line `fmt.Printf` at the end of `Run()` (same style as `main.go` output). Silent skips make "why is X missing?" hard to answer. If you'd rather not print, drop the counter. |
| D3 | Keep the in-memory dedupe? | **Keep** but name it (`seen` → `emitted`) and comment *why* (VG32 has no dupes today; guards a re-export). Alternative: delete it and rely on `ON CONFLICT DO NOTHING`. |
| D4 | Stale rows on re-run (`DO NOTHING` never removes a pair that left the CSV)? | Out of scope for the refactor; flag only. Revisit when regulations are seeded. |

---

## Phase 2 — Suggestions applied to the other strategies

Each item is independent and can be accepted or dropped individually.

### S1. `main.go` registers the new strategies (functional, not cosmetic)

Currently only type, type-efficacy, ability, nature, item, move run. Add, in
dependency order: `NewPokemonStrategy`, `NewMegaEvolutionStrategy`,
`NewPokemonMovesStrategy`.

### S2. Single error-wrap site

Every `Run()` repeats `return fmt.Errorf("XStrategy: %w", err)` (~12× in
mega-evolution, ~7× in move). Proposal: `main.go` wraps once:
`fmt.Errorf("%T: %w", strategy, err)`, and strategies return `err` bare
(inner helpers keep their `funcName():` context).

Cost: error prefixes become `etl.PokemonStrategy:`. Transaction-level
messages (`error inserting …`) keep their local wording.

### S3. Transaction helper

All 9 strategies repeat Begin / defer Rollback / loop Exec / Commit (~15
lines). Proposal in a new `tx.go`:

```go
func inTx(db *sql.DB, fn func(tx *sql.Tx) error) error
```

Each `Run()` then becomes: `return inTx(s.DB, func(tx *sql.Tx) error { for … })`.
Insert-error wording (`error inserting mega %v`) stays inside the callback.

### S4. Split long `Run()` bodies

- `MegaEvolutionStrategy.Run` (~110 lines): extract `loadInputs()` (6 CSVs +
  5 id maps into a small struct) and `upsertMegas(tx, megas)` (query +
  loop).
- `PokemonStrategy.Run` (~95 lines): same split.
- `MoveStrategy.Run`: same split for the upsert.

The `const query` moves to the `upsert*` function so `Run()` reads as the
pipeline: load → build → apply… → store.

### S5. Stat id constants / shared stat applier

`applyPokemonStats` and `applyMegaStats` have identical `case 1..6`
switches; `nature-strategy.go` has `statIDToSlug` for 2..6. Proposal in
`constants.go`: named ids (`statHP … statSpe`). Two options:

- **A (light):** constants only; both switches use them.
- **B (heavier):** embed a `BattleStats{HP, Atk, Def, SpA, SpD, Spe}` struct
  in `Pokemon` and `MegaEvolution`, with one `setStat(*BattleStats, id, v)`.

Recommendation: **A**. B also needs a decision for the type/ability
appliers, which differ enough that sharing them hurts readability.

### S6. Consolidate row structs and constants

- `megaFormConditionCSVRow` → `pokemon-rows.go`.
- `rayquazaMegaPokeapiID`, `dragonAscentPokeapiID` stay in
  `mega-evolution-strategy.go` (they are strategy-specific), but grouped in
  one `const ( … )` block like the pokemon-moves ones.
- `constants.go` keeps only truly cross-strategy values
  (`pokeapiLanguageID`, stat ids from S5).

### S7. Naming consistency

- `xMap` for "pokeapi id → surrogate id" is overloaded (`typesMap` is
  pokeapi→surrogate, `pokemonsMap` is pokeapi→slice index). Proposal:
  `…IDByPokeapiID` for the former, `…IndexByPokeapiID` for the latter, or
  keep current names and add a one-line comment at `loadPokeapiIDMap` /
  `build*`. Recommendation: rename only if you want it everywhere;
  half-renamed is worse than the current state.
- `slot1s` → `slot1AbilityByPokemon`.
- `megaTriggers` recomputes `megaEvolutionPokemonIDs(formRows)` although
  `Run()` already has it → pass it in.
- `nature` / `ability` / `item` `apply*Names` repeat the
  `LocalLanguageID != pokeapiLanguageID` guard 7× with the same comment;
  keep as is unless S8 is accepted.

### S8. English-only filter (optional)

The "skip if language != 9 / skip if unknown id" preamble is repeated in
every `apply*Names` / `apply*Effects`. Optionally a tiny
`isEnglish(languageID int) bool`. Low value; **recommend skipping** to avoid
abstraction for a one-line check.

### S9. Duplicate slot-1 rule in mega-evolution (optional)

`slot1Abilities` and `applyMegaAbilities` both iterate `abilityRows` with the
slot-1 rule. Could fold ability assignment into `buildMegaEvolutions` using
the `slot1s` map and drop `applyMegaAbilities`. Trade-off: breaks the
uniform build → apply×3 shape. **Recommend keeping as is.**

### S10. Naming of files

Files are kebab-case (`pokemon-moves-strategy.go`). `.agents/skills/golang-guidelines`
may prefer snake_case. Check the skill first; rename (with `git mv`) only if
it says so.

---

## Proposed order of work

1. Phase 1 (pokemon_moves) — one commit: `refactor(etl): simplify pokemon moves strategy`.
2. S1 — `feat(etl): register pokemon, mega evolution and pokemon moves strategies`.
3. S3 + S2 — transaction helper and error wrap (touches every strategy; one
   commit each).
4. S4, S5, S6, S7 — one commit per item.
5. Verify after each commit: `gofmt -l .`, `go vet ./...`. Run against the DB
   only after step 4, comparing row counts with expected values (`mega_evolution` 90,
   `pokemon` 1251).
