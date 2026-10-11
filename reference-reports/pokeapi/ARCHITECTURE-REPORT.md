> **Reference note (moved 2026-10-11):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/pokeapi` (PokeAPI — Django read-only reference-data API). Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Architecture Report — pokeapi

## Overview
- Read-only reference-data API over a large relational Pokédex: Django 5.2 + DRF 3.16 serving `GET`-only JSON at `/api/v2/*`, backed by PostgreSQL, fronted by nginx (edge cache + rate-limit), with Redis + `django-cachalot` as ORM cache (`config/settings.py:105-154`, `Resources/nginx/nginx.conf:67-153`, `pyproject.toml:8-19`).
- Single Django app `pokemon_v2/` holds ~140 models, ~50 viewsets, ~200 serializers; `config/` is thin project shell (settings overlays + root URLconf) (`config/settings.py:129-140`, `config/urls.py:1-7`, `pokemon_v2/models.py:1-234`).
- Offline-first ETL: `data/v2/csv/*.csv` (Veekun fork) + `sprites`/`cries` submodules + remote sprite manifest are bulk-loaded into Postgres by `data/v2/build.py:build_all()`; no write API (`data/v2/build.py:2477-2503`, `Makefile:32-36`).
- Dual serving: REST from Django/gunicorn (`gunicorn.conf.py:1-6`, `config/wsgi.py:16`); GraphQL from external Hasura engine pointed directly at the same Postgres, configured under `graphql/v1beta|v1beta2/` (`docker-compose.yml:50-75`, `graphql/`).
- Schema-as-contract: `drf-spectacular` generates checked-in `openapi.yml` via `make openapi-generate` (`config/settings.py:153-154,175-280`, `Makefile:59-60`).

## Core systems (responsibility + key files)
- HTTP routing / resource registry — `PokeAPIRouter(DefaultRouter)` + 50× `router.register(...)` + custom `PokeAPIRootView`, `PokemonEncounterView`, `PokeapiMetaView`: `pokemon_v2/urls.py:24-106`, `config/urls.py:5-7`.
- Read API layer — `PokeapiCommonViewset = ListOrDetailSerialRelation + NameOrIdRetrieval + ReadOnlyModelViewSet` plus ~50 thin `*Resource` subclasses fixing `queryset/serializer_class/list_serializer_class`: `pokemon_v2/api.py:99-182,199-986`.
- Lookup / filter mixins — `NameOrIdRetrieval.get_object()` (id-vs-name regex, `name__iexact`, int-overflow 404) + `?q=` icontains filter; `ListOrDetailSerialRelation.get_serializer_class()` (list=summary, detail=full): `pokemon_v2/api.py:99-157`.
- Domain model — `PokeApiModel/PokeApiManager` base (pk-ordered default queryset) + ~60 abstract `Has*` FK mixins (`HasAbility`, `HasVersionGroup`, …) + `IsName/IsDescription/IsFlavorText` (`HasLanguage`+payload) composed into concrete tables: `pokemon_v2/models.py:242-919`.
- Presentation layer — `SummarySerializer` (`{name,url}` hyperlinked stub) vs `DetailSerializer` (nested full graph via `SerializerMethodField` + `itertools.groupby` aggregation): `pokemon_v2/serializers.py:277-589,2932-3184,1905-2057,3465-3560`.
- ETL / build pipeline — `build_generic()` (clear_table + `bulk_create` batches of 200 + serial-sequence reset) × 26 `_build_*()` stages in dependency order; sprite/cry URL synthesis via `RESOURCE_IMAGES/RESOURCE_CRIES` + `POKEMON_SPRITE_CONFIG`: `data/v2/build.py:796-828,2477-2503,101-157`.
- Caching stack — L1 `django-cachalot` (auto ORM query cache) + L2 `django-redis` (`CACHES`) + L3 nginx `proxy_cache_path` (`/tmp/cache`, 2 GB, 10 d inactive) + gzip: `config/settings.py:117-140`, `Resources/nginx/nginx.conf:34-42,69`.
- Settings overlays — `config/settings.py` (prod defaults, Postgres+Redis, `CORS_ALLOW_METHODS=("GET",)`) extended by `config/local.py` (sqlite+DummyCache+DEBUG), `config/docker.py`, `config/docker_compose.py` (env-var Postgres/Redis hosts): `config/local.py:1-22`, `config/docker_compose.py:8-29`.
- GraphQL sidecar — Hasura `graphql-engine:v2.50.2` service + `graphql/v1beta2` metadata + nginx `/v1beta2` proxy with strict `1r/m` limit zone vs permissive `/api/` cache: `docker-compose.yml:50-75`, `Resources/nginx/nginx.conf:67,120-139`.
- Ops / distribution — `Makefile` (migrate/build-db/test/openapi/hasura/k8s targets), `docker-compose*.yml`, `Resources/{docker,compose,nginx,k8s/kustomize,gcp,scripts}`: `Makefile:1-217`.

## Data flow
- Flow 1: REST list/detail read (`pokemon_v2/urls.py:40-106` → `pokemon_v2/api.py:177-182` → `pokemon_v2/serializers.py:277-589`):
  1. `GET /api/v2/{resource}/[?limit&offset][?q=]` hits `PokeAPIRouter`; `PokeapiCommonViewset.list/retrieve` runs `PokeApiManager.get_queryset().order_by("pk")` (`pokemon_v2/models.py:252-254`) + DRF `LimitOffsetPagination(PAGE_SIZE=20)` (`config/settings.py:148-154`).
  2. `NameOrIdRetrieval.get_object()` branches on `ID_PATTERN`/`NAME_PATTERN` to `pk=` vs `name__iexact=` lookup (`pokemon_v2/api.py:137-157`); `?q=` adds `name__icontains` (`pokemon_v2/api.py:127-134`).
  3. `ListOrDetailSerialRelation` picks `*SummarySerializer` (hyperlinked `{name,url}`) for lists vs `*DetailSerializer` for retrieve (`pokemon_v2/api.py:106-114`); detail serializers fan out via reverse-FK `source=` (e.g. `source="abilityeffecttext"`) + extra queries in `SerializerMethodField` (`pokemon_v2/serializers.py:1227-1255`).
  4. `django-cachalot` short-circuits repeated ORM queries via Redis (`config/settings.py:137`); nginx caches `/api/` responses 1 min (`Resources/nginx/nginx.conf:141-153`); CORS limited to `GET` on `^/api/.*$` (`config/settings.py:142-146`).
- Flow 2: CSV → Postgres build (`data/v2/build.py:791-828,2477-2503`):
  1. Operator runs `make build-db` → `manage.py shell --settings=config.local` piping `build_all()` (`Makefile:35-36`); order matters: languages→regions→generations→versions→…→pokemons→encounters→pal_parks.
  2. Each `_build_*()` maps CSV rows to model instances via local `csv_record_to_objects` closures and `build_generic()` which `clear_table()` (delete + `setval(pg_get_serial_sequence…)` / sqlite-sequence reset) then `bulk_create` in ≤200-row batches (`data/v2/build.py:796-827`).
  3. Asset columns are synthesized, not stored: `file_path_or_none()` checks filename against `RESOURCE_IMAGES` (manifest → cache file → local `data/v2/sprites` walk) or `RESOURCE_CRIES` (submodule walk) and prefixes `POKEAPI_SPRITES_PREFIX/POKEAPI_CRIES_PREFIX` URLs; nested `POKEMON_SPRITE_CONFIG` enumerates every variant path (`data/v2/build.py:101-157,778-781`, `data/v2/build.py:157-775`).
  4. Result snapshotted as `pokeapi.pgdump` (`make docker-dump-db/docker-restore-db`, `make update-graphql-data-prod`) and consumed directly by both Django and Hasura (`Makefile:101-107,201-212`).
- Flow 3: Computed aggregation detail (Pokemon / Type / Encounters) (`pokemon_v2/serializers.py:2932-3184,1905-2057`, `pokemon_v2/api.py:1019-1068`):
  1. `PokemonResource.queryset = …select_related("pokemon_species").prefetch_related("pokemonsprites","pokemoncries")` seeds sprite/cry fast path (`pokemon_v2/api.py:871-874`); `get_pokemon_sprites/cries` read prefetched `pokemonsprites/pokemoncries` JSON blobs, with female-sprite fallback fill (`pokemon_v2/serializers.py:2981-3002`).
  2. Moves/held-items/past-abilities-stats-types are grouped in Python: filtered+ordered queryset → `itertools.groupby` by move/item/generation → per-group `{move, version_group_details[]}` dicts with local `vg_cache/mlm_cache` to dedupe nested summary serialization (`pokemon_v2/serializers.py:3004-3081`).
  3. `TypeDetailSerializer.get_type_relationships()` pivots `TypeEfficacy` rows into six `{double,half,no}_damage_{to,from}` lists; `get_type_past_relationships()` replays `TypeEfficacyPast` per generation (filter unintroduced types, remove+re-add entries) (`pokemon_v2/serializers.py:1961-2049`).
  4. `PokemonEncounterView.get()` (sub-resource `^api/v2/pokemon/(?P<pokemon_id>\d+)/encounters`, id-only, unlike main resources) + `LocationAreaDetailSerializer.get_encounters()` both `select_related+prefetch_related` then double-`groupby` (area→version / pokemon→version) computing `max_chance = sum(slot.rarity)` (`pokemon_v2/api.py:1028-1068`, `pokemon_v2/serializers.py:1096-1138`).
- Flow 4: GraphQL + schema/meta auxiliary path (`docker-compose.yml:50-75`, `Resources/nginx/nginx.conf:120-139`, `pokemon_v2/api.py:1083-1115`):
  1. Hasura connects via `HASURA_GRAPHQL_DATABASE_URL` straight to Postgres, bypassing Django ORM/serializers entirely; metadata tracked in `graphql/v1beta2/`, applied with `hasura md apply` (`Makefile:141-148`).
  2. nginx routes `/v1beta2` → `graphql-engine:8080/v1/graphql` with `limit_req zone=graphqlDefaultLimit rate=1r/m burst=200` + 30 m edge cache, vs `/api/` → `app:80` with 1 m cache (`Resources/nginx/nginx.conf:67,120-153`).
  3. `PokeapiMetaView` shells out to `git rev-parse/log/tag` for `{deploy_date,hash,tag}`; `openapi.yml` regenerated from `extend_schema` annotations on every viewset via `manage.py spectacular` (`pokemon_v2/api.py:1083-1115`, `Makefile:59-60`).

## Patterns & rationale
- Pattern: Uniform read-only resource (one `ReadOnlyModelViewSet` per table, no create/update/delete) — why this way / trade-offs: matches immutable reference dataset and lets CORS/middleware be `GET`-only and edge-cached aggressively; cost is no user writes and build-time-only corrections (`config/settings.py:144`, `pokemon_v2/api.py:177-182`).
- Pattern: Summary/Detail serializer duality — why: list payloads stay tiny (`{name,url}` following HATEOAS discovery) while detail carries full localized graph; trade-off is ~2× serializer classes per resource and N+1 risk pushed into `SerializerMethodField`s (`pokemon_v2/serializers.py:277-589`, `pokemon_v2/api.py:99-114`).
- Pattern: Mixin-composed models (`Has*` FK + `Is*` language-payload) — why: Veekun CSV schema is highly normalized (every entity has `*_names`, `*_prose`, `*_flavor_text`); mixins DRY the ~140-table definition; trade-off is deep reverse-relation graph and magic `source="…lowercase…" / related_name="%(class)s"` wiring that must be learned to extend (`pokemon_v2/models.py:273-919`).
- Pattern: Fat serializers, thin views — why: all shaping (groupby, past-vs-present overlays, female-sprite fill, encounter rollups) lives in `get_*` methods with manual per-request caches; keeps `api.py` declarative; trade-off is extra queries per detail (mitigated by `select_related/prefetch_related` + cachalot) (`pokemon_v2/serializers.py:3004-3046`, `pokemon_v2/api.py:871-874`).
- Pattern: DB-as-integration-point for GraphQL (Hasura reads same Postgres, no Python GraphQL layer) — why: zero backend duplication, instant access to all tables/FKs; trade-off is two query stacks to maintain and stricter nginx throttling on GraphQL (`docker-compose.yml:50-68`, `README.md:161-175`).
- Pattern: Env-overlay settings + 12-factor compose — why: same image runs locally (sqlite/dummy-cache), in compose (service DNS `db`/`cache`), and prod (nginx+gunicorn+Hasura) via `BASE_DIR/config/*.py` + env vars; trade-off is four settings files to keep in sync (`config/settings.py:54-115`, `config/local.py:5-16`, `config/docker_compose.py:8-29`).
- Pattern: Deterministic pk-ordering default manager — why: stable pagination/serialization across deploys; explicit `order_by()` overrides where semantic order needed (e.g. `PokemonSpeciesResource…order_by("id")`, encounter groupbys) (`pokemon_v2/models.py:242-254`, `pokemon_v2/api.py:887-890`).
- Deliberate non-pattern: no service/use-case layer, no DTOs beyond serializers, no async — Django ORM → DRF serializer is the whole pipeline; acceptable because domain logic is presentation grouping, not transactions.

## Coupling & extension points
- Coupling: every detail serializer fans into shared vocab tables (`Language`, `VersionGroup`, `Generation`, `Type`, `Move`) — adding a language/version multiplies `*_name/flavor_text` rows across all resources; `PokemonEvolution` is the most FK-dense join (item/move/type/location/region/form/expression) (`pokemon_v2/models.py:1934-2068`).
- Coupling: reverse-relation `source=` strings (e.g. `source="abilityeffecttext"`, `source="pokemongameindex"`) tie serializers to Django auto-generated related names; renaming a model breaks many serializers silently until tests run (`pokemon_v2/serializers.py:1228-1232,2935`).
- Coupling: Hasura metadata + `openapi.yml` are checked-in derivatives of DB + annotations; schema drift requires `make openapi-generate` + `hasura md export/apply` or prod GraphQL/REST docs diverge (`Makefile:59-60,141-148`).
- Coupling: sprite/cry availability depends on submodule checkout + manifest freshness (`POKEAPI_USE_LOCAL_SPRITES`, `POKEAPI_SPRITES_MANIFEST_URL`, `manifest.json.gz` cache with `Last-Modified` check); empty submodules here mean `file_path_or_none` returns `None` (`data/v2/build.py:59-144`).
- Extension: new REST resource = new model(s) in `pokemon_v2/models.py` + `*Summary/*DetailSerializer` in `pokemon_v2/serializers.py` + `*Resource(PokeapiCommonViewset)` in `pokemon_v2/api.py` + `router.register()` in `pokemon_v2/urls.py` + `extend_schema(tags=[…])` + `make openapi-generate`; tests follow `APIData.setup_*` factory pattern in `pokemon_v2/tests.py:15-~1500`.
- Extension: new computed field = `SerializerMethodField` + `extend_schema_field` + targeted `select_related/prefetch_related` in the viewset `queryset` (follow `PokemonResource` / `AbilityResource` examples) (`pokemon_v2/api.py:200-202,871-874`).
- Extension: new dataset column = add CSV under `data/v2/csv/` + `_build_*()` stage in `build_all()` order + migration via `make make-migrations/migrate`; sync with upstream Veekun via `make sync-from-veekun/sync-to-veekun` (`Makefile:128-136`, `data/v2/build.py:2477-2503`).

## Feature/System Inventory (no deep-dives — input for feature analysis)
- Ability resources — abilities + effect/flavor/names/changelog: `pokemon_v2/api.py:199-203`, `pokemon_v2/models.py:1007-1035`
- Berry resources — berries + firmness + flavors + flavor-maps: `pokemon_v2/api.py:215-251`
- Characteristic resources — highest-IV characteristics + descriptions: `pokemon_v2/api.py:263-266`
- Contest resources — contest-types + contest-effects + combos: `pokemon_v2/api.py:279-314`
- Egg-group resources — breeding groups + names: `pokemon_v2/api.py:311-314`
- Encounter resources — methods + conditions + condition-values + slots: `pokemon_v2/api.py:327-363`
- Evolution-chain resources — recursive chain tree builder: `pokemon_v2/api.py:375-378`, `pokemon_v2/serializers.py:3465-3560`
- Evolution-trigger resources — evolution event triggers: `pokemon_v2/api.py:391-394`
- Evolution-variable resources — RPN condition variables + expressions: `pokemon_v2/api.py:407-410`, `pokemon_v2/models.py:1801-1818`
- Generation resources — generations + main-region linkage: `pokemon_v2/api.py:423-426`
- Gender resources — gendered-species + evolution-requirement queries: `pokemon_v2/api.py:439-442`, `pokemon_v2/serializers.py:805-832`
- Growth-rate resources — XP formulas + level tables: `pokemon_v2/api.py:455-458`
- Item resources — items + categories + pockets + attributes + fling-effects + prices + sprites: `pokemon_v2/api.py:471-555`
- Currency resources — purchase currencies: `pokemon_v2/api.py:551-554`
- Language resources — translation vocab + local names: `pokemon_v2/api.py:567-570`
- Location resources — locations + game-indices + names: `pokemon_v2/api.py:583-586`
- Location-area resources — areas + encounter-rate + pokemon-encounter rollups: `pokemon_v2/api.py:599-602`, `pokemon_v2/serializers.py:1058-1138`
- Machine resources — TM/HM item→move→version-group links: `pokemon_v2/api.py:615-618`
- Move resources — moves + changes + flavor/names/meta: `pokemon_v2/api.py:631-634`, `pokemon_v2/serializers.py:2338-`
- Move-damage-class resources — physical/special/status classes: `pokemon_v2/api.py:647-650`
- Move-ailment resources — status-condition vocab: `pokemon_v2/api.py:663-666`
- Move-battle-style resources — Battle-Palace styles: `pokemon_v2/api.py:679-682`
- Move-category resources — move-effect categories: `pokemon_v2/api.py:695-698`
- Move-learn-method resources — level/machine/egg/tutor methods: `pokemon_v2/api.py:711-714`
- Move-target resources — battle targets: `pokemon_v2/api.py:727-730`
- Nature resources — stat-modifying natures + flavor prefs: `pokemon_v2/api.py:743-746`
- Pal-park-area resources — Pal-Park encounter areas: `pokemon_v2/api.py:759-762`
- Pokeathlon-stat resources — Pokéathlon performance stats: `pokemon_v2/api.py:775-778`
- Pokedex resources — regional/national dexes: `pokemon_v2/api.py:791-794`
- Pokemon-color resources — dex color sorting: `pokemon_v2/api.py:807-810`
- Pokemon-form resources — cosmetic forms + sprites + conditions: `pokemon_v2/api.py:823-826`
- Pokemon-habitat resources — terrain habitats: `pokemon_v2/api.py:839-842`
- Pokemon-shape resources — dex shape sorting: `pokemon_v2/api.py:855-858`
- Pokemon resources — pokemon + abilities/moves/stats/types/sprites/cries/held-items: `pokemon_v2/api.py:871-874`, `pokemon_v2/serializers.py:2932-3184`
- Pokemon-species resources — species + genera/flavor/varieties/evolution: `pokemon_v2/api.py:887-890`, `pokemon_v2/serializers.py:3273-`
- Region resources — regions + version-group backlinks: `pokemon_v2/api.py:903-905`
- Stat resources — battle stats + affecting moves/natures/items: `pokemon_v2/api.py:919-922`, `pokemon_v2/serializers.py:1289-1361`
- Super-contest-effect resources — super-contest appeal/flavor: `pokemon_v2/api.py:935-938`
- Type resources — types + damage-relations + past-generation overlays + sprites: `pokemon_v2/api.py:951-954`, `pokemon_v2/serializers.py:1905-2057`
- Version resources — game versions: `pokemon_v2/api.py:967-970`
- Version-group resources — grouped game releases: `pokemon_v2/api.py:983-986`
- Pokemon-encounters sub-resource — per-pokemon location encounters (id-only route): `pokemon_v2/urls.py:101-105`, `pokemon_v2/api.py:1019-1068`
- API meta endpoint — git deploy hash/date/tag: `pokemon_v2/urls.py:99`, `pokemon_v2/api.py:1083-1115`
- Name-or-ID lookup — dual pk/name retrieval on all common viewsets: `pokemon_v2/api.py:117-157`
- Local `?q=` name search — icontains filter (local-only, not on pokeapi.co): `pokemon_v2/api.py:126-134,160-165`
- Sprite asset pipeline — manifest/local-submodule sprite URL synthesis: `data/v2/build.py:101-157,157-775`
- Cry asset pipeline — cries-submodule sound URL synthesis: `data/v2/build.py:148-154,779-781`
- CSV ETL build — wipe-and-reload `build_all()` bulk pipeline: `data/v2/build.py:796-828,2477-2503`
- GraphQL sidecar — Hasura engine over same Postgres: `docker-compose.yml:50-75`, `graphql/v1beta2/`
- OpenAPI schema — drf-spectacular generated `openapi.yml`: `config/settings.py:175-280`, `openapi.yml`
- Multi-layer cache — cachalot + Redis + nginx edge cache: `config/settings.py:117-140`, `Resources/nginx/nginx.conf:69,141-153`
- Environment overlays — local/docker/compose settings variants: `config/local.py`, `config/docker.py`, `config/docker_compose.py`
- Container/k8s deployment — compose stacks, nginx, kustomize, gunicorn: `docker-compose*.yml`, `Resources/`, `gunicorn.conf.py`
