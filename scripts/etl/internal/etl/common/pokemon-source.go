package common

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/etl/entity"
)

// Row structs for the pokemon*.csv family, shared by every strategy that
// consumes them (pokemon, mega_evolution). Field sets are the union of what
// each strategy reads; csvutil ignores columns a strategy does not use.

type PokemonCSVRow struct {
	ID         int    `csv:"id"`
	Identifier string `csv:"identifier"`
	SpeciesID  int    `csv:"species_id"`
	Weight     int    `csv:"weight"`
	IsDefault  int    `csv:"is_default"`
}

type PokemonStatCSVRow struct {
	PokemonID int `csv:"pokemon_id"`
	StatID    int `csv:"stat_id"`
	BaseStat  int `csv:"base_stat"`
}

type PokemonTypeCSVRow struct {
	PokemonID int `csv:"pokemon_id"`
	TypeID    int `csv:"type_id"`
	Slot      int `csv:"slot"`
}

type PokemonAbilityCSVRow struct {
	PokemonID int `csv:"pokemon_id"`
	AbilityID int `csv:"ability_id"`
	IsHidden  int `csv:"is_hidden"`
	Slot      int `csv:"slot"`
}

type PokemonFormCSVRow struct {
	ID             int    `csv:"id"`
	PokemonID      int    `csv:"pokemon_id"`
	IsMega         int    `csv:"is_mega"`
	FormIdentifier string `csv:"form_identifier"`
	IsBattleOnly   int    `csv:"is_battle_only"`
}

type PokemonSource struct {
	PokemonRows []PokemonCSVRow
	MegaIDs     map[int]struct{}
	Stats       map[int]entity.Stats
	Types       map[int]entity.TypeSlots
	Abilities   map[int]entity.AbilitySlots
}

func ExtractPokemonSource(db *sql.DB, dataDir string) (PokemonSource, error) {
	var src PokemonSource

	pokemonRows, err := csv.ReadFile[PokemonCSVRow](dataDir + "pokemon.csv")
	if err != nil {
		return src, err
	}
	formRows, err := csv.ReadFile[PokemonFormCSVRow](dataDir + "pokemon_forms.csv")
	if err != nil {
		return src, err
	}
	statRows, err := csv.ReadFile[PokemonStatCSVRow](dataDir + "pokemon_stats.csv")
	if err != nil {
		return src, err
	}
	typeRows, err := csv.ReadFile[PokemonTypeCSVRow](dataDir + "pokemon_types.csv")
	if err != nil {
		return src, err
	}
	abilityRows, err := csv.ReadFile[PokemonAbilityCSVRow](dataDir + "pokemon_abilities.csv")
	if err != nil {
		return src, err
	}
	typesMap, err := QueryPokeapiIDMap(db, TableTypes)
	if err != nil {
		return src, err
	}
	abilitiesMap, err := QueryPokeapiIDMap(db, TableAbilities)
	if err != nil {
		return src, err
	}

	src.PokemonRows = pokemonRows
	src.MegaIDs = megaEvolutionPokemonIDs(formRows)
	if src.Stats, err = groupStats(statRows); err != nil {
		return src, err
	}
	if src.Types, err = groupTypes(typeRows, typesMap); err != nil {
		return src, err
	}
	if src.Abilities, err = groupAbilities(abilityRows, abilitiesMap); err != nil {
		return src, err
	}

	return src, nil
}

// megaEvolutionPokemonIDs returns the pokemon ids behind every row that
// belongs in vgcalc.mega_evolution: is_mega=1 forms plus the two primals
// (kyogre/groudon, is_battle_only=1 with form_identifier "primal").
// Other battle-only forms (gmax, ultra, crowned, stellar) are not megas
// and stay in vgcalc.pokemon.
func megaEvolutionPokemonIDs(rows []PokemonFormCSVRow) map[int]struct{} {
	megaIDs := make(map[int]struct{})
	for _, row := range rows {
		if row.IsMega == 1 || (row.IsBattleOnly == 1 && row.FormIdentifier == "primal") {
			megaIDs[row.PokemonID] = struct{}{}
		}
	}

	return megaIDs
}

func groupStats(rows []PokemonStatCSVRow) (map[int]entity.Stats, error) {
	stats := make(map[int]entity.Stats)
	for _, row := range rows {
		s := stats[row.PokemonID]
		switch row.StatID {
		case StatHP:
			s.HP = row.BaseStat
		case StatAtk:
			s.Atk = row.BaseStat
		case StatDef:
			s.Def = row.BaseStat
		case StatSpA:
			s.SpA = row.BaseStat
		case StatSpD:
			s.SpD = row.BaseStat
		case StatSpe:
			s.Spe = row.BaseStat
		default:
			return nil, fmt.Errorf("groupStats(): unknown stat_id=%d for pokemon_id=%d", row.StatID, row.PokemonID)
		}
		stats[row.PokemonID] = s
	}

	return stats, nil
}

func groupTypes(rows []PokemonTypeCSVRow, typesMap map[int]int) (map[int]entity.TypeSlots, error) {
	types := make(map[int]entity.TypeSlots)
	for _, row := range rows {
		typeID, exists := typesMap[row.TypeID]
		if !exists {
			return nil, fmt.Errorf("groupTypes(): unknown type_id=%d for pokemon_id=%d", row.TypeID, row.PokemonID)
		}
		slots := types[row.PokemonID]
		switch row.Slot {
		case 1:
			slots.Primary = typeID
		case 2:
			slots.Secondary = &typeID
		default:
			return nil, fmt.Errorf("groupTypes(): unknown slot=%d for pokemon_id=%d", row.Slot, row.PokemonID)
		}
		types[row.PokemonID] = slots
	}

	return types, nil
}

func groupAbilities(
	rows []PokemonAbilityCSVRow, abilitiesMap map[int]int,
) (map[int]entity.AbilitySlots, error) {
	abilities := make(map[int]entity.AbilitySlots)
	for _, row := range rows {
		abilityID, exists := abilitiesMap[row.AbilityID]
		if !exists {
			return nil, fmt.Errorf("groupAbilities(): unknown ability_id=%d for pokemon_id=%d",
				row.AbilityID, row.PokemonID)
		}
		slots := abilities[row.PokemonID]
		if row.IsHidden == 1 {
			slots.Hidden = &abilityID
			abilities[row.PokemonID] = slots
			continue
		}
		switch row.Slot {
		case 1:
			slots.Slot1 = abilityID
		case 2:
			slots.Slot2 = &abilityID
		default:
			return nil, fmt.Errorf("groupAbilities(): unknown slot=%d for pokemon_id=%d", row.Slot, row.PokemonID)
		}
		abilities[row.PokemonID] = slots
	}

	return abilities, nil
}
