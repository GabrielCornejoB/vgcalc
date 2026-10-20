package etl

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"vgcalc/etl/internal/csv"
)

type PokemonStrategy struct {
	pokemonCSVPath          string
	pokemonStatsCSVPath     string
	pokemonTypesCSVPath     string
	pokemonAbilitiesCSVPath string
	pokemonFormsCSVPath     string
	pokemonSpeciesCSVPath   string
	DB                      *sql.DB
}

type Pokemon struct {
	PokeapiID       int
	Slug            string
	DisplayName     string
	PrimaryTypeID   int
	SecondaryTypeID *int
	HP              int
	Atk             int
	Def             int
	SpA             int
	SpD             int
	Spe             int
	Ability1ID      int
	Ability2ID      *int
	HiddenAbilityID *int
	Weight          float64
	FixedGender     *string
}

type pokemonSpeciesCSVRow struct {
	ID         int `csv:"id"`
	GenderRate int `csv:"gender_rate"`
}

func NewPokemonStrategy(db *sql.DB, dataDir string) PokemonStrategy {
	return PokemonStrategy{
		DB:                      db,
		pokemonCSVPath:          dataDir + "pokemon.csv",
		pokemonStatsCSVPath:     dataDir + "pokemon_stats.csv",
		pokemonTypesCSVPath:     dataDir + "pokemon_types.csv",
		pokemonAbilitiesCSVPath: dataDir + "pokemon_abilities.csv",
		pokemonFormsCSVPath:     dataDir + "pokemon_forms.csv",
		pokemonSpeciesCSVPath:   dataDir + "pokemon_species.csv",
	}
}

func (s PokemonStrategy) Run() error {
	pokemonRows, err := csv.LoadFile[PokemonCSVRow](s.pokemonCSVPath)
	if err != nil {
		return err
	}

	formRows, err := csv.LoadFile[PokemonFormCSVRow](s.pokemonFormsCSVPath)
	if err != nil {
		return err
	}
	megaIDs := megaEvolutionPokemonIDs(formRows)

	speciesRows, err := csv.LoadFile[pokemonSpeciesCSVRow](s.pokemonSpeciesCSVPath)
	if err != nil {
		return err
	}
	speciesGenders := speciesFixedGenders(speciesRows)

	pokemons, pokemonsMap := buildPokemon(pokemonRows, megaIDs, speciesGenders)

	typesMap, err := loadPokeapiIDMap(s.DB, vgcalcTypes)
	if err != nil {
		return err
	}
	abilitiesMap, err := loadPokeapiIDMap(s.DB, vgcalcAbilities)
	if err != nil {
		return err
	}

	statRows, err := csv.LoadFile[PokemonStatCSVRow](s.pokemonStatsCSVPath)
	if err != nil {
		return err
	}
	if err := applyPokemonStats(statRows, pokemons, pokemonsMap); err != nil {
		return err
	}

	typeRows, err := csv.LoadFile[PokemonTypeCSVRow](s.pokemonTypesCSVPath)
	if err != nil {
		return err
	}
	if err := applyPokemonTypes(typeRows, pokemons, pokemonsMap, typesMap); err != nil {
		return err
	}

	abilityRows, err := csv.LoadFile[PokemonAbilityCSVRow](s.pokemonAbilitiesCSVPath)
	if err != nil {
		return err
	}
	if err := applyPokemonAbilities(abilityRows, pokemons, pokemonsMap, abilitiesMap); err != nil {
		return err
	}

	const query = `INSERT INTO vgcalc.pokemon (
		pokeapi_id, slug, display_name, primary_type_id, secondary_type_id,
		hp, atk, def, spa, spd, spe,
		ability1_id, ability2_id, hidden_ability_id, weight, fixed_gender
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (pokeapi_id) DO UPDATE SET
		slug = EXCLUDED.slug,
		display_name = EXCLUDED.display_name,
		primary_type_id = EXCLUDED.primary_type_id,
		secondary_type_id = EXCLUDED.secondary_type_id,
		hp = EXCLUDED.hp,
		atk = EXCLUDED.atk,
		def = EXCLUDED.def,
		spa = EXCLUDED.spa,
		spd = EXCLUDED.spd,
		spe = EXCLUDED.spe,
		ability1_id = EXCLUDED.ability1_id,
		ability2_id = EXCLUDED.ability2_id,
		hidden_ability_id = EXCLUDED.hidden_ability_id,
		weight = EXCLUDED.weight,
		fixed_gender = EXCLUDED.fixed_gender`

	return inTx(s.DB, func(tx *sql.Tx) error {
		for _, pokemon := range pokemons {
			_, err := tx.Exec(query,
				pokemon.PokeapiID, pokemon.Slug, pokemon.DisplayName, pokemon.PrimaryTypeID, pokemon.SecondaryTypeID,
				pokemon.HP, pokemon.Atk, pokemon.Def, pokemon.SpA, pokemon.SpD, pokemon.Spe,
				pokemon.Ability1ID, pokemon.Ability2ID, pokemon.HiddenAbilityID, pokemon.Weight, pokemon.FixedGender)
			if err != nil {
				return fmt.Errorf("error inserting pokemon %v. %w", pokemon, err)
			}
		}

		return nil
	})
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

// speciesFixedGenders keeps only species with a locked gender:
// gender_rate is the female chance in eighths, so -1 is genderless,
// 0 is all-male and 8 is all-female; 1-7 are mixed and stay out of the map.
func speciesFixedGenders(rows []pokemonSpeciesCSVRow) map[int]string {
	speciesGenders := make(map[int]string, len(rows))
	for _, row := range rows {
		switch row.GenderRate {
		case -1:
			speciesGenders[row.ID] = "genderless"
		case 0:
			speciesGenders[row.ID] = "male"
		case 8:
			speciesGenders[row.ID] = "female"
		}
	}

	return speciesGenders
}

func buildPokemon(
	rows []PokemonCSVRow, megaIDs map[int]struct{}, speciesGenders map[int]string,
) ([]Pokemon, map[int]int) {
	caser := cases.Title(language.English)
	pokemons := make([]Pokemon, 0, len(rows))
	pokemonsMap := make(map[int]int, len(rows))
	for _, row := range rows {
		if _, excluded := megaIDs[row.ID]; excluded {
			continue
		}
		// Source rows with non-positive weight cannot satisfy CHECK (weight > 0).
		// Currently only eternatus-eternamax (10190, weight 0, not playable).
		if row.Weight <= 0 {
			continue
		}
		pokemons = append(pokemons, Pokemon{
			PokeapiID:   row.ID,
			Slug:        row.Identifier,
			DisplayName: caser.String(strings.ReplaceAll(row.Identifier, "-", " ")),
			Weight:      float64(row.Weight) / 10,
			FixedGender: pokemonFixedGender(row.Identifier, speciesGenders[row.SpeciesID]),
		})
		pokemonsMap[row.ID] = len(pokemons) - 1
	}

	return pokemons, pokemonsMap
}

// pokemonFixedGender prefers the slug suffix, which is the only signal for
// form-split variants of a mixed species (e.g. indeedee-male, indeedee-female).
func pokemonFixedGender(slug string, speciesGender string) *string {
	gender := speciesGender
	switch {
	case strings.HasSuffix(slug, "-male"):
		gender = "male"
	case strings.HasSuffix(slug, "-female"):
		gender = "female"
	}
	if gender == "" {
		return nil
	}

	return &gender
}

func applyPokemonStats(rows []PokemonStatCSVRow, pokemons []Pokemon, pokemonsMap map[int]int) error {
	for _, row := range rows {
		pokemonIndex, exists := pokemonsMap[row.PokemonID]
		if !exists {
			continue
		}
		switch row.StatID {
		case 1:
			pokemons[pokemonIndex].HP = row.BaseStat
		case 2:
			pokemons[pokemonIndex].Atk = row.BaseStat
		case 3:
			pokemons[pokemonIndex].Def = row.BaseStat
		case 4:
			pokemons[pokemonIndex].SpA = row.BaseStat
		case 5:
			pokemons[pokemonIndex].SpD = row.BaseStat
		case 6:
			pokemons[pokemonIndex].Spe = row.BaseStat
		default:
			return fmt.Errorf("applyPokemonStats(): unknown stat_id=%d for pokemon_id=%d", row.StatID, row.PokemonID)
		}
	}

	return nil
}

func applyPokemonTypes(
	rows []PokemonTypeCSVRow, pokemons []Pokemon, pokemonsMap map[int]int, typesMap map[int]int,
) error {
	for _, row := range rows {
		pokemonIndex, exists := pokemonsMap[row.PokemonID]
		if !exists {
			continue
		}
		typeID, exists := typesMap[row.TypeID]
		if !exists {
			return fmt.Errorf("applyPokemonTypes(): unknown type_id=%d for pokemon_id=%d", row.TypeID, row.PokemonID)
		}
		switch row.Slot {
		case 1:
			pokemons[pokemonIndex].PrimaryTypeID = typeID
		case 2:
			secondaryTypeID := typeID
			pokemons[pokemonIndex].SecondaryTypeID = &secondaryTypeID
		default:
			return fmt.Errorf("applyPokemonTypes(): unknown slot=%d for pokemon_id=%d", row.Slot, row.PokemonID)
		}
	}

	return nil
}

func applyPokemonAbilities(rows []PokemonAbilityCSVRow, pokemons []Pokemon, pokemonsMap map[int]int,
	abilitiesMap map[int]int) error {
	for _, row := range rows {
		pokemonIndex, exists := pokemonsMap[row.PokemonID]
		if !exists {
			continue
		}
		abilityID, exists := abilitiesMap[row.AbilityID]
		if !exists {
			return fmt.Errorf("applyPokemonAbilities(): unknown ability_id=%d for pokemon_id=%d",
				row.AbilityID, row.PokemonID)
		}
		if row.IsHidden == 1 {
			hiddenAbilityID := abilityID
			pokemons[pokemonIndex].HiddenAbilityID = &hiddenAbilityID
			continue
		}
		switch row.Slot {
		case 1:
			pokemons[pokemonIndex].Ability1ID = abilityID
		case 2:
			ability2ID := abilityID
			pokemons[pokemonIndex].Ability2ID = &ability2ID
		default:
			return fmt.Errorf("applyPokemonAbilities(): unknown slot=%d for pokemon_id=%d", row.Slot, row.PokemonID)
		}
	}

	return nil
}
