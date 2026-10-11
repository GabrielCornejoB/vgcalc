> **Reference note (moved 2026-10-07):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/NCP-VGC-Damage-Calculator`. Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Feature Report — NCP-VGC-Damage-Calculator

> Sources: ARCHITECTURE-REPORT.md (inventory), EXPLORER-REPORT.md, direct reads. Arch report was present; inventory below follows it and was verified against code.

## Feature inventory
| Feature | One-line purpose | Key files |
| VGC main damage calculator (gens 1–10 incl. Champions) | Dual-panel 4-moves-each calc + field, auto-recalcs on any input | `index.html:119-835`, `script_res/ap_calc.js:1729-1793`, `script_res/damage_*.js` |
| Legends Z-A calculator variant | Cooldown-based Z-A ruleset reusing main engine | `za-calc.html:697-720`, `script_res/move_data_za.js:1-60`, `script_res/cooldown_za.js:1-34` |
| LGPE calculator variant | Gen 7.5 (Let's Go) ruleset page | `lgpe-calc.html:757-778`, `script_res/ap_calc.js:2290-2300` |
| Digimon April-Fools calculator | Reskinned variant (currently dangling refs) | `digicalc.html:434-442` |
| Spamton/Noise April-Fools mini-pages | Joke event pages + in-calc easter eggs | `spamtontest.html`, `noisetest.html`, `afd_res/spamton.js:3`, `afd_res/noise.js:3`, `sound_res/` |
| Per-gen damage formula modules | Isolated per-ruleset damage math + dispatcher | `script_res/damage_MASTER.js:4-25`, `damage_rby.js:1`, `damage_gsc.js:1`, `damage_rse.js:1`, `damage_dpp.js:1`, `damage_xy.js:4`, `damage_SV.js:6-72` |
| KO-chance / EOT / hazard engine | Roll distribution → OHKO/2HKO text with hazards + end-of-turn | `script_res/ko_chance.js:1-120,466-525` |
| Stat formula modules | HP/stat computation per system (DVs/EVs/AVs/SPs) | `script_res/stat_data.js:5-119` |
| Static dex data packs | Species/moves/types/items/abilities/natures as JS globals | `script_res/pokedex.js:1-40`, `script_res/move_data.js:1-40`, `script_res/type_data.js:1-50`, `script_res/nature_data.js:1-27`, `script_res/item_data.js`, `script_res/ability_data.js` |
| Curated preset-set libraries | Per-format sample sets merged per gen | `script_res/setdex_ncp-g*.js`, `setdex_showdown.js`, `setdex_smogon.js`, `setdex_globalLink.js`, `setdex_nuggetBridge.js`, `setdex_tt2017/2018/2019.js`, `setdex_boss-titan.js` |
| Setdex merge/orchestration layer | Gen→sources composition into one `CURRENT_SETDEX` | `script_res/setdex_dynamic.js:1-27`, `script_res/setdex_dynamic-main.js:1-18`, `script_res/setdex_dynamic-lgpe-za.js:1-6` |
| Custom-set save/export/import/delete | User sets + Showdown interop, localStorage-backed | `script_res/setdex_custom.js:1-109,123-260,475-600,601-790` |
| Sidebar team staging | 6 slots/side quick teams, persisted per-gen | `script_res/sidebars.js:1-233` |
| Field-condition system | Weather/terrain/screens/hazards/side model | `script_res/ap_calc.js:2092-2198`, `index.html:314+` (field panel) |
| Theme + dex + gen preference switches | Light/dark, VGC/National dex, auto-level, gen restore | `script_res/switch_mode.js:1-80`, `script_res/ap_calc.js:2202-2411` |
| Custom mechanics-test modifiers | Hidden BP/Attack/Defense/Final override harness | `script_res/ap_calc.js:40-85`, `script_res/switch_mode.js:111-133` |
| Z-A move-cooldown readout | Real-time cooldown seconds/frames per move | `script_res/cooldown_za.js:1-34`, `script_res/damage_SV.js:64-67`, `script_res/ap_calc.js:64-68,1864-1866` |
| Transform / forme / lock-item coordination | In-calc forme/item legality fixes at snapshot time | `script_res/ap_calc.js:94-160,1937-1997` |
| Tera / Dynamax / Gmax / Z-Move toggles | Per-move battle-gimmick switches | `index.html:243-261` (approx), `script_res/ap_calc.js:1954-1997,2070-2090`, `script_res/damage_MASTER.js:840-976`, `script_res/damage_SV.js:86-93` |
| Importable/exportable team text (Showdown format) | Clipboard round-trip of sets and sidebar teams | `script_res/setdex_custom.js:591-783`, `script_res/sidebars.js:202-213` |
| Legacy archived calculators/setdexes | Reference-only prior-gens | `legacy_scripts/` |
| Static asset packs | Sprites/icons/audio | `image_res/`, `sound_res/` |
| April-Fools overlay (gen-9 dex injection) | Date-gated joke chars merged into gen-9 pool | `script_res/statuscalc.js:3-42`, `afd_res/pokedex_afd24.js`, `afd_res/setdex_afd24.js` |

## Feature details
### VGC main damage calculator (gens 1–10 incl. Champions)
- What it does: Core product. Two mirrored Pokémon panels (`#p1`/`#p2`) each with 4 moves; any `.calc-trigger` input change re-snapshots both sides + field, runs all 8 damage calcs, renders min/max % overview plus a detailed `#mainResult`/`#damageValues` line for the selected `.result-move` radio.
- Where it lives (`path` refs): `index.html:119-835` (panels, field, results); `script_res/ap_calc.js:1729-1793` (`calculate()`), `script_res/ap_calc.js:1795-1871` (`.result-move` render), `script_res/ap_calc.js:1927-2090` (`Pokemon`/`getMoveDetails`), `script_res/ap_calc.js:2202-2440` (gen switch), `script_res/ap_calc.js:2661-2692` (boot).
- How it works: `calculate()` (`script_res/ap_calc.js:1729-1735`) bails if `calcQueue` guard is set (`script_res/ap_calc.js:14-15,1730`), else `new Pokemon($("#p1"))`, `new Pokemon($("#p2"))`, `new Field()`, then `damageResults = calculateAllMoves(p1, p2, field)` where `calculateAllMoves` is a function pointer swapped by gen. Per-move min/max via `calcMinMaxDamage` (`script_res/ap_calc.js:1742,1556`), KO text via `getKOChanceText` (`script_res/ap_calc.js:1750,1770`, `script_res/ko_chance.js:1`). Best-damage radio auto-checked unless `.locked-move` pinned (`script_res/ap_calc.js:1783-1790`).
- Inputs/outputs & data deps: Inputs are the DOM panels + `Field`; outputs are `damageResults` global, overview `%` labels, `#mainResult` description + `damageText -- koChanceText`, `#damageValues` roll list (with multi-hit per-hit breakdown at `script_res/ap_calc.js:1808-1841`).
- Entry-points (UI/CLI/API): Open `index.html`; pick gen radio (`.gen`), pick sets or hand-edit panels, toggle field panel; click a result radio for details. No CLI/API.

### Legends Z-A calculator variant
- What it does: Same engine restricted to the Z-A ruleset (gen `9.5`): real-time-action cooldowns shown per move instead of pure turn-based output.
- Where it lives (`path` refs): `za-calc.html:697-720` (script composition); `script_res/move_data_za.js:1-60` (allowlist + `cooldown` overlay); `script_res/cooldown_za.js:1-34`; hook in `script_res/damage_SV.js:64-67`; surfacing in `script_res/ap_calc.js:64-68,1864-1866` (`result.cooldown`); gen case at `script_res/ap_calc.js:2323-2333`.
- How it works: `MOVES_ZA` starts as an allowlist copied from `MOVES_ZA_NATDEX` (`script_res/move_data_za.js:3-36`) then overlaid with `{cooldown, bp, hitRange, plusEffects}` via `$.extend` (`script_res/move_data_za.js:38+`). `CALCULATE_ALL_MOVES_SV` attaches `results[*][i].cooldown = getMoveCooldown(...)` only when `gen == 9.5` (`script_res/damage_SV.js:64-67`). `getMoveCooldown` (`script_res/cooldown_za.js:1-12`) uses `move.plusEffects.cooldown` when `isPlusMove`, speed floor 25, `speedPower`, Drowsy ×2 frame divisor, `max(3, …)` floor, 30 fps frame→seconds conversion; `getCooldownText` (`script_res/cooldown_za.js:14-34`) formats `"<move>: <s> seconds (<frames> frames)"` with Drowsy/Speed-Power qualifiers.
- Inputs/outputs & data deps: Inputs: move `cooldown` table, attacker's `rawStats[SP]`, `speedPower`, `status == "Drowsy"`. Output: per-result `cooldown` string.
- Entry-points (UI/CLI/API): Open `za-calc.html`; uses minimal `COMPONENTS` map (`script_res/setdex_dynamic-lgpe-za.js:1-6`) so only gens 7.5/9/9.5 exist there.

### LGPE calculator variant
- What it does: Let's Go Pikachu/Eevee page implementing gen `7.5` rules: no abilities/items, AVs (Awakening Values) + friendship modifier, LGPE stat formulas.
- Where it lives (`path` refs): `lgpe-calc.html:757-778` (loads only `damage_SV.js`, no `damage_xy/rby/...`); `script_res/ap_calc.js:2290-2300` (gen 7.5 case: `POKEDEX_LGPE`, `MOVES_LGPE`, `items=[]`, `abilities=[]`, `CALC_HP/STAT_LGPE`).
- How it works: Reuses `CALCULATE_ALL_MOVES_SV` for damage but swaps stat functions to `CALC_HP_LGPE`/`CALC_STAT_LGPE` (`script_res/stat_data.js:60-90`): HP = `floor((base*2+ivs)*level/100)+level+10+avs`; other stats add nature × friendship (`1 + floor(10*friendship/255)/100`) + AVs. `Pokemon` snapshot reads `.avs`/`.friendship` only when `gen == 7.5` (`script_res/ap_calc.js:2020-2023,2032-2034,2064-2066`); transform backup also covers AVs (`script_res/ap_calc.js:118-120`).
- Inputs/outputs & data deps: Same calc I/O as main; data pack is the LGPE slice of `pokedex.js`/`move_data.js`.
- Entry-points (UI/CLI/API): Open `lgpe-calc.html`.

### Digimon April-Fools calculator
- What it does: Reskinned joke/variant calculator reusing the panel/calc shell with Digimon data. Currently broken as-shipped.
- Where it lives (`path` refs): `digicalc.html:434-442` (script tags).
- How it works: `digicalc.html:434-442` references `script_res/digidex.js`, `script_res/digi_move_data.js`, `script_res/digisetdex.js`, `script_res/damage.js` — none of which exist in `script_res/` (glob shows only `damage_MASTER/damage_*.js`, standard data files). This matches the arch-report note that `digidex` refs point at files no longer present; page will 404 those scripts if opened.
- Inputs/outputs & data deps: Intended to reuse `ap_calc.js` + `stat_data/type_data/nature/ability/item_data` + `ko_chance.js` with swapped dex/moves.
- Entry-points (UI/CLI/API): Open `digicalc.html` (expect missing-script errors).

### Spamton/Noise April-Fools mini-pages
- What it does: Standalone joke pages plus in-calc easter eggs where selecting the `Spamton` or `The Noise` set triggers animations/audio.
- Where it lives (`path` refs): `spamtontest.html`, `noisetest.html`; `afd_res/spamton.js:3` (`run_spamton`), `afd_res/noise.js:3` (`run_noise`); triggers in `script_res/ap_calc.js:1171-1174`; audio built in `script_res/statuscalc.js:23-33` from `sound_res/*.ogg`; archived variants in `afd_res/old_afd/`, `afd_res/unused_afd/` (incl. `delta_damage.js:1` `CALCULATE_DAMAGE_DELTARUNE`).
- How it works: `$(".set-selector").change` checks `pokemonName === 'Spamton'/'The Noise'` and calls the corresponding runner (`script_res/ap_calc.js:1171-1174`), which drives DOM/audio gags. The 2024 overlay path (`checkaprilfools`/`addjokechars`) injects `POKEDEX_AFD24`/`SETDEX_AFD24` into the gen-9 pool (see overlay feature below).
- Inputs/outputs & data deps: `afd_res/pokedex_afd24.js`, `setdex_afd24.js`, `sound_res/*.ogg`, `image_res/aprilfoolsday/`.
- Entry-points (UI/CLI/API): Open `spamtontest.html`/`noisetest.html`, or pick the joke sets in `index.html` on/around April 1st (or with `localStorage.isAFD` set).

### Per-gen damage formula modules
- What it does: Each generation's authoritative damage equation; a dispatcher routes to the right one so the rest of the app stays gen-agnostic.
- Where it lives (`path` refs): Dispatcher `GET_DAMAGE_HANDLER` at `script_res/damage_MASTER.js:4-25`; `CALCULATE_ALL_MOVES_RBY`/`CALCULATE_DAMAGE_RBY` in `script_res/damage_rby.js:1-60`; GSC `script_res/damage_gsc.js:1-20`; ADV `script_res/damage_rse.js:1-24`; DPP `script_res/damage_dpp.js:1-28`; XY (covers gens 5–6) `script_res/damage_xy.js:4-45`; SV (covers gens 7/8/9/10) `CALCULATE_ALL_MOVES_SV:6`, `GET_DAMAGE_SV:72` in `script_res/damage_SV.js`.
- How it works: `GET_DAMAGE_HANDLER` switches on global `gen` (1→RBY … 5/6→XY, 7/8/9/10→SV). Each `CALCULATE_ALL_MOVES_*` applies pre-calc mutators (boosts, weather/ability stat rewrites, weight) then calls its single-move function 4×2. Examples: RBY clamps stats to 1–999 and uses type-category physical/special split with single Special stat `SL` (`script_res/damage_rby.js:1-58`); SV runs ~20 `check*` hooks (Trace, Neutralizing Gas, Forecast, Terastal, Paradox, Intimidate, Download, … at `script_res/damage_SV.js:7-38`), finalizes boosted stats + `getFinalSpeed`, then `GET_DAMAGE_SV` handles Me First/type-change/Z/Max, `basePowerFunc` (`script_res/damage_MASTER.js:1304`) + `calcBPMods` (`script_res/damage_MASTER.js:1568`), STAB/effectiveness/crit/spread, 16-roll array. Shared helpers: `chainMods` (`script_res/damage_MASTER.js:193`), `getModifiedStat` (`:296`), `getFinalSpeed` (`:307`), `buildDescription` (`:31`).
- Inputs/outputs & data deps: Inputs `(attacker, defender, move, field/side)` snapshots; output `{damage: [...rolls or per-hit arrays...], description}`.
- Entry-points (UI/CLI/API): No direct UI; reached via `calculate()` → `calculateAllMoves` pointer set in the gen switch (`script_res/ap_calc.js:2220-2343`).

### KO-chance / EOT / hazard engine
- What it does: Converts a raw 16-roll damage array into human KO text (`guaranteed OHKO`, `~62% chance to 2HKO`, …), accounting for hazards, end-of-turn healing/chip, Berry timing, and multi-hit distributions.
- Where it lives (`path` refs): `script_res/ko_chance.js:1-120` (entry + hazards + EOT aggregation), `:176` (`damageArrToDict`), `:215` (`arrayToProbabilityDict`), `:245` (`getKOChance`), `:319` (`verifyKOChance`), `:466-525+` (`getAllEndOfTurnEffects`), `:742-744` (`END_TURN_ORDER_*`), `:737` (`checkThresholdCriteria`), `:746` (`eotProcess`).
- How it works: Early-outs for breakage/Pain Split/status/zero-damage incl. Harsh-Sun/Heavy-Rain negation (`script_res/ko_chance.js:1-18`). Berry handling via `getRestoreHP`/`applyRipen`/`getRestoreThreshold` with Knock Off/Thief/Covet/Psychic Noise/Bug Bite/Pluck/Incinerate suppression (`:19-31`). Hazards: Stealth Rock + Steelsurge by Rock/Steel effectiveness, Spikes by layers/groundedness, all negated by Magic Guard/Heavy-Duty Boots (`:57-87`). EOT dict summed in gen-specific order (`END_TURN_ORDER_GEN_3/4/5_ONWARDS` at `:742-744`; Sun/Rain/Sand/Hail, Leftovers/Black Sludge, Aqua Ring, Ingrain, Leech Seed, poison/Toxic counter, burn, Nightmare, Curse, Salt Cure, binding, Bad Dreams, Sticky Barb at `:466+`). Then `getKOChance(damage, multihit, curHP - hazards, …)` layers Berry-heal threshold timing to emit the final string (`:108-173`).
- Inputs/outputs & data deps: Inputs `(damage[], move, defender, side, isBadDreams, isItemlessAttacker)`; output plain-text KO sentence consumed by `calculate()`.
- Entry-points (UI/CLI/API): Rendered per move in overview + `#mainResult`; no standalone UI.

### Stat formula modules
- What it does: Computes displayed HP and per-stat totals from base/level/DVs/IVs/EVs/AVs/friendship/SPs + nature, per game system.
- Where it lives (`path` refs): `script_res/stat_data.js:5-25` (RBY DVs), `:27-58` (ADV EVs/IVs + nature, Shedinja-1 + Dynamax-×2 HP), `:60-90` (LGPE AVs/friendship), `:92-119` (Champions SPs, fixed level 50, Shedinja-1 + Dynamax-×2).
- How it works: Key formulas — RBY HP `floor(((base+dvs)*2+63)*level/100)+level+10`; ADV HP `floor((base*2+ivs+floor(evs/4))*level/100)+level+10` (×2 if `.max` checked); LGPE HP adds `+avs`, stats multiply nature × friendship then `+avs`; Champions HP `floor((base*2+31)*50/100)+50+10+statPoints`, stats `floor(floor((base*2+31)*50/100)+5+statPoints)*nature`. All write `.total`/`.max-hp`, then `calcCurrentHP` + `updateHPBar`. Wired to `.level` keyup and per-stat keyup/change handlers in `script_res/ap_calc.js:88-92,189-356` via `calcHP`/`calcStat` pointers.
- Inputs/outputs & data deps: Reads panel inputs; writes `.total` text that `new Pokemon()` later reads as `rawStats/maxHP` (`script_res/ap_calc.js:2003-2025`).
- Entry-points (UI/CLI/API): Editing level/base/EV/IV/AV/SP/nature fields in either panel.

### Static dex data packs
- What it does: Ships the entire game database as versioned JS globals so the app works from `file://` with zero fetch.
- Where it lives (`path` refs): `script_res/pokedex.js:1-40+` (`POKEDEX_RBY…POKEDEX_CHAMPIONS`, `POKEDEX_*_NATDEX`); `script_res/move_data.js:1-40+` (`MOVES_RBY…` with `{bp, type, category, makesContact, …}`); `script_res/move_data_za.js:1-60` (`MOVES_ZA`); `script_res/type_data.js:1-50+` (`TYPE_CHART_RBY/GSC/BW/XY/SV`, each type carrying `category` for pre-split gens); `script_res/item_data.js`, `script_res/ability_data.js`, `script_res/nature_data.js:1-27` (`NATURES` as `[plus, minus]` pairs).
- How it works: Pokédex entries `{t1, t2, bs:{hp,at,df,sa,sd,sp(,sl)}, w, ab, canEvolve, formes?}` (`script_res/pokedex.js:2-17`); moves carry flags consumed by calc branches (`hitRange`, `plusEffects`, `isPledge`, `linearAddBP`, `zp`, `isZ/isSignatureZ`, `cooldown`); gen switch points `pokedex/moves/typeChart/items/abilities` globals at the right table (`script_res/ap_calc.js:2212-2344`), with NatDex branching on `localStorage.dex` for gens 8/9/10.
- Inputs/outputs & data deps: No runtime inputs; consumed by set loading, select-population (`script_res/ap_calc.js:2386-2403`), and damage functions.
- Entry-points (UI/CLI/API): Indirect — species/move/ability/item `<select>` options.

### Curated preset-set libraries
- What it does: Provides one-click competitive spreads (EVs/nature/ability/item/moves) per format so users rarely hand-enter stats.
- Where it lives (`path` refs): `script_res/setdex_ncp-g3.js`, `-g5/-g6/-g7/-g8/-g9/-g10.js`, `script_res/setdex_showdown.js`, `setdex_smogon.js`, `setdex_globalLink.js`, `setdex_nuggetBridge.js`, `setdex_tt2017.js`, `setdex_tt2018.js`, `setdex_tt2019.js`, `setdex_boss-titan.js`; composed by `COMPONENTS` (next feature).
- How it works: Each file defines `SETDEX_*` globals shaped `{Species: {SetName: {evs|sps, ivs, nature, ability, item, moves[], level?, tera_type?}}}`; later sources win on merge (`$.extend` in `script_res/setdex_dynamic.js:22`). Gen 8 pins VGC2022/2021/BFD slices; gen 9/10 use `SETDEX_GEN9`/`SETDEX_GEN10`.
- Inputs/outputs & data deps: Loaded via `<script>` tags in `index.html:857-870`; selected via `.set-selector` → panel population (`script_res/ap_calc.js:1164-1260+`).
- Entry-points (UI/CLI/API): The set autocomplete dropdown in each panel.

### Setdex merge/orchestration layer
- What it does: Decides which preset libraries participate for the active gen and merges them plus user customs into one lookup.
- Where it lives (`path` refs): `script_res/setdex_dynamic.js:1-27` (`loadSetdexScript`, `CURRENT_SETDEX`); `script_res/setdex_dynamic-main.js:1-18` (`COMPONENTS` for `index.html`); `script_res/setdex_dynamic-lgpe-za.js:1-6` (minimal map for variant pages).
- How it works: `loadSetdexScript()` concatenates `COMPONENTS[gen] + ALL_SETDEX_CUSTOM[gen]`, shows/hides `.set-checkbox` on customs presence, then `$.extend`s each source into `CURRENT_SETDEX`/`setdex` (`script_res/setdex_dynamic.js:5-27`). Called on every gen switch (`script_res/ap_calc.js:2352`) and after custom save/delete. Variant pages use an empty-by-default map so they boot without the full library set.
- Inputs/outputs & data deps: Inputs `COMPONENTS[gen]`, `ALL_SETDEX_CUSTOM[gen]`; output global `setdex`.
- Entry-points (UI/CLI/API): No direct UI; runs automatically.

### Custom-set save/export/import/delete
- What it does: Lets users persist their own spreads per gen, import Showdown pastes (with forme/Hidden-Power handling), and delete them.
- Where it lives (`path` refs): Forme maps `showdownToCalcFormes:1` / `calcToShowdownFormes:46` + `defaultHiddenPowerSD:63`/`calcHiddenPower:87`/`verifyHiddenPowerType:82` in `script_res/setdex_custom.js:1-93`; per-gen localStorage hydration `:96-109` (`custom_gen_<gen>`); `processSave:175` (paste parser incl. gender-strip, Gmax/Mega/Primal handling at `:230-244`); `savecalc:475`, `runSaveCalc:576`, `saveCustomSidebar:465`, `saveSets:123`, `deleteSet:786`, `runDeleteSet:794`, `deletecustom:111`, `storedSetFixes:804`.
- How it works: Imports split pastes on blank lines, parse the 11-line Showdown format documented at `:199-211`, map formes both directions, synthesize Hidden-Power IVs/DVs from `defaultHiddenPowerSD` when Showdown omits them, then store under `ALL_SETDEX_CUSTOM[gen][species][spreadName]` + `localStorage["custom_gen_"+gen]` and re-run `loadSetdexScript()`. Sidebar slots are just reserved set names (`Left/Right Sidebar Slot N`) saved via `savecalc(new Pokemon(...), spreadName, …)` (`script_res/sidebars.js:16,33`).
- Inputs/outputs & data deps: Inputs: panel state or paste text; outputs: localStorage + merged `setdex`.
- Entry-points (UI/CLI/API): Save/export/delete buttons per panel (`.setCalc/.delset` hooks), `savecustom`/`deletecustom` global buttons.

### Sidebar team staging
- What it does: Two 6-slot sidebars (one per calc side) for staging full teams, with sprites, persistence, and Showdown export.
- Where it lives (`path` refs): `script_res/sidebars.js:1-50` (add/edit/remove), `:52-74` (`loadSidebar`), `:76-88` (`reloadSidebar`), `:89-151` (`getSidebarImg`), `:153-169` (preset/slot loading), `:171-194` (`shiftSidebar`), `:196-218` (remove/clear/export team), `:220-233` (button visibility).
- How it works: `CURRENT_SIDEBARS = [[left6],[right6]]` persisted as `localStorage['g'+gen+'_sidebars']` (`script_res/sidebars.js:11,55`); add/edit snapshot `new Pokemon()` into a reserved custom-set name and resolve sprite via `getSidebarImg`, which handles Mega (`MEGA_STONE_USER_LOOKUP`), Primal, Rayquaza-Dragon-Ascent, Meloetta-Relic-Song, Keldeo-Secret-Sword, Darmanitan-Zen, Greninja-Battle-Bond, Gmax, Urshifu singles/rapid images, with Smogon-minisprite → fallback → `image_res/empty-slot.png` chain (`:100-151`). `shiftSidebar` repairs name/slot desync on delete (`:171-194`); `loadSidebarSlot` refuses off-dex species (`:159-169`).
- Inputs/outputs & data deps: Reads current panel + `setdex`/`pokedex`; writes customs + localStorage.
- Entry-points (UI/CLI/API): Sidebar add/edit/load/clear/export buttons flanking the panels (`#sbL/#sbR`, `#edL/#edR`, slot images `#pkmnL1-6/#pkmnR1-6`).

### Field-condition system
- What it does: Models everything that isn't a Pokémon: format (singles/doubles), weather, terrain, gravity, screens, hazards, Helping Hand, auras, seeds, salt cure, binding, Leech Seed, Nightmare, etc.
- Where it lives (`path` refs): `Field` (`script_res/ap_calc.js:2092-2162`) + `Side` (`:2164-2198`); reset in `clearField` (`:2413+`); panel markup `#field-panel` in `index.html:314+`; consumption across `damage_SV.js`/`damage_MASTER.js`/`ko_chance.js`.
- How it works: `Field()` snapshots ~30 radio/checkbox states, with gen branches (GSC weather/spikes at `:2099-2101`, Z-A single-layer spikes at `:2106-2108`) and Neutralizing-Gas suppression of Friend Guard/Battery/Power Spot/Steely Spirit/Flower Gift (`:2117-2122`). `getSide(i)` builds per-side view with intentional L/R swaps for effects that "affect attacks against opposite side" (Helping Hand, Power Spot, Steely Spirit, Flower Gift Atk, Red Item at `:2114-2128`). `clearField()` defaults to doubles + clear weather/terrain on every gen change (`:2414-2417`).
- Inputs/outputs & data deps: Inputs: field panel controls; outputs: `field.getSide(0/1)` objects consumed by damage + KO text.
- Entry-points (UI/CLI/API): The central field panel between/below the two Pokémon panels.

### Theme + dex + gen preference switches
- What it does: Persists UI preferences: light/dark theme, VGC vs National dex data pool, auto-level (50 vs 100), last gen.
- Where it lives (`path` refs): `script_res/switch_mode.js:1-80` (theme/dex/level init + click handlers, `loadTheme`, `loadDex`, `loadSVColors`), `:82-109` (legacy WIP-set toggle), `:111-133` (custom-mods gate); `getGen` + boot restore in `script_res/ap_calc.js:2606-2618,2664`; gen persistence in `$(".gen").change` (`:2202-2208`); visibility toggles for `.gen-specific`/`.natdex-specific` at `:2350-2384`.
- How it works: `localStorage` keys `theme` (default follows `prefers-color-scheme` at `:3`), `dex` (`natdex|vgcdex`), `level`, `gen`. `loadDex("natdex")` swaps gens 8/9/10 to `POKEDEX_*_NATDEX`/`MOVES_*_NATDEX`/`ITEMS_*_NATDEX` tables (`script_res/ap_calc.js:2302-2339`); level toggle drives default set level 50 vs 100 in `$(".set-selector").change` (`:1214-1225`).
- Inputs/outputs & data deps: localStorage only; drives which globals/selects are populated.
- Entry-points (UI/CLI/API): `#switchTheme`, `#switchDex`, `#douswitch` controls in page header.

### Custom mechanics-test modifiers
- What it does: Hidden power-user harness to multiply Base Power / Attack / Defense / Final-roll chains by arbitrary ×/4096 factors for mechanics research.
- Where it lives (`path` refs): UI builders `newCustomMod/deleteCustomMod/checkCustomMods` + `mechanicsTests` global in `script_res/ap_calc.js:40-85`; gate `toggleCustomModifiers` in `script_res/switch_mode.js:128+` behind `#toggleCustMods` confirm (`:111-124`); consumption in damage paths via `pokemon.customModifiers/hasCustomModifiers` (`script_res/ap_calc.js:2056-2063`) and `#customModValues` echo (`:1842-1860+`).
- How it works: Each `.cmodlist` row is a `(bpMod|atMod|dfMod|fnMod, value/4096)` pair; non-4096 entries accumulate into `mechanicsTests` and are attached to each snapshotted `Pokemon` when `isCustomMods` is on. Visible only for gen ≥ 5 (`script_res/ap_calc.js:2353-2360`).
- Inputs/outputs & data deps: Inputs `#cMod` rows; output modified damage + textual echo of active mods.
- Entry-points (UI/CLI/API): `#toggleCustMods` checkbox (confirm-gated) → `#cMod` add/remove UI.

### Z-A move-cooldown readout
- What it does: Shows how many seconds/frames until the same move can be used again in Legends Z-A's real-time combat.
- Where it lives (`path` refs): `script_res/cooldown_za.js:1-34`; `script_res/damage_SV.js:64-67`; `script_res/ap_calc.js:64-68,1864-1866`.
- How it works: See Z-A variant above for formula. Only computed when `gen == 9.5`; displayed alongside the detailed result (`result.cooldown`).
- Inputs/outputs & data deps: Same inputs as Z-A variant.
- Entry-points (UI/CLI/API): `za-calc.html` detailed-result pane.

### Transform / forme / lock-item coordination
- What it does: Keeps Transform, alternate formes, and forme-locked items legal without user intervention.
- Where it lives (`path` refs): Transform checkbox handler `script_res/ap_calc.js:94-160+` (`preTransformVars`/`transformSpecies`); snapshot-time fixes in `Pokemon` at `script_res/ap_calc.js:1937-1997`.
- How it works: Checking `.transform` copies opponent's types + base/IV/EV(/AV/SPS) stats into the panel while stashing originals in `preTransformVars`; unchecking restores them. At snapshot, `LOCK_ITEM_LOOKUP[this.name]` forces the item select (e.g. Griseous Orb→Griseous Core rename for gen ≥ 9 at `:1938-1940`) and disables the item dropdown; Zacian/Zamazenta/Eternatus disable Dynamax (`:1955-1962`); Ogerpon mask↔Tera linkage (`:1965-1978`) and Terapagos Stellar lock (`:1979-1989`) enforce tera legality. Transformed name is suffixed `" (source)"` at `:2055`.
- Inputs/outputs & data deps: Panel fields + `LOCK_ITEM_LOOKUP` + `pokedex[...].formes/canEvolve`.
- Entry-points (UI/CLI/API): `.transform` checkbox per panel; `.forme` select; automatic during `calculate()`.

### Tera / Dynamax / Gmax / Z-Move toggles
- What it does: Per-move/per-Pokémon battle gimmicks that rewrite types, base powers, HP, and move identity.
- Where it lives (`path` refs): Checkboxes `.tera/.max/.gmax/.move-z/.move-plus/.move-stellar` snapshotted in `Pokemon`/`getMoveDetails` (`script_res/ap_calc.js:2009-2011,2070-2089`); `checkTerastal` in `script_res/damage_MASTER.js:438-445` (also called from `damage_SV.js:16-17`); `ZMoves:840` + `MaxMoves:892` in `script_res/damage_MASTER.js`; Dynamax application in `GET_DAMAGE_SV` (`script_res/damage_SV.js:86-93`) and HP doubling in `CALC_HP_ADV/CHAMP` (`script_res/stat_data.js:39-41,102-104`); Gmax sprite/side handling (`script_res/sidebars.js:130-132`, `Field` Gmax fields).
- How it works: Terastal replaces `type1/teraSTAB` (non-Stellar) before effectiveness/STAB; Z toggles map the move through `ZMOVES_LOOKUP`/signature-Z tables with bracketed BP (`100→200` scale at `damage_MASTER.js:866-877`) and force crit/hits passthrough; Dynamax (`gen 8` only in `GET_DAMAGE_SV:86`) maps to Max moves via `MaxMoves` exception tables (`:893-899+`) and quarters through Protect; Gmax flag alters sprite + G-Max field effects; ZA `move-plus` swaps to `plusEffects` (cooldown/hitRange at `cooldown_za.js:2`, `ap_calc.js:2080-2088`).
- Inputs/outputs & data deps: Panel gimmick checkboxes + `moves[...].zp/isZ/isSignatureZ/canDouble/isPledge/linearAddBP` flags.
- Entry-points (UI/CLI/API): Gimmick checkboxes on each move row + `.tera/.max/.gmax` on each panel; visibility gated by `.gen-specific` classes.

### Importable/exportable team text (Showdown format)
- What it does: Clipboard round-trip so teams move between this calc, Pokémon Showdown, and other tools.
- Where it lives (`path` refs): `Clipboard_CopyTo` (`script_res/setdex_custom.js:591-599`, deprecated `execCommand("copy")`); `exportset` (`:601-772`); `runExportSet` (`:774-783`); `exportSidebarTeam` (`script_res/sidebars.js:202-213`); import via `processSave` (`script_res/setdex_custom.js:175+`).
- How it works: Export builds `Species @ Item / Ability / Level / Tera Type / EVs-or-SPs / Nature / IVs / - moves` text, skipping `(No Move)`, omitting default 31 IVs, branching SPs (gen 10) vs EVs, appending As-One rider suffixes (`:625-629`), resolving species via `getExportSpecies` (`:900+`). Sidebar export loops 6 slots, sets `tempSet.name`, concatenates, copies. Import is the inverse parser (see custom-set feature).
- Inputs/outputs & data deps: `Pokemon` snapshot or sidebar customs; output clipboard text.
- Entry-points (UI/CLI/API): Per-panel Export button, sidebar Export-Team buttons.

### Legacy archived calculators/setdexes
- What it does: Frozen reference copies of older damage files and setdexes; not loaded by any page.
- Where it lives (`path` refs): `legacy_scripts/` (`damage_bw.js`, `damage_sm.js`, archived setdex/status files per explorer report).
- How it works: Reference only — no `<script>` tag points at them from `index.html:853-894`, `za-calc.html:697-720`, or `lgpe-calc.html:757-778`. Useful for diffing old formulas when porting rules.
- Inputs/outputs & data deps: None at runtime.
- Entry-points (UI/CLI/API): None (file browse only).

### Static asset packs
- What it does: All images and audio served locally except sidebar minisprites (hotlinked with local fallback).
- Where it lives (`path` refs): `image_res/` (UI, logos, favicons, `empty-slot.png`, `urshifu-single/rapid.png`, `minisprites/`, `aprilfoolsday/`); `sound_res/*.ogg` (wired in `script_res/statuscalc.js:24-32`).
- How it works: Sidebar sprites hotlink `https://www.smogon.com/forums/media/minisprites/<slug>.png` with on-error fallback to local `image_res/empty-slot.png` (`script_res/sidebars.js:135-143`); everything else is relative-path local.
- Inputs/outputs & data deps: None code-wise.
- Entry-points (UI/CLI/API): Rendered `<img>`/`<audio>` elements.

### April-Fools overlay (gen-9 dex injection)
- What it does: Date/localStorage-gated injection of joke characters into the live gen-9 pool.
- Where it lives (`path` refs): `script_res/statuscalc.js:3-42`; `afd_res/pokedex_afd24.js` (`POKEDEX_AFD24`); `afd_res/setdex_afd24.js` (`SETDEX_AFD24`).
- How it works: `checkaprilfools()` fires when `Date().toString()` contains `"Apr 01"` or `localStorage.isAFD` is set (`:4`); `loadrest()` injects `#spamtons/#noisechars` divs, `loadaudio()` injects 10 `<audio>` nodes from `sound_res/`, and `addjokechars()` copies every `POKEDEX_AFD24` key into `POKEDEX_SV` + `POKEDEX_SV_NATDEX` and pushes `SETDEX_AFD24` onto `COMPONENTS[9]` (`:35-42`). Variant pages pre-declare empty `COMPONENTS[9]` (`setdex_dynamic-lgpe-za.js:4`) specifically to avoid crashing on this push.
- Inputs/outputs & data deps: Date/localStorage; mutates live dex globals.
- Entry-points (UI/CLI/API): Automatic on boot; joke sets then appear in gen-9 set dropdown.

## Cross-feature notes
- Shared logic: Everything funnels through `calculate()` → `Pokemon`/`Field` snapshots → `calculateAllMoves` pointer → `GET_DAMAGE_*` → `getKOChanceText`. Gen switching is the master rewiring event (data tables, stat functions, calc entry, DOM visibility, setdex, sidebars). `setdex` merge order (curated → customs → sidebar reservations) is the single namespace all team features share — sidebar add/edit/delete and custom import/export are thin wrappers over the same `ALL_SETDEX_CUSTOM[gen]` + `localStorage` store.
- Gaps vs other repos: No server, build, tests, or API; compared with Honko-style upstream calcs this fork adds Champions/SPs (`CALC_HP/STAT_CHAMP`), Z-A cooldowns, VGC-curated NCP setdexes, sidebars-as-custom-sets, and the AFD overlay — but carries dead weight (`digicalc.html` dangling refs, `legacy_scripts/`, `script_res/css` stray file, duplicated NATDEX tables).
- Reimplementation hints: Replace global-name coupling (`gen/pokedex/moves/typeChart/abilities/items/STATS/calculateAllMoves/calcHP/calcStat/setdex/COMPONENTS/mechanicsTests`) with an explicit per-gen context object; split `GET_DAMAGE_SV`'s accreting `check*` preamble per gen; validate setdex shape at merge; split ZA cooldown (`cooldown_za.js`) and AFD injection (`statuscalc.js`) out of the hot calc path; fix or remove `digicalc.html` script refs; move sidebar sprites local-first instead of hotlink-with-fallback.
