package etl

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/etl/common"
	"vgcalc/etl/internal/etl/entity"
)

type PokemonStrategy struct {
	dataDir               string
	pokemonSpeciesCSVPath string
	DB                    *sql.DB
}

type pokemonSpeciesCSVRow struct {
	ID         int `csv:"id"`
	GenderRate int `csv:"gender_rate"`
}

func NewPokemonStrategy(db *sql.DB, dataDir string) PokemonStrategy {
	return PokemonStrategy{
		DB:                    db,
		dataDir:               dataDir,
		pokemonSpeciesCSVPath: dataDir + "pokemon_species.csv",
	}
}

func (s PokemonStrategy) Run() error {
	src, err := s.extract()
	if err != nil {
		return err
	}
	pokemons, err := transformPokemon(src)
	if err != nil {
		return err
	}

	return loadPokemon(s.DB, pokemons)
}

type pokemonStrategySource struct {
	common.PokemonSource
	fixedGenderBySpecies map[int]string
}

func (s PokemonStrategy) extract() (pokemonStrategySource, error) {
	var src pokemonStrategySource
	var err error

	if src.PokemonSource, err = common.ExtractPokemonSource(s.DB, s.dataDir); err != nil {
		return src, err
	}
	speciesRows, err := csv.ReadFile[pokemonSpeciesCSVRow](s.pokemonSpeciesCSVPath)
	if err != nil {
		return src, err
	}
	src.fixedGenderBySpecies = speciesFixedGenders(speciesRows)

	return src, nil
}

func transformPokemon(src pokemonStrategySource) ([]entity.Pokemon, error) {
	caser := cases.Title(language.English)
	pokemons := make([]entity.Pokemon, 0, len(src.PokemonRows))
	for _, row := range src.PokemonRows {
		if _, excluded := src.MegaIDs[row.ID]; excluded {
			continue
		}
		// Source rows with non-positive weight cannot satisfy CHECK (weight > 0).
		// Currently only eternatus-eternamax (10190, weight 0, not playable).
		if row.Weight <= 0 {
			continue
		}
		pokemon := entity.Pokemon{
			PokeapiID:    row.ID,
			Slug:         row.Identifier,
			DisplayName:  caser.String(strings.ReplaceAll(row.Identifier, "-", " ")),
			Stats:        src.Stats[row.ID],
			TypeSlots:    src.Types[row.ID],
			AbilitySlots: src.Abilities[row.ID],
			Weight:       float64(row.Weight) / 10,
			FixedGender:  pokemonFixedGender(row.Identifier, src.fixedGenderBySpecies[row.SpeciesID]),
		}
		if err := pokemon.Validate(); err != nil {
			return nil, fmt.Errorf("transformPokemon(): %w", err)
		}
		pokemons = append(pokemons, pokemon)
	}

	return pokemons, nil
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

func loadPokemon(db *sql.DB, pokemons []entity.Pokemon) error {
	const query = `INSERT INTO vgcalc.pokemon (
		pokeapi_id, slug, display_name, primary_type_id, secondary_type_id,
		hp, atk, def, spa, spd, spe,
		ability1_id, ability2_id, hidden_ability_id, weight, fixed_gender
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (pokeapi_id) DO NOTHING`

	return common.InTx(db, func(tx *sql.Tx) error {
		for _, pokemon := range pokemons {
			_, err := tx.Exec(query,
				pokemon.PokeapiID, pokemon.Slug, pokemon.DisplayName, pokemon.Primary, pokemon.Secondary,
				pokemon.HP, pokemon.Atk, pokemon.Def, pokemon.SpA, pokemon.SpD, pokemon.Spe,
				pokemon.Slot1, pokemon.Slot2, pokemon.Hidden, pokemon.Weight, pokemon.FixedGender)
			if err != nil {
				return fmt.Errorf("error inserting pokemon %v. %w", pokemon, err)
			}
		}

		return nil
	})
}
