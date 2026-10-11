> **Reference note (moved 2026-10-11):** This is a generated analysis report about the reference project cloned locally at `/Users/gabriel/Code/Playgrounds/vgc-calculator-projects/pokeapi` (PokeAPI — Django read-only reference-data API). Only this report was moved here — the full project source was intentionally NOT copied. Read the live source there when verifying claims.
> New app home: `~/Code/Projects/vgc/` (see `AGENTS.md`). Consolidated comparison: `reference-reports/CONSOLIDATED-FEATURE-REPORT.md`.

# Feature Report — pokeapi

> Sources: ARCHITECTURE-REPORT.md (inventory), EXPLORER-REPORT.md, direct reads. ARCHITECTURE-REPORT.md was present and used as the feature-discovery base; all claims below were spot-checked against code.

## Feature inventory
| Feature | One-line purpose | Key files |
|---|---|---|
| Ability resources | Abilities + effect/flavor/names/changelog + Pokémon holders | `pokemon_v2/api.py:199-202`, `pokemon_v2/serializers.py:1227-1255` |
| Berry resources | Berry growing/gift stats + flavor map + firmness/flavor links | `pokemon_v2/api.py:215-250`, `pokemon_v2/serializers.py:1731-1815` |
| Characteristic resources | Highest-IV characteristic (`gene_mod_5`) + computed possible values | `pokemon_v2/api.py:263-266`, `pokemon_v2/serializers.py:626-644` |
| Contest resources | Contest types/effects, appeal/jam, combos | `pokemon_v2/api.py:279-298`, `pokemon_v2/serializers.py:687-729` |
| Egg-group resources | Breeding groups + localized names + member species | `pokemon_v2/api.py:311-314`, `pokemon_v2/serializers.py:1831-1845` |
| Encounter resources | Encounter methods/conditions/values/slots vocab | `pokemon_v2/api.py:327-362`, `pokemon_v2/serializers.py:898-957` |
| Evolution-chain resources | Recursive family-tree builder with full evolution details | `pokemon_v2/api.py:375-378`, `pokemon_v2/serializers.py:3465-3560` |
| Evolution-trigger resources | Evolution event triggers + species requiring them | `pokemon_v2/api.py:391-394`, `pokemon_v2/serializers.py:3200-3221` |
| Evolution-variable resources | RPN condition variables/expressions for form-branching evos | `pokemon_v2/api.py:407-410`, `pokemon_v2/serializers.py:3233-3246,3445-3455` |
| Generation resources | Game generations + main region + version-group links | `pokemon_v2/api.py:423-426`, `pokemon_v2/serializers.py:767-803` |
| Gender resources | Gendered-species query by `gender_rate` + evolution-requirement query | `pokemon_v2/api.py:439-442`, `pokemon_v2/serializers.py:805-832` |
| Growth-rate resources | XP formula + per-level XP table + member species | `pokemon_v2/api.py:455-458`, `pokemon_v2/serializers.py:854-861` |
| Item resources | Items + categories/pockets/attributes/fling/prices/sprites/machines/holders | `pokemon_v2/api.py:471-538`, `pokemon_v2/serializers.py:1566-1646` |
| Currency resources | Purchase currencies for item prices | `pokemon_v2/api.py:551-554`, `pokemon_v2/serializers.py:1460-1479` |
| Language resources | Translation vocab (`iso639/iso3166/official` + local names) | `pokemon_v2/api.py:567-570`, `pokemon_v2/serializers.py:877-882` |
| Location resources | Locations + region + game indices + areas | `pokemon_v2/api.py:583-586`, `pokemon_v2/serializers.py:1157-1165` |
| Location-area resources | Areas + method-rate rollup + per-Pokémon encounter rollup | `pokemon_v2/api.py:599-602`, `pokemon_v2/serializers.py:1058-1138` |
| Machine resources | TM/HM item→move→version-group link | `pokemon_v2/api.py:615-618`, `pokemon_v2/serializers.py:2065-2072` |
| Move resources | Moves + meta/effects/changes/combos/machines/learners | `pokemon_v2/api.py:631-634`, `pokemon_v2/serializers.py:2338-2492` |
| Move-damage-class resources | physical/special/status + descriptions + moves | `pokemon_v2/api.py:647-650`, `pokemon_v2/serializers.py:2117-2129` |
| Move-ailment resources | Status-condition vocab + affected moves | `pokemon_v2/api.py:663-666`, `pokemon_v2/serializers.py:2146-2169` |
| Move-battle-style resources | Battle-Palace styles + nature preferences | `pokemon_v2/api.py:679-682`, `pokemon_v2/serializers.py:2088-2093` |
| Move-category resources | Move-effect categories + affected moves | `pokemon_v2/api.py:695-698`, `pokemon_v2/serializers.py:2171-2200` |
| Move-learn-method resources | level/machine/egg/tutor methods + version groups | `pokemon_v2/api.py:711-714`, `pokemon_v2/serializers.py:2755-2776` |
| Move-target resources | Battle targets + descriptions + moves | `pokemon_v2/api.py:727-730`, `pokemon_v2/serializers.py:2233-2262` |
| Nature resources | Stat modifiers + flavor prefs + pokeathlon/battle-style links | `pokemon_v2/api.py:743-746`, `pokemon_v2/serializers.py:1682-1715` |
| Pal-park-area resources | Pal-Park areas + species encounter rates | `pokemon_v2/api.py:759-762`, `pokemon_v2/serializers.py:2494-2509` |
| Pokeathlon-stat resources | Pokéathlon stats + affecting natures split | `pokemon_v2/api.py:775-778`, `pokemon_v2/serializers.py:3589-3611` |
| Pokedex resources | Dexes + entries + descriptions + version groups | `pokemon_v2/api.py:791-794`, `pokemon_v2/serializers.py:3644-3682` |
| Pokemon-color resources | Dex color buckets + member species | `pokemon_v2/api.py:807-810`, `pokemon_v2/serializers.py:2525-2531` |
| Pokemon-form resources | Cosmetic forms + sprites/types/conditions/flavor | `pokemon_v2/api.py:823-826`, `pokemon_v2/serializers.py:2612-2723` |
| Pokemon-habitat resources | Terrain habitats + member species | `pokemon_v2/api.py:839-842`, `pokemon_v2/serializers.py:2725-2753` |
| Pokemon-shape resources | Dex shape buckets + awesome/localized names | `pokemon_v2/api.py:855-858`, `pokemon_v2/serializers.py:2778-2815` |
| Pokemon resources | Core Pokémon: abilities/moves/stats/types/sprites/cries/items | `pokemon_v2/api.py:871-874`, `pokemon_v2/serializers.py:2932-3184` |
| Pokemon-species resources | Species: genera/flavor/varieties/evolution/pal-park | `pokemon_v2/api.py:887-890`, `pokemon_v2/serializers.py:3273-3375` |
| Region resources | Regions + dexes/generations/games/version-group backlinks | `pokemon_v2/api.py:903-906`, `pokemon_v2/serializers.py:726-765` |
| Stat resources | Battle stats + affecting moves/natures/items, vitamin/X-item map | `pokemon_v2/api.py:919-922`, `pokemon_v2/serializers.py:1289-1361` |
| Super-contest-effect resources | Super-contest appeal/flavor + moves | `pokemon_v2/api.py:935-938`, `pokemon_v2/serializers.py:660-668` |
| Type resources | Types + 6-way damage relations + past-gen overlays + sprites | `pokemon_v2/api.py:951-954`, `pokemon_v2/serializers.py:1905-2057` |
| Version resources | Game versions + names + version-group link | `pokemon_v2/api.py:967-970`, `pokemon_v2/serializers.py:3698-3710` |
| Version-group resources | Grouped releases + regions/learn-methods/pokedexes | `pokemon_v2/api.py:983-986`, `pokemon_v2/serializers.py:3713-3760` |
| Pokemon-encounters sub-resource | Per-Pokémon location encounters, id-only route | `pokemon_v2/urls.py:101-105`, `pokemon_v2/api.py:1019-1068` |
| API meta endpoint | Git deploy hash/date/tag | `pokemon_v2/urls.py:99`, `pokemon_v2/api.py:1083-1115` |
| Name-or-ID lookup | Dual pk/name retrieval on all common viewsets | `pokemon_v2/api.py:117-157` |
| Local `?q=` name search | `name__icontains` list filter, local-only | `pokemon_v2/api.py:126-134,160-165` |
| Summary/Detail duality | List `{name,url}` stubs vs full detail graphs | `pokemon_v2/api.py:99-114`, `pokemon_v2/serializers.py:277-589` |
| Sprite asset pipeline | Manifest/local-submodule sprite URL synthesis | `data/v2/build.py:59-147,157-775,778-781` |
| Cry asset pipeline | Cries-submodule sound URL synthesis | `data/v2/build.py:148-154,778-781` |
| CSV ETL build | Wipe-and-reload `build_all()` bulk pipeline, 26 stages | `data/v2/build.py:796-828,2477-2503` |
| GraphQL sidecar | Hasura engine over same Postgres | `docker-compose.yml:50-75`, `Resources/nginx/nginx.conf:120-139` |
| OpenAPI schema | drf-spectacular generated `openapi.yml` | `config/settings.py:153-154,175-280`, `Makefile:59-60` |
| Multi-layer cache | cachalot + Redis + nginx edge cache + gzip | `config/settings.py:117-140`, `Resources/nginx/nginx.conf:34-69,141-153` |
| Environment overlays | local/docker/compose settings variants | `config/local.py:1-22`, `config/docker.py`, `config/docker_compose.py:8-29` |
| Container/k8s deployment | compose stacks, nginx, kustomize, gunicorn | `docker-compose.yml`, `Resources/`, `gunicorn.conf.py:1-6` |

## Feature details
### Ability resources
- What it does: Serves passive battle/overworld effects; detail carries localized effect/flavor/names, generation, `is_main_series`, per-version-group effect changes, and holder Pokémon with hidden/slot flags.
- Where it lives (`path` refs): viewset `pokemon_v2/api.py:199-202`; serializers `pokemon_v2/serializers.py:1173-1255`; registered `pokemon_v2/urls.py:40`.
- How it works: `AbilityResource.queryset = Ability.objects.select_related("generation")`; `AbilityDetailSerializer` nests `abilityeffecttext/abilityflavortext/abilityname/abilitychange` reverse relations (`pokemon_v2/serializers.py:1228-1232`) and `get_ability_pokemon()` queries `PokemonAbility` → `AbilityPokemonDetailSerializer` (`pokemon_v2/serializers.py:1249-1255`).
- Inputs/outputs & data deps: In: ability id/name. Out: full ability graph. Deps: `Ability`, `AbilityEffectText/FlavorText/Name/Change`, `PokemonAbility`, `Generation`, `Language`, `VersionGroup`.
- Entry-points (UI/CLI/API): `GET /api/v2/ability/`, `GET /api/v2/ability/{id|name}/`.

### Berry resources
- What it does: Three endpoints — `berry` (growth_time/max_harvest/natural_gift/size/smoothness/soil_dryness + item/type/firmness/flavors), `berry-firmness` (member berries), `berry-flavor` (potency-ranked berries + contest type).
- Where it lives: `pokemon_v2/api.py:215-250`; `pokemon_v2/serializers.py:1723-1815`; routes `pokemon_v2/urls.py:41-43`.
- How it works: `BerryDetailSerializer.get_berry_flavors()` joins `BerryFlavorMap→berry_flavor` (`pokemon_v2/serializers.py:1809-1815`); `BerryFlavorDetailSerializer.get_berries_with_flavor()` filters `potency__gt=0`, orders by potency (`pokemon_v2/serializers.py:1765-1775`); firmness nests `berry` reverse set (`pokemon_v2/serializers.py:1731-1737`).
- Inputs/outputs & data deps: Berry/flavor/firmness id-or-name; outputs flavor-potency maps; deps `Berry/BerryFlavorMap/BerryFirmness`, `Item`, `Type`, `ContestType`.
- Entry-points: `GET /api/v2/berry*/`.

### Characteristic resources
- What it does: Highest-IV indicator; `gene_modulo` (=`gene_mod_5`) plus computed `possible_values` and `highest_stat`, with localized descriptions.
- Where it lives: `pokemon_v2/api.py:263-266`; `pokemon_v2/serializers.py:626-644`; only summary is `{url}` (`pokemon_v2/serializers.py:301-304`).
- How it works: Formula `list(range(obj.gene_mod_5, 32, 5))` in `get_values` (`pokemon_v2/serializers.py:642-644`) — e.g. modulo 0 → `[0,5,…,30]`; `highest_stat` via `source="stat"` (`pokemon_v2/serializers.py:628-629`).
- Inputs/outputs & data deps: Numeric id only (no name field); deps `Characteristic/CharacteristicDescription`, `Stat`.
- Entry-points: `GET /api/v2/characteristic/[{id}/]`.

### Contest resources
- What it does: `contest-type` (names incl. color + berry-flavor/move links), `contest-effect` (appeal/jam + effect/flavor texts), `super-contest-effect` (appeal + flavor + moves).
- Where it lives: `pokemon_v2/api.py:279-298,935-938`; `pokemon_v2/serializers.py:660-729`; routes `pokemon_v2/urls.py:45-46,86`.
- How it works: Straight reverse-FK nesting (`contesteffecteffecttext`, `contesttypename`, `supercontesteffectflavortext`, `move`) with no computed groupbys (`pokemon_v2/serializers.py:660-729`); appeal/jam are stored ints.
- Inputs/outputs & data deps: Contest ids; deps `ContestType/ContestEffect/SuperContestEffect`, `Move`, `BerryFlavor`, `Language`.
- Entry-points: `GET /api/v2/contest-type|contest-effect|super-contest-effect/`.

### Egg-group resources
- What it does: Breeding compatibility groups with localized names and member species list.
- Where it lives: `pokemon_v2/api.py:311-314`; `pokemon_v2/serializers.py:1831-1845`; route `pokemon_v2/urls.py:47`.
- How it works: `get_species()` filters `PokemonSpecies.objects.filter(pokemonegggroup__egg_group=obj)` (`pokemon_v2/serializers.py:1840-1845`).
- Inputs/outputs & data deps: Egg-group id/name → species summaries; deps `EggGroup/EggGroupName/PokemonEggGroup`.
- Entry-points: `GET /api/v2/egg-group/`.

### Encounter resources
- What it does: Wild-encounter vocab: `encounter-method` (order + names), `encounter-condition` (values + names), `encounter-condition-value` (parent condition + names); slot model renames `rarity→chance`.
- Where it lives: `pokemon_v2/api.py:327-362`; `pokemon_v2/serializers.py:890-957`; routes `pokemon_v2/urls.py:48-50`.
- How it works: Direct FK/reverse nesting; `EncounterSlotSerializer.chance = IntegerField(source="rarity")` (`pokemon_v2/serializers.py:940-946`); `EncounterDetailSerializer` adds `condition_values` via `SerializerMethodField` plus min/max level, method, slot (`pokemon_v2/serializers.py:960-1006`).
- Inputs/outputs & data deps: Encounter vocab ids/names; deps `Encounter/EncounterSlot/EncounterMethod/EncounterCondition(Value/Map)`.
- Entry-points: `GET /api/v2/encounter-method|encounter-condition|encounter-condition-value/`.

### Evolution-chain resources
- What it does: Full family tree: `{is_baby, species, evolution_details[], evolves_to[]}` recursively built, with `baby_trigger_item` at root.
- Where it lives: `pokemon_v2/api.py:375-378`; `pokemon_v2/serializers.py:3378-3560`; route `pokemon_v2/urls.py:51`.
- How it works: `build_chain()` loads species ordered by `order,pk`, builds `{species,evolves_from_species,is_baby}` ref data, prefetches all `PokemonEvolution` with every FK (`POKEMON_EVOLUTION_FK_FIELDS`, `pokemon_v2/serializers.py:3469-3498`), constructs tree via stack search on `evolves_from_species` (`pokemon_v2/serializers.py:3504-3533`), then recursively serializes with `PokemonEvolutionSerializer` (`pokemon_v2/serializers.py:3537-3560`). `PokemonEvolutionSerializer` also decodes `nature_bitmask` → nature list (`pokemon_v2/serializers.py:3432-3443`) and RPN `condition_expression` + referenced variables (`pokemon_v2/serializers.py:3445-3455`).
- Inputs/outputs & data deps: Chain id; deps `EvolutionChain`, `PokemonSpecies`, `PokemonEvolution` + ~15 FK tables (item/move/type/location/region/form/trigger/nature).
- Entry-points: `GET /api/v2/evolution-chain/{id}/`.

### Evolution-trigger resources
- What it does: Evolution events (level-up, trade, use-item…) with localized names and species that evolve by them.
- Where it lives: `pokemon_v2/api.py:391-394`; `pokemon_v2/serializers.py:3200-3221`; route `pokemon_v2/urls.py:52`.
- How it works: `get_species()` = `PokemonSpecies.objects.filter(evolved_species__evolution_trigger=obj).distinct().order_by("id")` (`pokemon_v2/serializers.py:3208-3214`); names via `evolutiontriggername` reverse set.
- Inputs/outputs & data deps: Trigger id/name; deps `EvolutionTrigger(Name)`, `PokemonEvolution`.
- Entry-points: `GET /api/v2/evolution-trigger/`.

### Evolution-variable resources
- What it does: Exposes RPN variables (e.g. EC/Personality-Value symbols) with descriptions, consumed by `PokemonEvolution.condition_expression`.
- Where it lives: `pokemon_v2/api.py:407-410`; `pokemon_v2/serializers.py:3233-3246`; models `pokemon_v2/models.py:1801-1818`; route `pokemon_v2/urls.py:53`.
- How it works: Plain detail (names+descriptions); reverse usage in `PokemonEvolutionSerializer.get_condition_expression()` which tokenizes `expression.split()`, looks up `EvolutionVariable(symbol__in=tokens)`, returns `{expression, percentage_chance, variables}` (`pokemon_v2/serializers.py:3445-3455`); `EvolutionConditionExpressionSerializer` documents the RPN shape (`pokemon_v2/serializers.py:355-366`).
- Inputs/outputs & data deps: Variable id/name; deps `EvolutionVariable(Name/Description)`.
- Entry-points: `GET /api/v2/evolution-variable/`.

### Generation resources
- What it does: Game generation with main region, localized names, and backlinks to abilities/moves/species/stats/types/version groups.
- Where it lives: `pokemon_v2/api.py:423-426`; `pokemon_v2/serializers.py:767-803`; route `pokemon_v2/urls.py:54`.
- How it works: Reverse-FK summary lists (`ability`, `move`, `pokemonspecies`, `stat`, `type`, `versiongroup`) plus `region` FK (`pokemon_v2/serializers.py:767-803`).
- Inputs/outputs & data deps: Generation id/name; deps `Generation/GenerationName`, `Region`.
- Entry-points: `GET /api/v2/generation/`.

### Gender resources
- What it does: Three genders; detail computes gendered species from `gender_rate` and species requiring that gender to evolve.
- Where it lives: `pokemon_v2/api.py:439-442`; `pokemon_v2/serializers.py:805-832`; route `pokemon_v2/urls.py:55`.
- How it works: Key formulas in `get_species`: female=`gender_rate>0`, male=`0..7`, genderless=`-1` (`pokemon_v2/serializers.py:815-824`); `get_required`: `PokemonSpecies.filter(evolved_species__gender=obj)` (`pokemon_v2/serializers.py:826-832`).
- Inputs/outputs & data deps: Gender id/name; deps `Gender`, `PokemonSpecies(gender_rate)`, `PokemonEvolution.gender`.
- Entry-points: `GET /api/v2/gender/`.

### Growth-rate resources
- What it does: Leveling speed: stored `formula` string + full `{level, experience}` table + member species.
- Where it lives: `pokemon_v2/api.py:455-458`; `pokemon_v2/serializers.py:840-861`; route `pokemon_v2/urls.py:56`.
- How it works: Nests `growthratedescription` and `experience` reverse sets (`pokemon_v2/serializers.py:854-857`); formula is opaque text from CSV, not evaluated in Python.
- Inputs/outputs & data deps: Growth-rate id/name; deps `GrowthRate/Description`, `Experience`.
- Entry-points: `GET /api/v2/growth-rate/`.

### Item resources
- What it does: Five endpoints: `item` (effect/flavor/prices/game-indices/category/fling/sprites/machines/holders/baby-trigger), `item-category`/`item-pocket` (bag taxonomy), `item-attribute` (usable-in-battle etc.), `item-fling-effect` (Fling texts), `currency` (price vocab).
- Where it lives: `pokemon_v2/api.py:471-554`; `pokemon_v2/serializers.py:1377-1564`; routes `pokemon_v2/urls.py:57-62`.
- How it works: `ItemDetailSerializer` mixes reverse nesting (`itemname/itemgameindex/itemprice/itemeffecttext/itemflavortext`) with computed fields: `get_item_attributes` via `ItemAttributeMap` (`pokemon_v2/serializers.py:1613-1619`), `get_held_by_pokemon` groups `PokemonItem` by Pokémon with `itertools.groupby` (`pokemon_v2/serializers.py:1621-1638`), `get_item_machines` queries `Machine(item=obj)` (`pokemon_v2/serializers.py:1600-1606`), `get_item_sprites` reads `ItemSprites` JSON (`pokemon_v2/serializers.py:1608-1611`), `get_baby_trigger_for` looks up `EvolutionChain(baby_trigger_item=obj)` (`pokemon_v2/serializers.py:1640-1646`).
- Inputs/outputs & data deps: Item-related id/name; deps `Item*` tables, `Machine`, `PokemonItem`, `EvolutionChain`, `ItemSprites`.
- Entry-points: `GET /api/v2/item|item-category|item-attribute|item-fling-effect|item-pocket|currency/`.

### Currency resources
- What it does: Tiny vocab (e.g. Pokédollars) backing `ItemPrice.currency`; detail lists local names.
- Where it lives: `pokemon_v2/api.py:551-554`; `pokemon_v2/serializers.py:1460-1479`; route `pokemon_v2/urls.py:62`.
- How it works: Direct model + `currencyname` nesting; no computed fields.
- Inputs/outputs & data deps: Currency id/name; deps `Currency/CurrencyName`.
- Entry-points: `GET /api/v2/currency/`.

### Language resources
- What it does: Translation backbone for every `*_name/prose/flavor` table: `{id,name,official,iso639,iso3166,names[]}`.
- Where it lives: `pokemon_v2/api.py:567-570`; `pokemon_v2/serializers.py:877-882`; route `pokemon_v2/urls.py:63`.
- How it works: `names` nests `languagename_language` (self-referential local-language names) (`pokemon_v2/serializers.py:877-882`); every other detail serializer embeds `LanguageSummarySerializer` for its language FKs.
- Inputs/outputs & data deps: Language id/name; deps `Language/LanguageName`.
- Entry-points: `GET /api/v2/language/`.

### Location resources
- What it does: `location` (city/route + region + game indices + sub-areas + names) — parent of location-areas.
- Where it lives: `pokemon_v2/api.py:583-586`; `pokemon_v2/serializers.py:1157-1165`; route `pokemon_v2/urls.py:64`.
- How it works: Nests `locationname/locationgameindex` and `locationarea` reverse sets plus nullable `region` FK (`pokemon_v2/serializers.py:1157-1165`).
- Inputs/outputs & data deps: Location id/name; deps `Location(Name/GameIndex)`, `LocationArea`, `Region`, `Generation`.
- Entry-points: `GET /api/v2/location/`.

### Location-area resources
- What it does: Sub-area (floor/cave section) with per-method encounter rates and per-Pokémon encounter rollups. Note: the only viewset that does NOT support name lookup (extends `ListOrDetailSerialRelation + ReadOnlyModelViewSet`, no `NameOrIdRetrieval`, `pokemon_v2/api.py:599-602`).
- Where it lives: `pokemon_v2/api.py:599-602`; `pokemon_v2/serializers.py:1008-1138`; route `pokemon_v2/urls.py:65`.
- How it works: `get_method_rates()` groups `LocationAreaEncounterRate` by method via `itertools.groupby` (`pokemon_v2/serializers.py:1076-1094`); `get_encounters()` filters `Encounter(location_area=obj)`, orders by `pokemon,version,slot`, double-groupbys Pokémon→version computing `max_chance = sum(slot.rarity)` (`pokemon_v2/serializers.py:1096-1138`).
- Inputs/outputs & data deps: Area id (numeric path); deps `LocationArea`, `LocationAreaEncounterRate`, `Encounter(+Slot/ConditionValueMap/Detail)`.
- Entry-points: `GET /api/v2/location-area/[{id}/]`.

### Machine resources
- What it does: Join of TM/HM item + taught move + version group.
- Where it lives: `pokemon_v2/api.py:615-618`; `pokemon_v2/serializers.py:2065-2072`; route `pokemon_v2/urls.py:66`.
- How it works: Three summary FKs (`item`, `version_group`, `move`), no nesting (`pokemon_v2/serializers.py:2065-2072`); reverse-consumed by `MoveDetailSerializer.get_move_machines` and `ItemDetailSerializer.get_item_machines`.
- Inputs/outputs & data deps: Machine id; deps `Machine`, `Item`, `Move`, `VersionGroup`.
- Entry-points: `GET /api/v2/machine/{id}/`.

### Move resources
- What it does: Richest combat endpoint: power/pp/accuracy/priority, type/target/damage-class/contest links, `meta` (ailment/category/flinch/stat-chance…), localized effect/flavor/names, `past_values` (type/power/pp/accuracy/effect per version group), contest/super-contest combos, machines teaching it, and all learner Pokémon.
- Where it lives: `pokemon_v2/api.py:631-634`; `pokemon_v2/serializers.py:2338-2492`; route `pokemon_v2/urls.py:67`.
- How it works: `get_effect_text/get_effect_change_text` dereference shared `MoveEffect` table (`pokemon_v2/serializers.py:2448-2462`); `get_combos` unions `ContestCombo` + `SuperContestCombo` into `{normal,super}×{use_before,use_after}` (`pokemon_v2/serializers.py:2403-2446`); `get_move_machines` and `get_learned_by_pokemon` query `Machine`/`PokemonMove` (`pokemon_v2/serializers.py:2387-2401`); `past_values` nests `movechange` incl. old type/effect (`pokemon_v2/serializers.py:2352`).
- Inputs/outputs & data deps: Move id/name; deps `Move/MoveMeta/MoveChange/MoveEffect*`, `Type`, `ContestType/Effect`, `Machine`, `PokemonMove`.
- Entry-points: `GET /api/v2/move/`.

### Move-damage-class resources
- What it does: physical/special/status vocab with descriptions and member moves.
- Where it lives: `pokemon_v2/api.py:647-650`; `pokemon_v2/serializers.py:2117-2129`; route `pokemon_v2/urls.py:71`.
- How it works: Nests `movedamageclassname/movedamageclassdescription` + `move` reverse set (`pokemon_v2/serializers.py:2117-2129`).
- Inputs/outputs & data deps: Damage-class id/name; deps `MoveDamageClass(Name/Description)`, `Move`, `Stat`, `Type`.
- Entry-points: `GET /api/v2/move-damage-class/`.

### Move-ailment resources
- What it does: Status-condition vocab (paralysis, burn…) with names and affected moves.
- Where it lives: `pokemon_v2/api.py:663-666`; `pokemon_v2/serializers.py:2146-2169`; route `pokemon_v2/urls.py:68`.
- How it works: Nests `movemetaailmentname` + `movemeta` reverse set (`pokemon_v2/serializers.py:2146-2169`).
- Inputs/outputs & data deps: Ailment id/name; deps `MoveMetaAilment(Name)`, `MoveMeta`.
- Entry-points: `GET /api/v2/move-ailment/`.

### Move-battle-style resources
- What it does: Battle-Palace move styles; consumed by nature battle-style preferences.
- Where it lives: `pokemon_v2/api.py:679-682`; `pokemon_v2/serializers.py:2088-2093`; route `pokemon_v2/urls.py:69`.
- How it works: Names-only detail (`movebattlestylename`, `pokemon_v2/serializers.py:2088-2093`).
- Inputs/outputs & data deps: Style id/name; deps `MoveBattleStyle(Name)`, `NatureBattleStylePreference`.
- Entry-points: `GET /api/v2/move-battle-style/`.

### Move-category resources
- What it does: General move-effect categories (damage/ailment/net-good-stats…) with descriptions and moves.
- Where it lives: `pokemon_v2/api.py:695-698`; `pokemon_v2/serializers.py:2171-2200`; route `pokemon_v2/urls.py:70`.
- How it works: Nests descriptions + `movemeta` reverse set.
- Inputs/outputs & data deps: Category id/name; deps `MoveMetaCategory(Description)`, `MoveMeta`.
- Entry-points: `GET /api/v2/move-category/`.

### Move-learn-method resources
- What it does: How Pokémon learn moves (level-up/machine/egg/tutor/stadium-surprise…) with descriptions and version groups.
- Where it lives: `pokemon_v2/api.py:711-714`; `pokemon_v2/serializers.py:2755-2776`; route `pokemon_v2/urls.py:72`.
- How it works: Nests names/descriptions + `versiongroup` reverse set; reverse-consumed by Pokémon move grouping (`pokemon_v2/serializers.py:2755-2776`).
- Inputs/outputs & data deps: Method id/name; deps `MoveLearnMethod(Name/Description)`, `VersionGroupMoveLearnMethod`.
- Entry-points: `GET /api/v2/move-learn-method/`.

### Move-target resources
- What it does: Battle targets (selected-pokemon, all-opponents, user's-field…) with descriptions and moves.
- Where it lives: `pokemon_v2/api.py:727-730`; `pokemon_v2/serializers.py:2233-2262`; route `pokemon_v2/urls.py:73`.
- How it works: Nests names/descriptions + `move` reverse set.
- Inputs/outputs & data deps: Target id/name; deps `MoveTarget(Name/Description)`, `Move`.
- Entry-points: `GET /api/v2/move-target/`.

### Nature resources
- What it does: Stat-growth natures: ±stats, liked/hated berry flavors, preferred berries, Pokéathlon deltas, battle-style preferences.
- Where it lives: `pokemon_v2/api.py:743-746`; `pokemon_v2/serializers.py:1682-1715`; route `pokemon_v2/urls.py:74`.
- How it works: Direct FKs (`increased_stat/decreased_stat/likes_flavor/hates_flavor`) + reverse `berry` + `get_pokeathlon_stats` via `NaturePokeathlonStat` (`pokemon_v2/serializers.py:1709-1715`) + `naturebattlestylepreference` nesting.
- Inputs/outputs & data deps: Nature id/name; deps `Nature(Name/BattleStylePreference/PokeathlonStat)`, `Stat`, `BerryFlavor`, `Berry`.
- Entry-points: `GET /api/v2/nature/`.

### Pal-park-area resources
- What it does: Pal-Park zones with `{base_score, rate, pokemon_species}` encounters.
- Where it lives: `pokemon_v2/api.py:759-762`; `pokemon_v2/serializers.py:2494-2509`; route `pokemon_v2/urls.py:75`.
- How it works: `get_encounters()` = `PalPark.objects.filter(pal_park_area=obj).select_related("pokemon_species")` → `PalParkEncounterSerializer` (`pokemon_v2/serializers.py:2502-2509`); also surfaced per-species via `PokemonSpeciesDetailSerializer.get_encounters` (`pokemon_v2/serializers.py:3369-3375`).
- Inputs/outputs & data deps: Area id/name; deps `PalParkArea(Name)`, `PalPark`, `PokemonSpecies`.
- Entry-points: `GET /api/v2/pal-park-area/`.

### Pokeathlon-stat resources
- What it does: Pokéathlon performance stats (speed/power/skill/stamina/jump) with affecting natures split into increase/decrease.
- Where it lives: `pokemon_v2/api.py:775-778`; `pokemon_v2/serializers.py:3589-3611`; route `pokemon_v2/urls.py:83`.
- How it works: `get_natures_that_affect` splits `NaturePokeathlonStat` on `max_change>0` vs `<=0` (`pokemon_v2/serializers.py:3597-3611`).
- Inputs/outputs & data deps: Stat id/name; deps `PokeathlonStat(Name)`, `NaturePokeathlonStat`, `Nature`.
- Entry-points: `GET /api/v2/pokeathlon-stat/`.

### Pokedex resources
- What it does: Regional/national dexes: descriptions, ordered `{entry_number, pokemon_species}` entries, and version groups.
- Where it lives: `pokemon_v2/api.py:791-794`; `pokemon_v2/serializers.py:3644-3682`; route `pokemon_v2/urls.py:76`.
- How it works: `get_pokedex_entries` orders `PokemonDexNumber` by `pokedex_number` (`pokemon_v2/serializers.py:3664-3674`); `get_pokedex_version_groups` via `PokedexVersionGroup` join (`pokemon_v2/serializers.py:3676-3682`).
- Inputs/outputs & data deps: Pokédex id/name; deps `Pokedex(Name/Description/VersionGroup)`, `PokemonDexNumber`.
- Entry-points: `GET /api/v2/pokedex/`.

### Pokemon-color resources
- What it does: Sort-by-color buckets (red/blue/…) listing member species.
- Where it lives: `pokemon_v2/api.py:807-810`; `pokemon_v2/serializers.py:2525-2531`; route `pokemon_v2/urls.py:78`.
- How it works: Nests `pokemoncolorname` + `pokemonspecies` reverse set.
- Inputs/outputs & data deps: Color id/name; deps `PokemonColor(Name)`, `PokemonSpecies`.
- Entry-points: `GET /api/v2/pokemon-color/`.

### Pokemon-form resources
- What it does: Cosmetic forms (Alolan display forms, Mega…): sprites, types, version group, names, form-switch conditions (trigger/item/ability/move/base_form).
- Where it lives: `pokemon_v2/api.py:823-826`; `pokemon_v2/serializers.py:2539-2723`; route `pokemon_v2/urls.py:79`.
- How it works: `PokemonFormDetailSerializer` nests `PokemonFormSprites`, `PokemonFormType`, names; conditions via `PokemonFormCondition` (`trigger/item/ability/move/base_form`, `pokemon_v2/serializers.py:2545-2554`); referenced from `PokemonDetailSerializer.forms` (`pokemon_v2/serializers.py:2942`).
- Inputs/outputs & data deps: Form id/name; deps `PokemonForm(Name/Sprites/Type/Condition/FlavorText)`, `Pokemon`.
- Entry-points: `GET /api/v2/pokemon-form/`.

### Pokemon-habitat resources
- What it does: Terrain habitats (cave/forest/grassland…) listing member species.
- Where it lives: `pokemon_v2/api.py:839-842`; `pokemon_v2/serializers.py:2725-2753`; route `pokemon_v2/urls.py:80`.
- How it works: Nests names + `pokemonspecies` reverse set.
- Inputs/outputs & data deps: Habitat id/name; deps `PokemonHabitat(Name)`, `PokemonSpecies`.
- Entry-points: `GET /api/v2/pokemon-habitat/`.

### Pokemon-shape resources
- What it does: Pokédex shape buckets (ball/no-legs/wings…) with localized names + `awesome_name`s + member species.
- Where it lives: `pokemon_v2/api.py:855-858`; `pokemon_v2/serializers.py:2786-2815`; route `pokemon_v2/urls.py:81`.
- How it works: Nests `pokemonshapename` twice (names + awesome_names) plus `pokemonspecies` (`pokemon_v2/serializers.py:2586-2600,2778-2815`).
- Inputs/outputs & data deps: Shape id/name; deps `PokemonShape(Name)`, `PokemonSpecies`.
- Entry-points: `GET /api/v2/pokemon-shape/`.

### Pokemon resources
- What it does: Core creature-variant endpoint: base stats, height/weight/XP, abilities (+past), moves grouped by move with per-version-group learn details, held items grouped by item, types (+past), stats (+past), forms, game indices, species link, sprite/cry JSON blobs, and an encounters URL.
- Where it lives: `pokemon_v2/api.py:871-874`; `pokemon_v2/serializers.py:2932-3184`; route `pokemon_v2/urls.py:77`.
- How it works: Viewset prefetches `pokemon_species/pokemonsprites/pokemoncries` (`pokemon_v2/api.py:872`); `get_pokemon_sprites` reads prefetched JSON with all-female (`gender_rate==8`) fallback fill (`pokemon_v2/serializers.py:2981-2998`); `get_pokemon_cries` reads JSON (`pokemon_v2/serializers.py:2999-3002`); `get_pokemon_moves` groups `PokemonMove` by move with `vg_cache/mlm_cache` dedup (`pokemon_v2/serializers.py:3004-3046`); `get_pokemon_held_items` groups `PokemonItem` by item (`pokemon_v2/serializers.py:3048-3081`); past abilities/stats/types group `*Past` rows by generation via `itertools.groupby` (`pokemon_v2/serializers.py:3091-3180`); `get_encounters` reverses the `pokemon_encounters` sub-resource URL (`pokemon_v2/serializers.py:3182-3184`).
- Inputs/outputs & data deps: Pokémon id/name; deps `Pokemon/Sprites/Cries/Move/Item/Ability/Stat/Type(+Past)`, `PokemonSpecies`, `VersionGroup`, `MoveLearnMethod`.
- Entry-points: `GET /api/v2/pokemon/[{id|name}/]`; detail links to `/api/v2/pokemon/{id}/encounters`.

### Pokemon-species resources
- What it does: Species-level record shared across varieties: dex numbers, genera/names/flavor, egg groups, gender/capture/happiness flags, growth/color/shape/habitat/generation, `evolves_from`, evolution chain, varieties (`{is_default, pokemon}`), and Pal-Park encounters.
- Where it lives: `pokemon_v2/api.py:887-890` (explicit `.order_by("id")`); `pokemon_v2/serializers.py:3273-3375`; route `pokemon_v2/urls.py:82`.
- How it works: `get_pokemon_names/genera` query `PokemonSpeciesName` (genera = non-empty genus rows, `pokemon_v2/serializers.py:3326-3344`); `get_pokemon_egg_groups` via `PokemonEggGroup` join (`pokemon_v2/serializers.py:3346-3352`); `get_pokemon_varieties` zips `Pokemon` rows with summaries preserving `is_default` (`pokemon_v2/serializers.py:3354-3367`); `get_encounters` via `PalPark` (`pokemon_v2/serializers.py:3369-3375`).
- Inputs/outputs & data deps: Species id/name; deps `PokemonSpecies(Name/Description/FlavorText/DexNumber)`, `Pokemon`, `EggGroup`, `EvolutionChain`, `PalPark`, `Generation/GrowthRate/Color/Shape/Habitat`.
- Entry-points: `GET /api/v2/pokemon-species/`.

### Region resources
- What it does: World regions with localized names, dexes, generations, games, and version-group backlinks.
- Where it lives: `pokemon_v2/api.py:903-906`; `pokemon_v2/serializers.py:726-765`; route `pokemon_v2/urls.py:84`.
- How it works: Reverse summary lists (`pokedex`, `generation`, `versiongroup` via `versiongroupregion`, `pokemon_species` games) + names (`pokemon_v2/serializers.py:726-765`).
- Inputs/outputs & data deps: Region id/name; deps `Region(Name)`, `Pokedex`, `Generation`, `VersionGroupRegion`.
- Entry-points: `GET /api/v2/region/`.

### Stat resources
- What it does: Battle stats with affecting moves (split increase/decrease), affecting natures (split), affecting items (hardcoded vitamin/X-item map), characteristics, damage class, names.
- Where it lives: `pokemon_v2/api.py:919-922`; `pokemon_v2/serializers.py:1289-1361`; route `pokemon_v2/urls.py:85`.
- How it works: `get_moves_that_affect` splits `MoveMetaStatChange` on `change>0` vs `<=0` (`pokemon_v2/serializers.py:1312-1321`); `get_natures_that_affect` queries `increased_stat/decreased_stat` (`pokemon_v2/serializers.py:1323-1331`); `get_items_that_affect` uses hardcoded `stat_item_mapping` (hp→hp-up; attack→protein/x-attack; … accuracy→x-accuracy; evasion→x-evasion) then `Item(name__in=…)` (`pokemon_v2/serializers.py:1333-1361`).
- Inputs/outputs & data deps: Stat id/name; deps `Stat(Name)`, `MoveMetaStatChange`, `Nature`, `Item`, `Characteristic`.
- Entry-points: `GET /api/v2/stat/`.

### Super-contest-effect resources
- What it does: Super-contest effects: appeal value + flavor texts + moves using them.
- Where it lives: `pokemon_v2/api.py:935-938`; `pokemon_v2/serializers.py:660-668`; route `pokemon_v2/urls.py:86`.
- How it works: Nests `supercontesteffectflavortext` + `move` reverse set; appeal is a stored int.
- Inputs/outputs & data deps: Effect id; deps `SuperContestEffect(FlavorText)`, `Move`.
- Entry-points: `GET /api/v2/super-contest-effect/`.

### Type resources
- What it does: Battle types with 6-way current damage relations, per-generation past overlays, game indices, generation, damage class, names, member Pokémon/moves, and sprite JSON.
- Where it lives: `pokemon_v2/api.py:951-954`; `pokemon_v2/serializers.py:1905-2057`; route `pokemon_v2/urls.py:87`.
- How it works: `get_type_relationships` pivots `TypeEfficacy` rows via `FACTOR_PREFIX_MAP {200:double, 50:half, 0:no}` into `*_damage_{to,from}` (`pokemon_v2/serializers.py:1920-1977`); `get_type_past_relationships` replays `TypeEfficacyPast` per generation: starts from current relations, drops unintroduced types (`name_to_gen_pk <= current_gen.pk`), then remove+re-add offensive/defensive entries (`pokemon_v2/serializers.py:1990-2049`); `get_type_pokemon` via `PokemonType` (`pokemon_v2/serializers.py:2051-2057`); sprites from `TypeSprites` (`pokemon_v2/serializers.py:1946-1949`).
- Inputs/outputs & data deps: Type id/name; deps `Type(Efficacy(Past)/Name/GameIndex/Sprites)`, `Generation`, `Move`, `PokemonType`.
- Entry-points: `GET /api/v2/type/`.

### Version resources
- What it does: Individual game versions (red/blue/yellow…) with names and parent version group.
- Where it lives: `pokemon_v2/api.py:967-970`; `pokemon_v2/serializers.py:3698-3710`; route `pokemon_v2/urls.py:88`.
- How it works: Nests `versionname` + `version_group` summary (kept a summary to avoid circular detail, noted in docstring `pokemon_v2/serializers.py:3698-3703`).
- Inputs/outputs & data deps: Version id/name; deps `Version(Name)`, `VersionGroup`.
- Entry-points: `GET /api/v2/version/`.

### Version-group resources
- What it does: Grouped near-identical releases (red-blue, x-y…): order, generation, versions, regions, learn methods, pokedexes.
- Where it lives: `pokemon_v2/api.py:983-986`; `pokemon_v2/serializers.py:3713-3760`; route `pokemon_v2/urls.py:89`.
- How it works: `get_version_group_regions` via `VersionGroupRegion` (`pokemon_v2/serializers.py:3733-3739`); `get_learn_methods` via `VersionGroupMoveLearnMethod` (`pokemon_v2/serializers.py:3741-3747`); pokedexes via `PokedexVersionGroup`; consumed pervasively by ability/move flavor, machine, and Pokémon move grouping.
- Inputs/outputs & data deps: Version-group id/name; deps `VersionGroup`, `Version`, `Region`, `MoveLearnMethod`, `Pokedex`, `Generation`.
- Entry-points: `GET /api/v2/version-group/`.

### Pokemon-encounters sub-resource
- What it does: All wild encounters for one Pokémon, grouped area→version with per-version `max_chance` and per-slot details (level range, method, condition values).
- Where it lives: route `pokemon_v2/urls.py:101-105`; view `pokemon_v2/api.py:1019-1068`; response serializers `pokemon_v2/api.py:989-1011`.
- How it works: Id-only lookup (`Pokemon.objects.get(pk=pokemon_id)`, unlike name-or-id main resources); `Encounter.filter(pokemon).select_related(location_area/version/slot/method).prefetch_related(condition-maps/details).order_by(area,version,slot)` then double `itertools.groupby` with `max_chance = sum(slot.rarity)` (`pokemon_v2/api.py:1028-1062`); serialized by `PokemonLocationAreaEncounterSerializer` (`pokemon_v2/api.py:1064-1068`). Mirror image exists per-area in `LocationAreaDetailSerializer.get_encounters` grouping Pokémon→version (`pokemon_v2/serializers.py:1096-1138`).
- Inputs/outputs & data deps: In: numeric `pokemon_id`. Out: `[{location_area, version_details:[{version, max_chance, encounter_details}]}]`. Deps `Encounter/Slot/Version/LocationArea/ConditionValueMap/Detail`.
- Entry-points: `GET /api/v2/pokemon/{id}/encounters` (also linked from each Pokémon detail's `location_area_encounters` field).

### API meta endpoint
- What it does: Returns deployed build identity `{deploy_date, hash, tag}` (nullable when git unavailable) plus it is appended to the API root index.
- Where it lives: route `pokemon_v2/urls.py:99`; view `pokemon_v2/api.py:1083-1115`; root augmentation `pokemon_v2/urls.py:24-30`.
- How it works: Shells out to `git rev-parse HEAD`, `git log -1 --format=%ct`, `git describe --tags` (each `try/except → None`, `pokemon_v2/api.py:1085-1110`); `PokeAPIRootView` adds sorted `meta` reverse link to the router index.
- Inputs/outputs & data deps: No inputs; depends on git checkout presence at runtime.
- Entry-points: `GET /api/v2/meta/`; `GET /api/v2/` (index includes `meta` URL).

### Name-or-ID lookup
- What it does: Every `PokeapiCommonViewset` detail accepts either numeric id or case-insensitive name in the same `{pk}` path slot.
- Where it lives: `pokemon_v2/api.py:117-157`; base `pokemon_v2/api.py:176-182`.
- How it works: `ID_PATTERN ^-?[0-9]+$` → `pk=` lookup with int-overflow (>2³¹-1) guard → 404 (`pokemon_v2/api.py:141-146`); else `NAME_PATTERN ^[0-9A-Za-z\-\+ ]+$` → `name__iexact=` with `FieldError→404` for nameless models (`pokemon_v2/api.py:148-152`); anything else → 404. `LocationAreaResource` deliberately opts out (`pokemon_v2/api.py:599-602`); encounters sub-resource is id-only (`pokemon_v2/api.py:1020-1026`).
- Inputs/outputs & data deps: Path string; no extra deps beyond each model's `name` column where present.
- Entry-points: All `GET /api/v2/{resource}/{id|name}/` except location-area/encounters/meta.

### Local `?q=` name search
- What it does: Case-insensitive substring filter on list endpoints; explicitly local-only (absent from pokeapi.co).
- Where it lives: `pokemon_v2/api.py:126-134,160-165`; OpenAPI annotation `q_query_string_parameter` (`pokemon_v2/api.py:160-165`).
- How it works: `get_queryset()` applies `Q(name__icontains=filter_q)` when `?q=` present (`pokemon_v2/api.py:127-134`); documented in schema as "Only available locally and not at pokeapi.co".
- Inputs/outputs & data deps: `?q=` query string; requires `name` field (errors on nameless models surface as FieldError).
- Entry-points: `GET /api/v2/{resource}/?q={substring}` (list only).

### Summary/Detail duality
- What it does: Lists stay tiny (`{name,url}` hyperlinks, or `{url}` for nameless models) while details expand the full localized graph; the public discovery mechanism (HATEOAS-style).
- Where it lives: mixin `pokemon_v2/api.py:99-114`; summaries `pokemon_v2/serializers.py:277-589`; root view `pokemon_v2/urls.py:24-30`.
- How it works: `ListOrDetailSerialRelation.get_serializer_class()` switches on `self.action == "list"` (`pokemon_v2/api.py:106-114`); summary serializers are `HyperlinkedModelSerializer` with `fields=("name","url")` (characteristics/contest-effects/evolution-chains use `("url",)` since they lack names, `pokemon_v2/serializers.py:301-310,369-372`).
- Inputs/outputs & data deps: `?limit&offset` pagination (DRF `LimitOffsetPagination`, `PAGE_SIZE=20`, `config/settings.py:148-154`); default pk-ordered manager (`pokemon_v2/models.py:242-254`).
- Entry-points: Every `GET /api/v2/{resource}/` (summary page) vs `GET .../{id}/` (detail).

### Sprite asset pipeline
- What it does: Decides which sprite URLs exist at build time and stores the resulting JSON blobs (`PokemonSprites`, `PokemonFormSprites`, `ItemSprites`, `TypeSprites`).
- Where it lives: manifest loading `data/v2/build.py:59-147`; path config `data/v2/build.py:157-775`; membership test `data/v2/build.py:778-781`; URL prefixes `data/v2/build.py:43-56`.
- How it works: Priority: `POKEAPI_USE_LOCAL_SPRITES=1` + non-empty `data/v2/sprites` walk (`data/v2/build.py:101-117`) → remote `manifest.json.gz` with `Last-Modified` freshness check and local cache (`data/v2/build.py:68-98`, urls `data/v2/build.py:120-129`) → stale cache → local walk; `POKEMON_SPRITE_CONFIG` enumerates every variant (default/female/shiny × front/back + other/dream-world/home/official-artwork/showdown + version generations I–IX, `data/v2/build.py:157-775`); `file_path_or_none()` returns `POKEAPI_SPRITES_PREFIX + path` iff path ∈ `RESOURCE_IMAGES` else `None` (`data/v2/build.py:778-781`). Empty submodules in this checkout ⇒ sprites resolve to `None`/empty.
- Inputs/outputs & data deps: In: `data/v2/sprites` submodule + remote manifest + `POKEAPI_SPRITES_*` env. Out: sprite JSON columns consumed by `get_pokemon_sprites/get_item_sprites/get_type_sprites`.
- Entry-points: Build-time only (`build_all()` → `_build_pokemons/_build_items/_build_types`); no runtime API.

### Cry asset pipeline
- What it does: Same membership-test pattern for Pokémon cry audio URLs (`.ogg`) stored in `PokemonCries`.
- Where it lives: `data/v2/build.py:148-154,778-781`; prefix `data/v2/build.py:49-54`; submodule path `data/v2/build.py:56`.
- How it works: At import, walks `data/v2/cries/cries/` into `RESOURCE_CRIES` (`data/v2/build.py:150-154`); `file_path_or_none(file, image_file=False)` returns `POKEAPI_CRIES_PREFIX + file` iff present (`data/v2/build.py:779-781`); served at runtime via `PokemonDetailSerializer.get_pokemon_cries` (`pokemon_v2/serializers.py:2999-3002`).
- Inputs/outputs & data deps: `data/v2/cries` submodule + `POKEAPI_CRIES_PREFIX`; empty here ⇒ cries empty.
- Entry-points: Build-time (`_build_pokemons`); runtime `cries` field on Pokémon detail.

### CSV ETL build
- What it does: Offline wipe-and-reload of the whole Postgres/SQLite DB from Veekun-fork CSVs — the only write path (API itself is read-only).
- Where it lives: helpers `data/v2/build.py:784-828`; 26 stages `data/v2/build.py:2477-2503`; invocation `Makefile:32-36`; sources `data/v2/csv/*.csv`.
- How it works: `build_generic()` per CSV: `clear_table()` (delete + `setval(pg_get_serial_sequence…)` or sqlite-sequence reset, `data/v2/build.py:796-804`) then row→object closures `csv_record_to_objects` with `bulk_create` in ≤200-row batches (`data/v2/build.py:807-827`); `build_all()` orders stages languages→regions→generations→versions→…→pokemons→encounters→pal_parks so FK targets exist (`data/v2/build.py:2477-2503`); operator runs `make build-db` (pipes `build_all()` into `manage.py shell`, `Makefile:35-36`); snapshot via `pokeapi.pgdump` (`Makefile:101-107`).
- Inputs/outputs & data deps: In: `data/v2/csv/`, sprites/cries, manifest. Out: all ~140 tables. Deps: Django shell, Postgres/sqlite.
- Entry-points: CLI/build: `make build-db|docker-setup|update-graphql-data-prod`; `from data.v2.build import build_all`.

### GraphQL sidecar
- What it does: Full relational GraphQL access to the same Postgres, bypassing Django/DRF entirely, via external Hasura engine.
- Where it lives: service `docker-compose.yml:50-75`; nginx routes `Resources/nginx/nginx.conf:93-139`; metadata `graphql/v1beta2/` (+ legacy `v1beta/`); targets `Makefile:141-148`.
- How it works: `graphql-engine:v2.50.2` connects with `HASURA_GRAPHQL_DATABASE_URL` straight to `db:5432` (`docker-compose.yml:60`); nginx `/v1beta2` proxies to `graphql-engine:8080/v1/graphql` with strict `limit_req zone=graphqlDefaultLimit rate=1r/m burst=200` + 30 m edge cache vs permissive `/api/` 1 m cache (`Resources/nginx/nginx.conf:67,120-153`); metadata applied with `hasura md apply` (`Makefile:141-148`); `/v1beta2/admin|console` expose Hasura console.
- Inputs/outputs & data deps: GraphQL queries/mutations over Postgres tables; admin secret `HASURA_GRAPHQL_ADMIN_SECRET`, role `anon`.
- Entry-points: `POST /v1beta2` (GraphQL), `/v1beta2/admin|console`; not part of Python codebase.

### OpenAPI schema
- What it does: Checked-in contract (`openapi.yml`) generated from `extend_schema` annotations on every viewset/serializer.
- Where it lives: settings `config/settings.py:153-154,175-280`; annotations e.g. `pokemon_v2/api.py:189-202,1077-1083`; artifact `openapi.yml`; target `Makefile:59-60`.
- How it works: `DEFAULT_SCHEMA_CLASS = drf_spectacular.openapi.AutoSchema` (`config/settings.py:153`); `SPECTACULAR_SETTINGS` pins title/version `2.10.0`, OAS `3.1.0`, servers, and per-tag docs (`config/settings.py:175-280`); each resource has `extend_schema(description/tags/summary)` + list/retrieve summaries; computed fields typed with `extend_schema_field`; regenerate with `make openapi-generate` (`manage.py spectacular`).
- Inputs/outputs & data deps: Code annotations → `openapi.yml`; drift requires re-generation.
- Entry-points: `make openapi-generate`; `openapi.yml` consumed by docs/clients.

### Multi-layer cache
- What it does: Three read-cache layers making the immutable dataset cheap to serve, plus gzip.
- Where it lives: Django `config/settings.py:117-140`; nginx `Resources/nginx/nginx.conf:34-69,141-153`; usage `pokemon_v2/api.py:871-874`, `pokemon_v2/serializers.py:3004-3046`.
- How it works: L1 `django-cachalot` auto-ORM query cache (in `INSTALLED_APPS`, `config/settings.py:137`); L2 `django-redis` (`redis://127.0.0.1:6379/1`, `CONN_MAX_AGE=30`, `config/settings.py:105-125`); L3 nginx `proxy_cache_path /tmp/cache … max_size=2g inactive=10d` (`Resources/nginx/nginx.conf:69`) with `/api/` 1 m + `/v1beta2` 30 m `expires` and public `Cache-Control` (`Resources/nginx/nginx.conf:120-153`); gzip level 4 for JSON/SVG (`Resources/nginx/nginx.conf:34-42`); serializer-level `select_related/prefetch_related` + per-request `*_cache` dicts reduce DB hits before cache.
- Inputs/outputs & data deps: Redis (`cache` service, `docker-compose.yml:2-6`), `/tmp/cache` volume (`graphql_cache`).
- Entry-points: Transparent to API callers; local dev disables via `DummyCache` (`config/local.py:12-16`).

### Environment overlays
- What it does: Same image runs in prod/compose/local by swapping settings modules and env vars.
- Where it lives: base `config/settings.py:54-115`; `config/local.py:1-22`; `config/docker.py`; `config/docker_compose.py:8-29`.
- How it works: Prod defaults: Postgres `localhost`, Redis `127.0.0.1:6379/1`, `DEBUG=False`, CORS `GET`-only on `^/api/.*$` (`config/settings.py:56-146`); `local.py` swaps sqlite + `DummyCache` + `DEBUG=True` (`config/local.py:5-18`); `docker_compose.py` reads `POSTGRES_*/REDIS` service-DNS hosts; `make build-db/serve` select `--settings=config.local` vs compose wiring.
- Inputs/outputs & data deps: Env vars `POSTGRES_*, SECRET_KEY, ALLOWED_HOSTS, BASE_URL, POKEAPI_SPRITES_*`.
- Entry-points: `DJANGO_SETTINGS_MODULE` / `manage.py --settings=`.

### Container/k8s deployment
- What it does: Production topology: Django/gunicorn app + Postgres + Redis + nginx edge + Hasura, also expressed as kustomize/kind/GCP assets.
- Where it lives: `docker-compose.yml:1-80` (+ `docker-compose-dev.yml`, `docker-compose.override.yml`); `Resources/{docker,compose,nginx,k8s/kustomize,gcp,scripts}`; `gunicorn.conf.py:1-6`; `config/wsgi.py:16`; targets `Makefile:1-217`.
- How it works: `app` (bind-mounts repo, waits on healthy `db`, `docker-compose.yml:23-37`) serves WSGI via gunicorn (`SERVER_PORT` bind); `web`/nginx terminates 80/443 and routes `/api/`→`app:80`, `/v1beta2`→Hasura (`Resources/nginx/nginx.conf:76-153`); `db` is `postgres:18-alpine` with `pg_isready` healthcheck; `cache` is `redis:8-alpine`; `make docker-up|docker-setup|docker-test|docker-dump-db|docker-restore-db` orchestrate lifecycle.
- Inputs/outputs & data deps: Docker, env `POSTGRES_USER/PASSWORD/DB`, `./Resources/nginx/ssl`, `graphql_cache` volume.
- Entry-points: `make docker-up|serve`; `gunicorn config.wsgi`; `manage.py`.

## Cross-feature notes
- Shared logic: all 50 REST resources share `PokeapiCommonViewset` (list/detail split + name-or-id + `?q=`), pk-ordered `PokeApiManager` (`pokemon_v2/models.py:242-254`), `Has*` FK + `Is*` language-payload mixins (`pokemon_v2/models.py:273-919`), and the `{name,url}` summary convention — so a reimplementation can build one generic resource factory plus per-resource detail shaping.
- Computed-pattern reuse: Pokémon moves/held-items, location-area encounters, type relations, stat/nature/pokeathlon splits, and evolution trees all follow select→order→`itertools.groupby`→per-group summary serialization with local caches; copy that template for any new rollup.
- Reverse-relation coupling: detail serializers hardcode lowercase `source=` reverse names (e.g. `abilityeffecttext`, `pokemongameindex`); renaming a model breaks many serializers — the highest-risk extension point, covered only by `pokemon_v2/tests.py` factories.
- Localization coupling: every vocab table fans into `Language`/`VersionGroup`/`Generation`; adding a language or game version multiplies rows across abilities/moves/items/species — ETL order in `build_all()` must be preserved.
- Gaps vs a full game-data API: no write path (build-time only), no damage calculator / stat recomputation (only stored `formula` strings and vitamin maps), no search beyond local `?q=` substring, location-area has no name lookup, encounters sub-resource is id-only, and sprites/cries are build-time URL synthesis (empty when submodules/manifest are absent, as in this checkout).
- Reimplementation hints: start from `PokeapiCommonViewset` + summary/detail serializer duality + `build_generic()` bulk ETL; add resources in the order `models → serializers → api.Resource → urls.router.register → extend_schema → make openapi-generate`; mirror the Hasura sidecar by pointing any GraphQL engine at the same Postgres rather than writing a Python GraphQL layer.
