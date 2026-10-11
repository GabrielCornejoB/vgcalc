# AGENTS.md — vgc

> New application home: `~/Code/Projects/vgc/`.
> Reference analyses live in `reference-reports/` (moved from the playground; full sources were NOT copied).
> Original reference projects (read live source here when verifying claims):
> - `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/damage-calc/` — Smogon calc library + UI, gens 0-9 (0 = Champions)
> - `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/NCP-VGC-Damage-Calculator/` — NCP VGC fork, gens 1-10 (10 = Champions)
> - `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/ps/` — `@pkmn/*` data/protocol substrate + `mods/src/champions*`
> - `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/engine/` — gen 1-2 sim only (out of scope for product code)
> - `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/pokeapi/` — PokeAPI read-only reference-data API (Django + Postgres); relational schema reference, no damage formula

## Purpose

Build a **Pokemon Champions VGC application**, starting from a clean, well-designed **relational data core** and growing toward team building, damage calculation, and related domains later.

- Primary competitive context: **VGC doubles, Pokemon Champions ruleset, Mega Evolution as the only generational gimmick.** No Tera, no Dynamax/Gmax, no Z-Moves, no Stellar (see `reference-reports/CONSOLIDATED-FEATURE-REPORT.md` TL;DR section).
- This repo is the **product**, not the analysis dump. The `reference-reports/` folder is read-only background.

## Conventions

- `AGENTS.md` (this file) is the source of truth for agent behavior.
- `reference-reports/` is **frozen**: do not edit except to fix the provenance banner. Cite as `reference-reports/<proj>/FEATURE-REPORT.md` and always verify against live source under `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/<proj>/`.
- Champions-only means: every entity set must answer "is this legal in Champions?" — model legality as data, not as comments.
- Mega-only means: any Z/Max/Dynamax/Tera/Stellar field encountered in a reference source is **out of scope** — note it as intentionally dropped, do not model it.

## How to work in this repo

1. Read this file first.
2. For data questions, consult `reference-reports/CONSOLIDATED-FEATURE-REPORT.md` first, then the per-project report for the system in question, then live source.
3. Keep changes small and self-contained.
