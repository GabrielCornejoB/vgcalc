package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/etl/common"
	"vgcalc/etl/internal/etl/entity"
)

type MegaEvolutionStrategy struct {
	dataDir               string
	pokemonFormsCSVPath   string
	formConditionsCSVPath string
	DB                    *sql.DB
}

func NewMegaEvolutionStrategy(db *sql.DB, dataDir string) MegaEvolutionStrategy {
	return MegaEvolutionStrategy{
		DB:                    db,
		dataDir:               dataDir,
		pokemonFormsCSVPath:   dataDir + "pokemon_forms.csv",
		formConditionsCSVPath: dataDir + "pokemon_form_conditions.csv",
	}
}

func (s MegaEvolutionStrategy) Run() error {
	src, err := s.extract()
	if err != nil {
		return err
	}
	megas, err := transformMegaEvolutions(src)
	if err != nil {
		return err
	}

	return loadMegaEvolutions(s.DB, megas)
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

type megaEvolutionSource struct {
	common.PokemonSource
	itemIDByMega           map[int]int
	basePokemonIDBySpecies map[int]int
	dragonAscentMoveID     int
}

func (s MegaEvolutionStrategy) extract() (megaEvolutionSource, error) {
	var src megaEvolutionSource
	var err error

	if src.PokemonSource, err = common.ExtractPokemonSource(s.DB, s.dataDir); err != nil {
		return src, err
	}
	formRows, err := csv.ReadFile[common.PokemonFormCSVRow](s.pokemonFormsCSVPath)
	if err != nil {
		return src, err
	}
	conditionRows, err := csv.ReadFile[megaFormConditionCSVRow](s.formConditionsCSVPath)
	if err != nil {
		return src, err
	}
	itemsMap, err := common.QueryPokeapiIDMap(s.DB, common.TableItems)
	if err != nil {
		return src, err
	}
	pokemonMap, err := common.QueryPokeapiIDMap(s.DB, common.TablePokemon)
	if err != nil {
		return src, err
	}
	movesMap, err := common.QueryPokeapiIDMap(s.DB, common.TableMoves)
	if err != nil {
		return src, err
	}

	triggers := megaTriggers(formRows, conditionRows, src.MegaIDs)
	if src.itemIDByMega, err = resolveMegaItems(triggers, itemsMap); err != nil {
		return src, err
	}
	src.basePokemonIDBySpecies = resolveBasePokemon(basePokemonBySpecies(src.PokemonRows, src.MegaIDs), pokemonMap)
	dragonAscentMoveID, ok := movesMap[dragonAscentPokeapiID]
	if !ok {
		return src, fmt.Errorf(
			"extract(): move pokeapi_id=%d (dragon-ascent) not in vgcalc.moves", dragonAscentPokeapiID)
	}
	src.dragonAscentMoveID = dragonAscentMoveID

	return src, nil
}

// megaTriggers maps every mega/primal pokemon id to its trigger item
// (mega stone or primal orb) from pokemon_form_conditions. Battle-only +
// item-triggered alone would wrongly capture ultra necrozma, crowned
// zacian/zamazenta and terapagos-stellar; the megaEvolutionPokemonIDs
// filter (is_mega=1 or primal) keeps them out.
func megaTriggers(
	formRows []common.PokemonFormCSVRow, conditionRows []megaFormConditionCSVRow, megaIDs map[int]struct{},
) map[int]int {
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
		if _, ok := megaIDs[pokemonID]; !ok {
			continue
		}
		triggers[pokemonID] = row.TriggerItemID
	}

	return triggers
}

func resolveMegaItems(triggers map[int]int, itemsMap map[int]int) (map[int]int, error) {
	itemIDByMega := make(map[int]int, len(triggers))
	for megaID, triggerItemID := range triggers {
		itemID, ok := itemsMap[triggerItemID]
		if !ok {
			return nil, fmt.Errorf("extract(): item pokeapi_id=%d not in vgcalc.items", triggerItemID)
		}
		itemIDByMega[megaID] = itemID
	}

	return itemIDByMega, nil
}

// basePokemonBySpecies maps species_id -> base pokemon pokeapi id.
// Transformation rows (mega/primal) are never base candidates; when several
// non-transformation varieties share a species, is_default=1 wins. The
// pokemon_form_conditions.base_form_id column is only populated for 8 of 96
// mega rows, so species-sharing stays the usable rule.
func basePokemonBySpecies(pokemonRows []common.PokemonCSVRow, megaIDs map[int]struct{}) map[int]int {
	candidates := make(map[int]common.PokemonCSVRow)
	for _, row := range pokemonRows {
		if _, excluded := megaIDs[row.ID]; excluded {
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

// resolveBasePokemon drops species whose base is not in vgcalc.pokemon, so a
// mega of such a species fails in transform as having no base pokemon.
func resolveBasePokemon(bases map[int]int, pokemonMap map[int]int) map[int]int {
	resolved := make(map[int]int, len(bases))
	for speciesID, basePokeapiID := range bases {
		if baseID, ok := pokemonMap[basePokeapiID]; ok {
			resolved[speciesID] = baseID
		}
	}

	return resolved
}

func transformMegaEvolutions(src megaEvolutionSource) ([]entity.MegaEvolution, error) {
	megas := make([]entity.MegaEvolution, 0, len(src.itemIDByMega)+1)
	for _, row := range src.PokemonRows {
		itemID, hasTrigger := src.itemIDByMega[row.ID]
		isRayquazaMega := row.ID == rayquazaMegaPokeapiID
		if !hasTrigger && !isRayquazaMega {
			continue
		}
		// Speculative megas (e.g. tatsugiri-drop-mega) carry a trigger item
		// but no confirmed ability in pokemon_abilities — no usable data,
		// so they are skipped.
		abilities, ok := src.Abilities[row.ID]
		if !ok || abilities.Slot1 == 0 {
			continue
		}
		if abilities.Slot2 != nil || abilities.Hidden != nil {
			return nil, fmt.Errorf(
				"transformMegaEvolutions(): mega %s has more than one ability (megas have a single slot-1 ability)",
				row.Identifier)
		}
		baseID, ok := src.basePokemonIDBySpecies[row.SpeciesID]
		if !ok {
			return nil, fmt.Errorf(
				"transformMegaEvolutions(): no base pokemon for species_id=%d (mega %s)", row.SpeciesID, row.Identifier)
		}

		mega := entity.MegaEvolution{
			PokeapiID:     row.ID,
			BasePokemonID: baseID,
			Slug:          row.Identifier,
			AbilityID:     abilities.Slot1,
			Stats:         src.Stats[row.ID],
			TypeSlots:     src.Types[row.ID],
			Weight:        float64(row.Weight) / 10,
		}
		if hasTrigger {
			mega.RequiredItemID = &itemID
		}
		if isRayquazaMega {
			moveID := src.dragonAscentMoveID
			mega.RequiredMoveID = &moveID
		}
		if err := mega.Validate(); err != nil {
			return nil, fmt.Errorf("transformMegaEvolutions(): %w", err)
		}
		megas = append(megas, mega)
	}

	return megas, nil
}

func loadMegaEvolutions(db *sql.DB, megas []entity.MegaEvolution) error {
	const query = `INSERT INTO vgcalc.mega_evolution (
		pokeapi_id, base_pokemon_id, slug, required_item_id, required_move_id,
		ability_id, primary_type_id, secondary_type_id,
		hp, atk, def, spa, spd, spe, weight
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	ON CONFLICT (pokeapi_id) DO NOTHING`

	return common.InTx(db, func(tx *sql.Tx) error {
		for _, mega := range megas {
			_, err := tx.Exec(query,
				mega.PokeapiID, mega.BasePokemonID, mega.Slug, mega.RequiredItemID, mega.RequiredMoveID,
				mega.AbilityID, mega.Primary, mega.Secondary,
				mega.HP, mega.Atk, mega.Def, mega.SpA, mega.SpD, mega.Spe, mega.Weight)
			if err != nil {
				return fmt.Errorf("error inserting mega %v. %w", mega, err)
			}
		}

		return nil
	})
}
