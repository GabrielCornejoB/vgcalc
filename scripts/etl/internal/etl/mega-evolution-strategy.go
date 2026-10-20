package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
)

type MegaEvolutionStrategy struct {
	pokemonCSVPath          string
	pokemonFormsCSVPath     string
	formConditionsCSVPath   string
	pokemonStatsCSVPath     string
	pokemonTypesCSVPath     string
	pokemonAbilitiesCSVPath string
	DB                      *sql.DB
}

type MegaEvolution struct {
	PokeapiID       int
	BasePokemonID   int
	Slug            string
	RequiredItemID  *int
	RequiredMoveID  *int
	AbilityID       int
	PrimaryTypeID   int
	SecondaryTypeID *int
	HP              int
	Atk             int
	Def             int
	SpA             int
	SpD             int
	Spe             int
	Weight          float64
}

func NewMegaEvolutionStrategy(db *sql.DB, dataDir string) MegaEvolutionStrategy {
	return MegaEvolutionStrategy{
		DB:                      db,
		pokemonCSVPath:          dataDir + "pokemon.csv",
		pokemonFormsCSVPath:     dataDir + "pokemon_forms.csv",
		formConditionsCSVPath:   dataDir + "pokemon_form_conditions.csv",
		pokemonStatsCSVPath:     dataDir + "pokemon_stats.csv",
		pokemonTypesCSVPath:     dataDir + "pokemon_types.csv",
		pokemonAbilitiesCSVPath: dataDir + "pokemon_abilities.csv",
	}
}

func (s MegaEvolutionStrategy) Run() error {
	pokemonRows, err := csv.LoadFile[PokemonCSVRow](s.pokemonCSVPath)
	if err != nil {
		return err
	}
	formRows, err := csv.LoadFile[PokemonFormCSVRow](s.pokemonFormsCSVPath)
	if err != nil {
		return err
	}
	conditionRows, err := csv.LoadFile[megaFormConditionCSVRow](s.formConditionsCSVPath)
	if err != nil {
		return err
	}
	statRows, err := csv.LoadFile[PokemonStatCSVRow](s.pokemonStatsCSVPath)
	if err != nil {
		return err
	}
	typeRows, err := csv.LoadFile[PokemonTypeCSVRow](s.pokemonTypesCSVPath)
	if err != nil {
		return err
	}
	abilityRows, err := csv.LoadFile[PokemonAbilityCSVRow](s.pokemonAbilitiesCSVPath)
	if err != nil {
		return err
	}

	typesMap, err := loadPokeapiIDMap(s.DB, vgcalcTypes)
	if err != nil {
		return err
	}
	abilitiesMap, err := loadPokeapiIDMap(s.DB, vgcalcAbilities)
	if err != nil {
		return err
	}
	itemsMap, err := loadPokeapiIDMap(s.DB, vgcalcItems)
	if err != nil {
		return err
	}
	pokemonMap, err := loadPokeapiIDMap(s.DB, vgcalcPokemon)
	if err != nil {
		return err
	}
	movesMap, err := loadPokeapiIDMap(s.DB, vgcalcMoves)
	if err != nil {
		return err
	}

	megaEvolutionIDs := megaEvolutionPokemonIDs(formRows)
	triggers := megaTriggers(formRows, conditionRows)
	slot1s := slot1Abilities(abilityRows)
	bases := basePokemonBySpecies(pokemonRows, megaEvolutionIDs)

	megas, megasMap, err := buildMegaEvolutions(pokemonRows, triggers, slot1s, bases, pokemonMap, itemsMap, movesMap)
	if err != nil {
		return err
	}

	if err := applyMegaStats(statRows, megas, megasMap); err != nil {
		return err
	}
	if err := applyMegaTypes(typeRows, megas, megasMap, typesMap); err != nil {
		return err
	}
	if err := applyMegaAbilities(abilityRows, megas, megasMap, abilitiesMap); err != nil {
		return err
	}

	return upsertMegaEvolutions(s.DB, megas)
}

func upsertMegaEvolutions(db *sql.DB, megas []MegaEvolution) error {
	const query = `INSERT INTO vgcalc.mega_evolution (
		pokeapi_id, base_pokemon_id, slug, required_item_id, required_move_id,
		ability_id, primary_type_id, secondary_type_id,
		hp, atk, def, spa, spd, spe, weight
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	ON CONFLICT (pokeapi_id) DO UPDATE SET
		base_pokemon_id = EXCLUDED.base_pokemon_id,
		slug = EXCLUDED.slug,
		required_item_id = EXCLUDED.required_item_id,
		required_move_id = EXCLUDED.required_move_id,
		ability_id = EXCLUDED.ability_id,
		primary_type_id = EXCLUDED.primary_type_id,
		secondary_type_id = EXCLUDED.secondary_type_id,
		hp = EXCLUDED.hp,
		atk = EXCLUDED.atk,
		def = EXCLUDED.def,
		spa = EXCLUDED.spa,
		spd = EXCLUDED.spd,
		spe = EXCLUDED.spe,
		weight = EXCLUDED.weight`

	return inTx(db, func(tx *sql.Tx) error {
		for _, mega := range megas {
			_, err := tx.Exec(query,
				mega.PokeapiID, mega.BasePokemonID, mega.Slug, mega.RequiredItemID, mega.RequiredMoveID,
				mega.AbilityID, mega.PrimaryTypeID, mega.SecondaryTypeID,
				mega.HP, mega.Atk, mega.Def, mega.SpA, mega.SpD, mega.Spe, mega.Weight)
			if err != nil {
				return fmt.Errorf("error inserting mega %v. %w", mega, err)
			}
		}

		return nil
	})
}

type megaFormConditionCSVRow struct {
	PokemonFormID int `csv:"pokemon_form_id"`
	TriggerItemID int `csv:"trigger_item_id,omitempty"`
}

// rayquazaMegaPokeapiID / dragonAscentPokeapiID pin rayquaza-mega manually:
// it is the only is_mega=1 form with no pokemon_form_conditions row
// (move-triggered, not item-triggered), so neither its identity nor its
// trigger move can be derived from source data. Explicit constants fail
// loudly if PokeAPI ever renumbers them.
const (
	rayquazaMegaPokeapiID = 10079
	dragonAscentPokeapiID = 620
)

// megaTriggers maps every mega/primal pokemon id to its trigger item
// (mega stone or primal orb) from pokemon_form_conditions. Battle-only +
// item-triggered alone would wrongly capture ultra necrozma, crowned
// zacian/zamazenta and terapagos-stellar; the megaEvolutionPokemonIDs
// filter (is_mega=1 or primal) keeps them out.
func megaTriggers(formRows []PokemonFormCSVRow, conditionRows []megaFormConditionCSVRow) map[int]int {
	megaEvolutionIDs := megaEvolutionPokemonIDs(formRows)
	formPokemon := make(map[int]int, len(formRows))
	for _, row := range formRows {
		formPokemon[row.ID] = row.PokemonID
	}

	triggers := make(map[int]int)
	for _, row := range conditionRows {
		if row.TriggerItemID == 0 {
			continue
		}
		pokemonID, ok := formPokemon[row.PokemonFormID]
		if !ok {
			continue
		}
		if _, ok := megaEvolutionIDs[pokemonID]; !ok {
			continue
		}
		triggers[pokemonID] = row.TriggerItemID
	}

	return triggers
}

// slot1Abilities maps pokemon id -> slot-1 ability pokeapi id.
func slot1Abilities(rows []PokemonAbilityCSVRow) map[int]int {
	slot1s := make(map[int]int, len(rows))
	for _, row := range rows {
		if row.Slot != 1 {
			continue
		}
		if _, ok := slot1s[row.PokemonID]; !ok {
			slot1s[row.PokemonID] = row.AbilityID
		}
	}

	return slot1s
}

// basePokemonBySpecies maps species_id -> base pokemon pokeapi id.
// Transformation rows (mega/primal) are never base candidates; when several
// non-transformation varieties share a species, is_default=1 wins. The
// pokemon_form_conditions.base_form_id column is only populated for 8 of 96
// mega rows, so species-sharing stays the usable rule.
func basePokemonBySpecies(pokemonRows []PokemonCSVRow, megaEvolutionIDs map[int]struct{}) map[int]int {
	candidates := make(map[int]PokemonCSVRow)
	for _, row := range pokemonRows {
		if _, excluded := megaEvolutionIDs[row.ID]; excluded {
			continue
		}
		current, exists := candidates[row.SpeciesID]
		if !exists || (row.IsDefault == 1 && current.IsDefault != 1) {
			candidates[row.SpeciesID] = row
		}
	}

	bases := make(map[int]int, len(candidates))
	for speciesID, candidate := range candidates {
		bases[speciesID] = candidate.ID
	}

	return bases
}

// buildMegaEvolutions emits skeleton rows (identity, trigger FKs, weight);
// stats, types and abilities are filled in by the applyMega* passes.
func buildMegaEvolutions(
	pokemonRows []PokemonCSVRow,
	triggers map[int]int,
	slot1s map[int]int,
	bases map[int]int,
	pokemonMap map[int]int,
	itemsMap map[int]int,
	movesMap map[int]int,
) ([]MegaEvolution, map[int]int, error) {
	megas := make([]MegaEvolution, 0, len(triggers)+1)
	megasMap := make(map[int]int, len(triggers)+1)
	for _, row := range pokemonRows {
		triggerItemID, hasTrigger := triggers[row.ID]
		isRayquazaMega := row.ID == rayquazaMegaPokeapiID
		if !hasTrigger && !isRayquazaMega {
			continue
		}
		// Speculative megas (e.g. tatsugiri-drop-mega) carry a trigger item
		// but no confirmed ability in pokemon_abilities — no usable data,
		// so they are skipped.
		if _, ok := slot1s[row.ID]; !ok {
			continue
		}
		basePokeapiID, ok := bases[row.SpeciesID]
		if !ok {
			return nil, nil, fmt.Errorf(
				"buildMegaEvolutions(): no base pokemon for species_id=%d (mega %s)", row.SpeciesID, row.Identifier)
		}
		baseID, ok := pokemonMap[basePokeapiID]
		if !ok {
			return nil, nil, fmt.Errorf("buildMegaEvolutions(): base pokemon pokeapi_id=%d not in vgcalc.pokemon", basePokeapiID)
		}

		mega := MegaEvolution{
			PokeapiID:     row.ID,
			BasePokemonID: baseID,
			Slug:          row.Identifier,
			Weight:        float64(row.Weight) / 10,
		}
		if hasTrigger {
			itemID, ok := itemsMap[triggerItemID]
			if !ok {
				return nil, nil, fmt.Errorf("buildMegaEvolutions(): item pokeapi_id=%d not in vgcalc.items", triggerItemID)
			}
			mega.RequiredItemID = &itemID
		}
		if isRayquazaMega {
			moveID, ok := movesMap[dragonAscentPokeapiID]
			if !ok {
				return nil, nil, fmt.Errorf(
					"buildMegaEvolutions(): move pokeapi_id=%d (dragon-ascent) not in vgcalc.moves", dragonAscentPokeapiID)
			}
			mega.RequiredMoveID = &moveID
		}

		megas = append(megas, mega)
		megasMap[row.ID] = len(megas) - 1
	}

	return megas, megasMap, nil
}

func applyMegaStats(rows []PokemonStatCSVRow, megas []MegaEvolution, megasMap map[int]int) error {
	for _, row := range rows {
		megaIndex, exists := megasMap[row.PokemonID]
		if !exists {
			continue
		}
		switch row.StatID {
		case 1:
			megas[megaIndex].HP = row.BaseStat
		case 2:
			megas[megaIndex].Atk = row.BaseStat
		case 3:
			megas[megaIndex].Def = row.BaseStat
		case 4:
			megas[megaIndex].SpA = row.BaseStat
		case 5:
			megas[megaIndex].SpD = row.BaseStat
		case 6:
			megas[megaIndex].Spe = row.BaseStat
		default:
			return fmt.Errorf("applyMegaStats(): unknown stat_id=%d for mega pokemon_id=%d", row.StatID, row.PokemonID)
		}
	}

	return nil
}

func applyMegaTypes(rows []PokemonTypeCSVRow, megas []MegaEvolution, megasMap map[int]int, typesMap map[int]int) error {
	for _, row := range rows {
		megaIndex, exists := megasMap[row.PokemonID]
		if !exists {
			continue
		}
		typeID, exists := typesMap[row.TypeID]
		if !exists {
			return fmt.Errorf("applyMegaTypes(): unknown type_id=%d for mega pokemon_id=%d", row.TypeID, row.PokemonID)
		}
		switch row.Slot {
		case 1:
			megas[megaIndex].PrimaryTypeID = typeID
		case 2:
			secondaryTypeID := typeID
			megas[megaIndex].SecondaryTypeID = &secondaryTypeID
		default:
			return fmt.Errorf("applyMegaTypes(): unknown slot=%d for mega pokemon_id=%d", row.Slot, row.PokemonID)
		}
	}

	return nil
}

func applyMegaAbilities(rows []PokemonAbilityCSVRow, megas []MegaEvolution, megasMap map[int]int,
	abilitiesMap map[int]int) error {
	for _, row := range rows {
		megaIndex, exists := megasMap[row.PokemonID]
		if !exists {
			continue
		}
		if row.Slot != 1 {
			return fmt.Errorf(
				"applyMegaAbilities(): unknown slot=%d for mega pokemon_id=%d (megas have a single slot-1 ability)",
				row.Slot, row.PokemonID)
		}
		abilityID, exists := abilitiesMap[row.AbilityID]
		if !exists {
			return fmt.Errorf("applyMegaAbilities(): unknown ability_id=%d for mega pokemon_id=%d", row.AbilityID, row.PokemonID)
		}
		megas[megaIndex].AbilityID = abilityID
	}

	return nil
}
