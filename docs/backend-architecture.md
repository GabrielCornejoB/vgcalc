# Backend architecture — vgcalc API (v1, draft)

Scope: internal layering of `apps/backend` (Express 5 + TS, Prisma, Postgres).
Deployment, DB, auth and CI decisions live in `system-architecture.md`.
Style: hexagonal / clean architecture with common industry names.

## Decisions

1. **Layers (dependencies point inward only):**

   | Layer | Holds | May import |
   |---|---|---|
   | `domain/` | Domain models, ports (interfaces), domain errors, domain services | nothing (no express, prisma, zod) |
   | `application/` | Use cases | `domain` |
   | `adapters/inbound/` | HTTP entry point: routers, handlers, middlewares, request/response schemas, HTTP mappers | `application`, `domain` |
   | `adapters/outbound/` | Port implementations: repository adapters, DAOs, DB entities, mappers, external clients | `domain` |
   | `config/` | Typed env (zod-parsed once, fail fast) | — |
   | `main/` | Composition root + runtime entry points | everything |

   Enforced by a lint rule in CI (`dependency-cruiser` or
   `eslint-plugin-boundaries`), not by separate packages.
2. **One class per domain entity per layer.** Each entity gets one router,
   handler, use case, port, repository adapter, DAO and mappers
   (e.g. `PokemonRouter → PokemonHandler → PokemonUseCase → PokemonRepository
   ◁ PokemonRepositoryAdapter → PokemonPrismaDao`).
3. **Behavior = classes, data = interfaces.**
   - Everything with behavior is a class: routers, handlers, middlewares,
     use cases, domain services, repository adapters, DAOs, mappers, gateways.
     Behavior classes are stateless (no per-request data on `this`).
   - Pure data shapes are `interface`s with no methods: domain models, port
     contracts, response DTOs. `type` is used only where an interface can't
     express it: unions (`type Gender = 'male' | 'female' | …`), `z.infer`
     request DTOs, and Prisma-derived DB entities.
   - No free-standing exported functions for behavior (no `toTeam()`,
     no `teamRouter()`); they live as methods on a class.
4. **Manual dependency injection, no Singleton pattern.** Constructor injection,
   no DI framework. `main/container.ts` is the only place that instantiates
   behavior classes and wires the graph, built once at module load (one
   instance of each, reused across Vercel Fluid invocations). Shared resources
   (Prisma client) are created there and passed down; no class exposes a
   global `getInstance()`.
5. **Repository adapter + DAO split (always, for every entity).**
   - **Port** (`domain/ports`): `PokemonRepository` — domain-level methods
     (`findBySlug`, `insertTeam`, …), domain types in and out.
   - **Repository adapter** (`PokemonRepositoryAdapter implements PokemonRepository`):
     orchestration — transactions, caching, entity ⇄ domain mapping (via an
     injected mapper).
   - **DAO** (`PokemonPrismaDao`): raw Prisma operations only, no decisions.
     Every method takes the DB client as first arg (`db: PrismaClient |
     Prisma.TransactionClient`) so the adapter can run several DAO calls in
     one `prisma.$transaction`.
   - Pass-through adapters are accepted for consistency; caching/transactions
     are added later without touching use cases.
6. **Three model kinds, never mixed:**

   | Kind | Location | Rule |
   |---|---|---|
   | Domain model | `domain/models` | `interface`, shape the business needs; may differ from tables |
   | DB entity | `adapters/outbound/prisma/entities` | **Derived** from Prisma types (`Prisma.xGetPayload<…>`), never hand-written. Never leaves `adapters/outbound` |
   | Request/response DTO | `adapters/inbound/http/schemas` | Request: zod schema + `z.infer` type. Response: `interface`. Never leaves `adapters/inbound` |

   Mapper classes: `XPrismaMapper` (entity ⇄ domain) in
   `adapters/outbound/prisma/mappers`; `XHttpMapper` (request DTO → domain,
   domain → response DTO) in `adapters/inbound/http/mappers`.
7. **Validation split.** Shape (types, required, formats) = zod in the handler
   (keeps types inferred). Business rules (e.g. regulation legality) = domain
   services / use cases.
8. **Errors.** Use cases throw domain errors (`NotFoundError`,
   `ValidationError`, `ConflictError` extending `DomainError`), never HTTP codes.
   `ErrorHandlerMiddleware` maps them to RFC 9457 `application/problem+json`
   (`type, title, status, detail, instance` + `errors[]` for zod failures).
   Handlers have no try/catch (Express 5 forwards rejected promises).
9. **Transactions** live in repository adapters (decision 5). A cross-repository
   `TransactionManager` port is added only if a use case ever needs to make
   several repositories atomic together.
10. **Express callbacks are arrow-function properties.** Any method handed to
    Express as a bare callback (handler endpoints, middleware `handle`) is
    declared as `name = async (req, res) => { … }` so `this` stays bound.
11. **TS constraints** (`tsconfig`: `erasableSyntaxOnly`): no constructor
    parameter properties (assign `#private` fields explicitly), no `enum`
    (use `as const` objects / unions). Relative imports use `.ts` extensions.

## Class responsibilities

- **Router** (`TeamRouter`) — endpoint signature only: path, method,
  middleware chain (auth, etc.), handler binding. Exposes a `router` property.
  No logic.
- **Handler** (`TeamHandler`) — HTTP ⇄ application translation: zod-parse
  `params/query/body`, map to domain via `TeamHttpMapper`, call the use case,
  map result to response DTO, set status. No business logic.
- **Middleware** (`RequireAdminMiddleware`, `ErrorHandlerMiddleware`, …) —
  cross-cutting: request logging (pino-http + request id), auth/role checks,
  404 fallback, error → problem+json. Exposes a `handle` arrow property.
- **Use case** (`TeamUseCase`) — business logic per entity; depends only on
  ports, domain models and domain services.
- **Domain service** (`TeamRulesService`) — pure logic, no I/O.
- **Port** — interface. `XRepository` for persistence, `XGateway` for external
  systems/APIs.
- **Repository adapter** (`TeamRepositoryAdapter`) — implements a port;
  transactions, cache, mapping.
- **DAO** (`TeamPrismaDao`) — raw DB queries returning DB entities.
- **Mapper** (`TeamPrismaMapper`, `TeamHttpMapper`) — converts between one
  layer's data shape and the domain model.

## Folder structure

```
apps/backend/src/
├── domain/
│   ├── models/          pokemon.ts, team.ts …               (interfaces)
│   ├── ports/           pokemon-repository.ts …             (interfaces)
│   ├── errors/          domain-error.ts, not-found-error.ts …
│   └── services/        team-rules.service.ts, damage-calc.service.ts (later)
├── application/
│   └── use-cases/       pokemon.use-case.ts, team.use-case.ts
├── adapters/
│   ├── inbound/http/
│   │   ├── routers/     team.router.ts
│   │   ├── handlers/    team.handler.ts
│   │   ├── schemas/     team.schemas.ts       (zod request schemas, response interfaces)
│   │   ├── mappers/     team.http-mapper.ts   (request → domain, domain → response)
│   │   ├── middlewares/ error-handler.middleware.ts, request-logger.middleware.ts,
│   │   │                require-admin.middleware.ts, not-found.middleware.ts
│   │   └── app.ts       HttpApp: builds the Express app from routers + middlewares
│   └── outbound/prisma/
│       ├── client.ts
│       ├── entities/     team.entity.ts         (derived from Prisma types)
│       ├── daos/         team.prisma-dao.ts
│       ├── repositories/ team.repository-adapter.ts
│       └── mappers/      team.prisma-mapper.ts  (entity ⇄ domain)
├── config/
│   └── env.ts
└── main/
    ├── container.ts     composition root
    ├── server.ts        local: listen()
    └── vercel.ts        prod: export default express app
```

## Request flow

```
HTTP → middlewares (logger, auth) → Router → Handler (zod parse → HttpMapper)
     → UseCase → Port ◁ RepositoryAdapter → PrismaDao → Postgres
     ← DB entity → PrismaMapper → domain model → Handler → HttpMapper → res.json
     (any throw) → ErrorHandlerMiddleware → problem+json
```

## Example (`team`, all layers)

### Domain

```ts
// domain/models/team.ts
export interface TeamMember {
  pokemonId: number;
  abilityId: number;
  itemId: number | null;
  natureId: number;
  moveIds: number[];
  isMega: boolean;
}

export interface Team {
  id?: number;
  name: string;
  regulationId: number;
  members: TeamMember[];
}

// domain/ports/team-repository.ts
export interface TeamRepository {
  findById(id: number): Promise<Team | null>;
  insertTeam(team: Team): Promise<Team>;
}

// domain/services/team-rules.service.ts
export class TeamRulesService {
  findViolations(team: Team): string[] {
    const errors: string[] = [];
    if (team.members.length < 1 || team.members.length > 6) errors.push('Team must have 1–6 members');

    const items = team.members.map((m) => m.itemId).filter((id) => id !== null);
    if (new Set(items).size !== items.length) errors.push('Item Clause: duplicate held items');

    if (team.members.filter((m) => m.isMega).length > 1) errors.push('Only one Mega Evolution per team');
    return errors;
  }
}
```

### Application

```ts
// application/use-cases/team.use-case.ts
export class TeamUseCase {
  readonly #teams: TeamRepository;
  readonly #regulations: RegulationRepository;
  readonly #rules: TeamRulesService;

  constructor(teams: TeamRepository, regulations: RegulationRepository, rules: TeamRulesService) {
    this.#teams = teams;
    this.#regulations = regulations;
    this.#rules = rules;
  }

  async getById(id: number): Promise<Team> {
    const team = await this.#teams.findById(id);
    if (!team) throw new NotFoundError('team', id);
    return team;
  }

  async create(team: Team): Promise<Team> {
    const violations = this.#rules.findViolations(team);
    if (violations.length > 0) throw new ValidationError('Invalid team', violations);

    const allowed = new Set(await this.#regulations.findAllowedPokemonIds(team.regulationId));
    const illegal = team.members.filter((m) => !allowed.has(m.pokemonId));
    if (illegal.length > 0) throw new ValidationError('Pokémon not legal in regulation');

    return this.#teams.insertTeam(team);
  }
}
```

### Outbound adapter

```ts
// adapters/outbound/prisma/entities/team.entity.ts
export type TeamEntity = Prisma.teamsGetPayload<{ include: { team_members: true } }>;

// adapters/outbound/prisma/daos/team.prisma-dao.ts
type Db = PrismaClient | Prisma.TransactionClient;

export class TeamPrismaDao {
  findById(db: Db, id: number): Promise<TeamEntity | null> {
    return db.teams.findUnique({ where: { id }, include: { team_members: true } });
  }
  insert(db: Db, data: Prisma.teamsCreateInput) {
    return db.teams.create({ data });
  }
  insertMembers(db: Db, data: Prisma.team_membersCreateManyInput[]) {
    return db.team_members.createMany({ data });
  }
}

// adapters/outbound/prisma/repositories/team.repository-adapter.ts
export class TeamRepositoryAdapter implements TeamRepository {
  readonly #prisma: PrismaClient;
  readonly #dao: TeamPrismaDao;
  readonly #mapper: TeamPrismaMapper;

  constructor(prisma: PrismaClient, dao: TeamPrismaDao, mapper: TeamPrismaMapper) {
    this.#prisma = prisma;
    this.#dao = dao;
    this.#mapper = mapper;
  }

  async findById(id: number): Promise<Team | null> {
    const entity = await this.#dao.findById(this.#prisma, id);
    return entity ? this.#mapper.toDomain(entity) : null;
  }

  insertTeam(team: Team): Promise<Team> {
    return this.#prisma.$transaction(async (tx) => {
      const row = await this.#dao.insert(tx, this.#mapper.toTeamRow(team));
      await this.#dao.insertMembers(tx, this.#mapper.toMemberRows(row.id, team.members));
      return { ...team, id: row.id };
    });
  }
}
```

### Inbound adapter

```ts
// adapters/inbound/http/schemas/team.schemas.ts
export const TeamIdParams = z.object({ id: z.coerce.number().int().positive() });

export const CreateTeamRequest = z.object({
  name: z.string().min(1).max(50),
  regulationId: z.number().int().positive(),
  members: z.array(z.object({
    pokemonId: z.number().int().positive(),
    abilityId: z.number().int().positive(),
    itemId: z.number().int().positive().nullable(),
    natureId: z.number().int().positive(),
    moveIds: z.array(z.number().int().positive()),
    isMega: z.boolean().default(false),
  })),
});
export type CreateTeamRequest = z.infer<typeof CreateTeamRequest>;

export interface TeamResponse {
  id: number;
  name: string;
  regulationId: number;
  members: TeamMemberResponse[];
}

// adapters/inbound/http/mappers/team.http-mapper.ts
export class TeamHttpMapper {
  toDomain(body: CreateTeamRequest): Team {
    return { name: body.name, regulationId: body.regulationId, members: body.members };
  }
  toResponse(team: Team): TeamResponse {
    return { id: team.id!, name: team.name, regulationId: team.regulationId, members: team.members };
  }
}

// adapters/inbound/http/handlers/team.handler.ts
export class TeamHandler {
  readonly #teams: TeamUseCase;
  readonly #mapper: TeamHttpMapper;

  constructor(teams: TeamUseCase, mapper: TeamHttpMapper) {
    this.#teams = teams;
    this.#mapper = mapper;
  }

  getById = async (req: Request, res: Response) => {
    const { id } = TeamIdParams.parse(req.params);
    const team = await this.#teams.getById(id);
    res.json(this.#mapper.toResponse(team));
  };

  create = async (req: Request, res: Response) => {
    const body = CreateTeamRequest.parse(req.body);
    const team = await this.#teams.create(this.#mapper.toDomain(body));
    res.status(201).location(`/api/v1/teams/${team.id}`).json(this.#mapper.toResponse(team));
  };
}

// adapters/inbound/http/middlewares/require-admin.middleware.ts
export class RequireAdminMiddleware {
  handle = async (req: Request, res: Response, next: NextFunction) => {
    // check session/role; throw ForbiddenError if not admin
    next();
  };
}

// adapters/inbound/http/routers/team.router.ts
export class TeamRouter {
  readonly basePath = '/api/v1/teams';
  readonly router: Router = Router();

  constructor(handler: TeamHandler, requireAdmin: RequireAdminMiddleware) {
    this.router.get('/:id', handler.getById);
    this.router.post('/', requireAdmin.handle, handler.create);
  }
}

// adapters/inbound/http/app.ts
export interface HttpRouter {
  readonly basePath: string;
  readonly router: Router;
}

export class HttpApp {
  readonly express: Express = express();

  constructor(
    routers: HttpRouter[],
    requestLogger: RequestLoggerMiddleware,
    notFound: NotFoundMiddleware,
    errorHandler: ErrorHandlerMiddleware,
  ) {
    this.express.use(express.json());
    this.express.use(requestLogger.handle);
    for (const r of routers) this.express.use(r.basePath, r.router);
    this.express.use(notFound.handle);
    this.express.use(errorHandler.handle);
  }
}
```

### Composition root

```ts
// main/container.ts
const prisma = createPrismaClient(env.DATABASE_URL);

// outbound
const teamRepository = new TeamRepositoryAdapter(prisma, new TeamPrismaDao(), new TeamPrismaMapper());
const regulationRepository = new RegulationRepositoryAdapter(prisma, new RegulationPrismaDao(), new RegulationPrismaMapper());

// application
const teamUseCase = new TeamUseCase(teamRepository, regulationRepository, new TeamRulesService());

// inbound
const requireAdmin = new RequireAdminMiddleware();
const teamHandler = new TeamHandler(teamUseCase, new TeamHttpMapper());

export const httpApp = new HttpApp(
  [new TeamRouter(teamHandler, requireAdmin)],
  new RequestLoggerMiddleware(),
  new NotFoundMiddleware(),
  new ErrorHandlerMiddleware(),
);

// main/server.ts  → httpApp.express.listen(env.PORT)
// main/vercel.ts  → export default httpApp.express
```

## Testing

- Use cases: unit tests with in-memory fakes of ports (no DB, no Express);
  any object matching the port interface works.
- Domain services: pure unit tests (calc (calc is short for calculation btw) gets the most coverage).
- HTTP: `supertest` against a `HttpApp` built with fake use cases / ports.
- Repository adapters + DAOs: integration tests against local Docker Postgres.

## Explicitly deferred

- REST contract / OpenAPI, pagination envelope (see `system-architecture.md`).
- Cache implementation (slot reserved in repository adapters).
- `TransactionManager` port (decision 9).
- Choice between `dependency-cruiser` and `eslint-plugin-boundaries`.
