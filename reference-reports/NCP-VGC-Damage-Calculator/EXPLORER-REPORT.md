> **Reference note (moved 2026-10-07):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/NCP-VGC-Damage-Calculator`. Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Explorer Report — NCP-VGC-Damage-Calculator

## Stack
- Language(s): HTML, JavaScript (vanilla, jQuery-era), CSS. No server-side language.
- Framework(s): None (static site built on jQuery + Select2). Pokémon damage-calculation engine in plain JS.
- Package manager / build: None. No package.json, lockfile, bundler, or build step — files are served/opened directly.
- Scripts (build/test/lint/dev): None declared. Development = open an `.html` file directly.

## Entry-points
- `index.html` (~113 KB) — primary VGC damage calculator (main app; links all `script_res/*` data + calc scripts).
- `za-calc.html` — Legends Z-A calculator (uses `script_res/move_data_za.js`, `cooldown_za.js`).
- `lgpe-calc.html` — Let's Go Pikachu/Eevee calculator.
- `digicalc.html` — Digimon variant (April Fools / Digimon-themed calculator).
- `spamtontest.html`, `noisetest.html` — April Fools mini-pages (use `afd_res/`).
- `standardbody.html` — shared/standalone HTML fragment.
- `legacy_scripts/` + `script_res/` — JS engines loaded by the entry HTML files.
- `jquery-3.1.1.min.js` (root) and `script_res/jquery-2.1.0.min.js` — vendored jQuery.

## Top-level structure
- `script_res/` — main JS engine + data: `damage_*.js` (per-gen formulas), `ap_calc.js`, `statuscalc.js`, `ko_chance.js`, `switch_mode.js`, `sidebars.js`, `setdex_*.js`, `*_data.js`, plus CSS (Select2 / nb_calc / ap_calc) and `jquery-2.1.0.min.js`. Also a stray file literally named `css`.
- `legacy_scripts/` — archived older calculators/setdexes (`damage_bw.js`, `damage_sm.js`, `setdex_*`, status data) kept for reference.
- `afd_res/` — April Fools assets: `spamton.js`, `noise.js`, `pokedex_afd24.js`, `setdex_afd24.js`, plus `old_afd/` and `unused_afd/` subfolders.
- `image_res/` — UI images, logos, sprites, favicons; subdirs `aprilfoolsday/` and `minisprites/`.
- `sound_res/` — `.ogg` audio used by the April Fools pages.
- Root `*.html` — the multiple calculator front-ends listed above.
- `README.md` — project description (VGC 2026 calculator, MIT, maintained by nerd-of-now).
- `LICENSE` — MIT.
- `debug.log` — stray Windows Chromium debug output (not part of app).
- `index.sublime-workspace` — editor workspace file.
- `.gitignore` — ignores `/.vs` only.

## Config & tooling
- No config files found (no package.json, pyproject.toml, Cargo.toml, go.mod, requirements, CI config).
- `index.sublime-workspace` indicates Sublime Text workflow.
- `.gitignore` present but directory is **not** a git repository (no `.git`).

## Notes / unknowns
- Repo is a purely static, client-side app; all logic runs in browser, all data shipped as JS globals.
- Version/branching is handled by swapping `setdex_*.js` / `damage_*.js` includes rather than build tooling.
- `script_res/css` is a 2.2 KB file (not a directory) — likely misnamed; contents not inspected.
- No tests, linting, or automation exist.
- Exact division of responsibility between the several root HTML files is inferred from filenames/titles; not deep-dived.
