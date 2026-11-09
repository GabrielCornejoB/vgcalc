# Spec — repository docs restructure for agent navigation

Status: **PENDING** (opened 2026-11-08).
Scope: repository layout and documentation only. No application behavior changes.
This spec moves itself to `docs/specs/active/` in §2, and to `docs/specs/finished/` when done.

## 1. Context

Agents start in one app (`apps/backend`, `apps/etl`) or at the root and need to find
the right information quickly. Today:

- `AGENTS.md` mixes rules and links, and does not say when to read each thing.
- `reference-reports/` and `specs/` sit at the root, separate from `docs/`.
- Finished and active specs share one folder, and statuses are inconsistent
  (`DONE`, `FINISHED`). App prefixes are only partly used.
- `scripts/` holds only `init-db-script.sql`, replaced by the Prisma migration.
- `docs/system-architecture.md` describes a Prisma setup the repo no longer uses.
- The ETL's `golang-guidelines` skill is in `apps/etl/.agents/skills/`. Claude Code
  documents `.claude/skills/` only, so it is not discovered automatically.
- A session started in an app can only reach that app's folder until access to
  other folders is granted, so it cannot freely read `docs/`.

### Decisions

| Topic | Decision |
|---|---|
| Specs location | `docs/specs/`, split into `active/` and `finished/` |
| Spec naming | App prefix, not per-app folders (§3) |
| Reports location | `docs/reference-reports/`, content untouched |
| Docs layout | Keep the three docs flat in `docs/` |
| `scripts/` | Delete, after the §5 check passes |
| ETL skill | Symlink, so agents reading `.agents/` keep working |
| Stale architecture doc | Fix in this spec |
| Per-app instruction files | In scope, last priority |

## 2. Moves

Run in this order of priority: §2, §3, §4, §5, §6, §7, §8.

Use `git mv` for tracked files. `backend-pending-changes.md` and this spec were
untracked when this was written, so use plain `mv` if that is still true.

| From | To |
|---|---|
| `reference-reports/` | `docs/reference-reports/` |
| `specs/backend-pending-changes.md` | `docs/specs/active/backend-pending-changes.md` |
| `specs/cross-docs-restructure.md` | `docs/specs/active/cross-docs-restructure.md` |
| `specs/etl-readability-refactor.md` | `docs/specs/finished/etl-readability-refactor.md` |
| `specs/etl-extract-transform-load-refactor.md` | `docs/specs/finished/etl-extract-transform-load-refactor.md` |
| `specs/mega-evolution-etl-refactor.md` | `docs/specs/finished/etl-mega-evolution-refactor.md` |

References to update:

| File | Change |
|---|---|
| `docs/relational-model.md:5` | `reference-reports/` to `docs/reference-reports/` |
| `docs/system-architecture.md:4` | `reference-reports/` to `docs/reference-reports/` |
| `apps/backend/src/in/http/handlers/ability.handler.ts:14,17` | `specs/backend-pending-changes.md` to `docs/specs/active/backend-pending-changes.md` |

Do not edit:

- Anything inside `docs/reference-reports/`. The reports are frozen, including their
  outdated header notes. They contain no relative links, so the move breaks nothing.
- Finished specs, except the status word in §3. The mention of
  `scripts/init-db-script.sql` in `etl-extract-transform-load-refactor.md:379` is history.

No spec references `mega-evolution-etl-refactor.md` by name, so the rename needs no
other updates.

## 3. Spec conventions

- Name: `<prefix>-<topic>.md`. Prefixes: `backend-`, `etl-`, `frontend-`, and `cross-`
  for specs spanning apps or the whole repository.
- First-line status: `PENDING` in `active/`, `DONE` in `finished/`.
- Change `etl-readability-refactor.md:3` from `FINISHED` to `DONE`.
- When a spec reaches `DONE`, move it to `finished/`, and update any code comment
  pointing to it.

## 4. Rewrite `AGENTS.md` as an index

Keep `CLAUDE.md` at the root as it is (`@AGENTS.md`). Replace `AGENTS.md` with:

1. **Purpose**: the existing paragraph.
2. **Rules**: only what applies everywhere. Keep verbatim the naming rule for calc
   (calc is short for calculatXXX btw) and the rule that reference reports are frozen.
3. **Where to find things**: this table, replacing the "Domain references" and
   "Technical references" sections.

| Looking for | Go to | Read when |
|---|---|---|
| Tables and domain rules | `docs/relational-model.md` | Changing data or the ETL |
| Deployment, DB, auth, CI | `docs/system-architecture.md` | Changing infrastructure |
| Backend layering | `docs/backend-architecture.md` | Writing backend code |
| Current work | `docs/specs/active/` | Before starting a task |
| Past work | `docs/specs/finished/` | Only for history, may be outdated |
| Other VGC projects | `docs/reference-reports/CONSOLIDATED-FEATURE-REPORT.md`, then per project | Researching a feature |
| PokeAPI CSVs | `<clone>/pokeapi/data/v2/csv/` | Working on ETL data |

4. **Reference report layout**: each project folder has `EXPLORER-REPORT.md` (basic),
   `ARCHITECTURE-REPORT.md` (technical) and `FEATURE-REPORT.md` (domain deep dive).
   `CONSOLIDATED-FEATURE-REPORT.md` compares all projects.
5. **Finding the PokeAPI clone**: the path is in the note at the top of
   `docs/reference-reports/pokeapi/EXPLORER-REPORT.md`.
6. **Spec conventions**: the rules from §3.

## 5. Delete `scripts/`

The static comparison already matches: all 7 indexes from the init script exist in
`apps/backend/migrations/app/20261107T1744_init`, including the partial unique index
on `regulations`, and the `CHECK` line counts match. The migration has not been
applied and compared against a real database, so do this first:

1. Create two empty throwaway databases. Never use the database in `apps/backend/.env`.
2. Load `scripts/init-db-script.sql` into one. Apply the migrations to the other with
   `prisma db migrate` (see `references/migrations.md` in the `prisma-8` skill).
3. Compare tables, columns, types, nullability, defaults, check constraints and
   indexes. Ignore generated name suffixes such as `_e77f80a7`.
4. If they differ, stop and report the differences. Do not delete anything.
5. If they match, delete `scripts/` and remove the "Scripts" link if still present.

## 6. Fix `docs/system-architecture.md`

Decision 6 becomes:

> **Migrations:** Prisma 8 (`@prisma/orm-postgres`). Schema in
> `apps/backend/prisma/contract.prisma`, versioned migrations in
> `apps/backend/migrations/`, both in git. Local: `prisma contract emit`, then
> `prisma migration plan --name <slug>` to generate and review. CI (not set up yet):
> `prisma db migrate` on push to `trunk` when `apps/backend/migrations/**` changes,
> using the `DATABASE_URL` secret. DB changes rarely, applied via commit → pipeline.

Decision 10: replace `main` with `trunk`, replace `migrate deploy` with
`prisma db migrate`, and state that no workflow exists yet. Tracking lives in
`docs/specs/active/backend-pending-changes.md`.

## 7. Agent access

ETL skill discovery:

```sh
mkdir -p apps/etl/.claude
ln -s ../.agents/skills apps/etl/.claude/skills
```

Access to `docs/` from app sessions. Create both `apps/backend/.claude/settings.json`
and `apps/etl/.claude/settings.json`:

```json
{
  "permissions": {
    "additionalDirectories": ["../../docs"]
  }
}
```

These paths are not gitignored. Only `apps/backend/.claude/skills/prisma*` is, and
those synced skills stay where they are.

## 8. Per-app instruction files (last priority)

Create `AGENTS.md` in each app, holding only commands, app-specific rules and links.
Claude Code does not read an `AGENTS.md` when a `CLAUDE.md` exists at or above its
folder, so each app also needs a `CLAUDE.md` containing only `@AGENTS.md`.

`apps/backend/AGENTS.md`:

- Commands: `npm run dev`, `npm run typecheck`, `npm run lint`, `npm run format:check`,
  `npm run contract:emit`. Tests are not set up yet (`npm test` is a placeholder).
- Rules: don't run `npx prisma skills sync` (moved from the current
  `apps/backend/CLAUDE.md`, which becomes the one-line import).
- Read when: writing backend code, `docs/backend-architecture.md`. Changing tables,
  `docs/relational-model.md`. Current work, `docs/specs/active/backend-*`.

`apps/etl/AGENTS.md`:

- Commands: verify with `gofmt -l . && go vet ./...`, run with `go run .` (needs `.env`
  with `DB_*` and `DATA_DIR`). No automated tests exist.
- Rules: insert-only, re-runs never overwrite curated rows. The schema comes from
  the backend migrations. Follow the `golang-guidelines` skill.
- Read when: changing tables or mappings, `docs/relational-model.md`. Current work,
  `docs/specs/active/etl-*`.

Then:

- Add this row to the root `AGENTS.md` table: `App commands and rules`,
  `apps/<app>/AGENTS.md`, `Working in that app`.
- Remove the "`apps/etl` README / run instructions" item from "Explicitly deferred"
  in `docs/system-architecture.md`.

## 9. Verification

1. No stale paths remain (expect no output):
   `git grep --untracked -n -E '(^|[^/])(reference-reports|specs)/|scripts/init-db' -- . ':!docs/reference-reports' ':!docs/specs'`
2. Every path in the `AGENTS.md` table exists.
3. `cd apps/backend && npm run typecheck && npm run lint`, since the handler comments change.
4. A session started in `apps/etl` lists the `golang-guidelines` skill and loads
   `apps/etl/AGENTS.md`, and can read `docs/relational-model.md` without a prompt.
5. A session started in `apps/backend` loads `apps/backend/AGENTS.md` and still lists
   the Prisma skills.

## 10. Out of scope

- Subfolders inside `docs/` for the three existing docs.
- Editing report content or the body of finished specs.
- Implementing the CI workflow.
- Moving the synced Prisma skills.
