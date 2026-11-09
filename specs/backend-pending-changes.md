# Spec — backend pending changes (ability pagination follow-ups)

Status: **PENDING** (opened 2026-11-07, last updated 2026-11-08).
Origin: review of the `QUERY /abilities` paginated endpoint on `feature/prisma-init`.
Files involved: `apps/backend/src/in/http/handlers/ability.handler.ts`,
`apps/backend/src/in/http/routers/app.ts`,
`apps/backend/src/domain/models/ability.model.ts`,
`apps/backend/src/domain/ports/ability.repository.ts`,
`apps/backend/src/out/prisma/mappers/ability.mapper.ts`,
`apps/backend/src/out/prisma/repositories/ability.repository.ts`,
`apps/backend/src/main/server.ts`,
`apps/backend/package.json`.

Items §1 and §2 have a matching `TODO` comment in `ability.handler.ts`.
Items §3–§5 are minor and have no comment in the code.
Items §6 and §7 are tooling, and both feed the CI requirement at the end of this spec.

## Progress

| § | Item | State |
|---|---|---|
| 1 | Error handling (RFC 9457) | Designed, not implemented |
| 2 | Total count in the response | Not started |
| 3 | Domain `Ability` nullability | Not started |
| 4 | Rename `#prisma-migrate/*` to `#prisma/*` | Done |
| 5 | Graceful shutdown | Not started |
| 6 | Unit testing | Runner and coverage gate decided, not set up |
| 7 | Linting and format on save | Done |
| CI | GitHub Actions workflow | Not started, depends on §6 |

Already done from the original review: Node subpath imports replaced the
tsconfig `paths`, the missing `await` in the handler is fixed, and `page` is
0-based (`nonnegative()`). The endpoint stays on `QUERY`.

## 1. Error handling (TODO in code)

**State.** Designed, not implemented.

**Problem.** `PaginationSchema.parse(request.body)` throws a `ZodError` on
invalid input (`{"page": -1}`, `{"page": "1"}`, missing fields). Express 5
forwards the rejected promise to its default error handler, which answers
**HTTP 500 with an HTML page containing the full error and stack trace**.
Database errors from the repository end up the same way.

**Wanted.** A single central handler, like a WebFlux `@RestControllerAdvice`.
Every error leaves as RFC 9457 `application/problem+json`, no non-custom error
ever reaches a client, and handlers contain no try/catch.

**Design.**

- `ProblemDetails` interface in `src/in/http/errors`: `type`, `title`, `status`,
  optional `detail` and `instance`, plus extension members.
- Domain errors in `src/domain/errors` know nothing about HTTP: `DomainError`
  and subclasses such as `NotFoundError` and `ConflictError`.
- `HttpProblemError` in `src/in/http/errors` carries a full `ProblemDetails`, for
  problems that only exist at the HTTP edge, such as an unmatched route.
- `toProblemDetails(error: unknown, instance: string)` is the single mapping
  table, checked from most to least specific:

| Error | Result |
|---|---|
| `HttpProblemError` | Its own problem |
| `ZodError` | `400` with an `errors` extension of `{ path, message }` |
| Body parser `entity.parse.failed` | `400` |
| Body parser `entity.too.large` | `413` |
| `NotFoundError` | `404` |
| `ConflictError` | `409` |
| Other `DomainError` | `422` |
| Anything else | `500` with no `detail`, original error logged server-side only |

- Pipeline in `HttpApp`, in this order: `express.json()`, routers,
  `notFoundMiddleware`, `problemErrorMiddleware`.
- `problemErrorMiddleware` must declare all four parameters, because Express
  identifies error middleware by parameter count. It calls `next(error)` if
  `response.headersSent`, and wraps the mapper in try/catch with a hard-coded
  `500` problem so the mapper cannot leak Express's HTML handler.
- Handlers keep calling `parse` and let the error propagate.
- `package.json` `imports` gains `#errors` and `#http-errors`. Add `#http-errors`
  to the `src/domain` `no-restricted-imports` list in `.oxlintrc.json`.
- With `erasableSyntaxOnly`, fields are assigned in the constructor body.

**Not covered by the middleware.** Errors outside the request's promise chain
never reach it. Add `unhandledRejection` and `uncaughtException` handlers in
`server.ts` that log and exit (ties into §5). `typescript/no-floating-promises`
catches unawaited calls, but not a promise nested in an argument such as
`json({ data: promise })`.

**Replaces** the first TODO in `ability.handler.ts`.

## 2. Total count in the paginated response (TODO in code)

**Problem.** The response echoes the request's `pagination` back
(`{ data, pagination: { page, size } }`), so a client cannot tell how many
pages exist or whether a next page exists.

**Wanted.** The response carries at least `total`, and optionally the derived
`totalPages` / `hasNext`.

**Approach.**

- Count with the ORM aggregate:
  `db.orm.vgcalc.Ability.aggregate((a) => ({ total: a.count() }))`
  (`count()` is a `number`, never `null`).
- The port must change shape: `findByPagination` returns a paginated result
  (e.g. `{ items: Ability[]; total: number }`) instead of `Ability[]`.
  Define that as a generic domain model next to `Pagination` (e.g.
  `Paginated<T>`) so every future list endpoint reuses it.
- The page query and the count can run in parallel (`Promise.all`); they do
  not need a transaction for a read-only listing.
- The handler derives `totalPages` / `hasNext` from `total`, `page`, `size`.

## 3. Domain `Ability` nullability

`apps/backend/src/domain/models/ability.model.ts` disagrees with the contract
(`prisma/contract.prisma`, model `Ability`):

| Field | Contract | Domain today | Proposed |
|---|---|---|---|
| `pokeapiId` | `Int` (required) | `number \| null` | `number` |
| `shortEffect` | `String?` | `string` | `string \| null` |
| `effect` | `String?` | `string` | `string \| null` |

The mapper currently turns missing effects into `""` with `??`, which hides
"no effect text" from the domain and the API consumers. With the proposed types
the mapper copies the values as they are and the `??` goes away.

## 4. Rename the `#prisma-migrate/*` import alias

**State.** Done (`1dfb9ba`). Renamed to `#prisma/*` in `package.json` `imports`,
every import site, and the `no-restricted-imports` list in `.oxlintrc.json`.
There is a single `#prisma/*` key mapping to `./prisma/*.ts`, so the emitted
types are imported as `#prisma/contract.d`.

Verified 2026-11-08: `typecheck`, `lint` and `format:check` exit with code 0, the
domain import restriction still rejects `#prisma/db`, and a real
`QUERY /abilities` against the running app returns `200` with data. The backend
has no build step (`noEmit`, Node strips the types), so `typecheck` is the
compile check.

The alias points at the runtime client (`prisma/db.ts`) and the emitted
contract types (`prisma/contract.d.ts`), not at migrations. The `#` prefix
cannot clash with the `@prisma/*` packages, which is what the old name avoided.

## 5. Graceful shutdown

`src/main/server.ts` never closes the HTTP server or the Prisma client
(`db.close()`). Add `SIGTERM` / `SIGINT` handlers that stop accepting
connections, wait for in-flight requests, then call `await db.close()`.
Harmless locally; needed once the backend runs under a process manager or in
a container.

## 6. Unit testing setup and tests for existing code

**Problem.** `apps/backend` has no test runner: the `test` script in
`package.json` is still the `npm init` placeholder.

**Wanted.**

- A unit testing setup that works with the backend's stack: native ESM,
  `#` subpath imports, TypeScript type stripping.
- Unit tests for the existing code.
- A coverage report that CI can check.

**Decision.** Use Node's built-in test runner (`node:test` + `node:assert`), with
no test dependencies. Use cases receive their ports through the constructor,
so tests pass in fakes and need no module mocking.

**Coverage gate (verified on Node 22.16 against a copy of this backend).**
A threshold miss prints `Error: 60.56% line coverage does not meet threshold`
and exits with code 1, even when every test passes, so it blocks CI.
`--test-coverage-lines`, `--test-coverage-branches` and
`--test-coverage-functions` are available.

```
node --experimental-strip-types --test --experimental-test-coverage \
  --test-coverage-include="src/**/*.ts" \
  --test-coverage-exclude="src/**/*.test.ts" \
  --test-coverage-lines=<N> --test-coverage-branches=<N> --test-coverage-functions=<N> \
  "src/**/*.test.ts"
```

**Constraints to handle during setup.**

- Files that no test imports are left out of the report, and
  `--test-coverage-include` does not change that: a project with one tested
  file reports 100%. Add a test that dynamically imports every source file
  except `src/main/` (importing `server.ts` would start the server). Verified:
  untested files then count, and the same copy dropped from 100% to 60.56%.
- Barrels and type-only files always report 100% and inflate the total.
  Exclude them with `--test-coverage-exclude`.
- `--experimental-test-coverage` is still an experimental flag on Node 22.
- The runner does not type-check, so CI also runs `npm run typecheck`.

**Still to define.** The threshold values, and replacing the placeholder `test`
script in `package.json`.

## 7. Linting and format on save

**State.** Done and committed in the backend. Verified 2026-11-08: `lint`,
`format:check` and `typecheck` all exit with code 0, and the type-aware rules
fire on a deliberately broken file.

**Decision.** Oxlint with type-aware rules (`oxlint-tsgolint`) for linting,
because `typescript-eslint` only accepts `typescript <6.1.0` and the backend is
on 7.0.2. Prettier for formatting.

**Done.**

- [x] Format on save in Neovim (2026-11-08, uncommitted in the dotfiles repo).
  conform.nvim runs Prettier for JS, TS and HTML only, and only in projects with
  a Prettier config.
- [x] Oxlint and Prettier installed in `apps/backend`, with `.oxlintrc.json`,
  `.prettierrc` and `.prettierignore`.
- [x] Scripts: `lint` (`oxlint`), `format` (`prettier --write .`) and
  `format:check` (`prettier --check .`). `format:check` changes nothing and exits
  with code 1 on unformatted files, so it can block CI.
- [x] Prettier options: 4 spaces, single quotes, semicolons, `printWidth: 100`.
  Imports are sorted by `@ianvs/prettier-plugin-sort-imports`.
- [x] Oxlint plugins: `eslint`, `typescript`, `unicorn`, `oxc`, `import`.
- [x] Oxlint rules, all errors: the `correctness`, `suspicious` and `perf`
  categories, plus `import/no-cycle`, `typescript/no-floating-promises`,
  `typescript/require-await`, `typescript/prefer-nullish-coalescing`,
  `typescript/no-explicit-any`, and `no-restricted-imports` for the `src/domain`
  boundary. `import/no-unassigned-import` allows `dotenv/config`.

- [x] Fixed the 3 lint errors in existing code (`e5aa940`): the named `zod`
  import in `pagination.schema.ts` and `??` in `ability.mapper.ts`.
- [x] Formatted the codebase (`7c08fad`).

**Keep in mind.** `.prettierignore` must keep excluding `migrations/`,
`prisma/contract.json` and `prisma/contract.d.ts`. Migrations are hash-checked by
Prisma, so reformatting them risks `MIGRATION.HASH_MISMATCH`. `*.json` is also
ignored, so Prettier does not check `package.json` or `.oxlintrc.json`.

## CI requirement for §6 and §7

**State.** Not started. §7 is done, so it only needs §6 finished first.

The future GitHub Actions workflow must block PR approval unless unit test
coverage meets its thresholds and the project passes linting, the formatting
check and `npm run typecheck`. Each must be a required status check in the
branch protection rules for `trunk`; a failing workflow alone does not block
a merge.
