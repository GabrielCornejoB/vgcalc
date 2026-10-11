> **Reference note (moved 2026-10-11):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/pokeapi` (PokeAPI — Django read-only reference-data API). Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Explorer Report — pokeapi

## Stack
- Language(s): Python (>=3.10,<3.15; pinned via `.python-version` to 3.14)
- Framework(s): Django 5.2 + Django REST Framework 3.16; drf-spectacular (OpenAPI); optional GraphQL (Hasura engine, see docker-compose)
- Package manager / build: `uv` (has `uv.lock`); Makefile wrappers; Docker / docker-compose for runtime
- Scripts (build/test/lint/dev):
  - `make install` — install deps (local dev)
  - `make setup` / `make build-db` / `make migrate` / `make make-migrations`
  - `make serve` — run project locally
  - `make test` — run tests
  - `make format` / `make format-check`, `make lint-check` / `make lint-fix`, `make typecheck` (ty)
  - `make openapi-generate` — regenerate `openapi.yml`
  - `make pre-commit` / `make pre-commit-install`
  - Docker: `make docker-up`, `make docker-setup`, `make docker-test`, etc.

## Entry-points
- `manage.py` — Django CLI; sets `DJANGO_SETTINGS_MODULE=config.settings`
- `config/wsgi.py` — WSGI callable `application`
- `config/asgi.py` — ASGI callable `application`
- `config/urls.py` — root URLconf, includes `pokemon_v2.urls`
- `config/settings.py` — main settings (production); `config/local.py`, `config/docker.py`, `config/docker_compose.py` are env overlays
- `pokemon_v2/urls.py` — DRF router registering all API resources
- `gunicorn.conf.py` — Gunicorn server config (bind port from `SERVER_PORT`)
- `openapi.yml` — generated OpenAPI schema (checked in)

## Top-level structure
- `config/` — Django project: settings and wsgi/asgi/urls for the project
- `pokemon_v2/` — main Django app: models, DRF API/serializers, URLs, migrations, tests
- `data/` — data pipeline; `data/v2/build.py` builds the DB, `data/v2/csv/` holds source CSVs, `sprites`/`cries` are git submodules
- `graphql/` — Hasura GraphQL configs for `v1beta` and `v1beta2` (config.yaml, examples, metadata)
- `Resources/` — deployment/infra assets: docker, compose, nginx, k8s/kustomize, kind, gcp, helper scripts
- `.circleci/` — CI config (static checks + tests)
- `docker-compose*.yml` — service orchestration (app, postgres, redis cache, nginx, graphql-engine)
- `Makefile` — primary developer task runner

## Config & tooling
- `pyproject.toml` — project metadata + deps + dev dependency group (`uv`)
- `uv.lock` — locked dependency graph
- `ruff.toml` — linter/formatter config (line-length 120, extensive rule set)
- `ty.toml` — type-checker (`ty`) config
- `.pre-commit-config.yaml` — pre-commit hooks (ruff, ty, CSV checks, etc.)
- `docker-compose.yml` / `docker-compose-dev.yml` / `docker-compose.override.yml` — compose stacks
- `default.nix` — Nix dev shell
- `.gitmodules` — submodules: `data/v2/sprites`, `data/v2/cries`
- `.dockerignore`, `.gitignore`, `.python-version`

## Notes / unknowns
- Not a git repository in this working copy (`git status`/`log` fail), though `.gitmodules` and CI suggest it normally is (PokeAPI/pokeapi). Submodules `data/v2/sprites` and `data/v2/cries` appear empty here.
- GraphQL is served via external Hasura engine configured under `graphql/` + `Resources/`, not a Python GraphQL layer.
- `openapi.yml`, `pokemon_v2/serializers.py`, `pokemon_v2/tests.py`, and `data/v2/build.py` are large generated-ish files; not inspected in depth.
