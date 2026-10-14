package etl

import (
	"database/sql"
	"fmt"
	"slices"
	"strconv"

	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/env"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TypeStrategy struct {
	filePathUrl string
	db          *sql.DB
}

func NewTypeStrategy(db *sql.DB) (TypeStrategy, error) {
	filePath, err := env.GetEnvByKey("PATH_TYPES_CSV")
	if err != nil {
		return TypeStrategy{}, fmt.Errorf("NewTypeEtlStrategy(): %w", err)
	}
	return TypeStrategy{filePath, db}, nil
}

type PokemonType struct {
	PokeapiId   int
	Slug        string
	DisplayName string
}

func (strategy TypeStrategy) extract() ([][]string, error) {
	records, err := csv.ReadCsvFile(strategy.filePathUrl)
	if err != nil {
		return nil, fmt.Errorf("TypeStrategy.extract(): %w", err)
	}

	return records, nil
}

func (strategy TypeStrategy) transform(records [][]string) ([]PokemonType, error) {
	records = records[1:]
	pokemonTypes := make([]PokemonType, 0)
	caser := cases.Title(language.English)

	for _, record := range records {
		pokeapiId, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("TypeStrategy.transform(): invalid pokeapi_id=%q. %w", record[0], err)
		}
		pokemonTypes = append(pokemonTypes, PokemonType{PokeapiId: pokeapiId, Slug: record[1], DisplayName: caser.String(record[1])})
	}

	return pokemonTypes, nil
}

func (strategy TypeStrategy) load(pokemonTypes []PokemonType) error {
	const query = "INSERT INTO vgcalc.types (pokeapi_id, slug, display_name) VALUES ($1, $2, $3)"
	var blackListedTypes = []string{"unknown", "shadow"}

	for _, pokemonType := range pokemonTypes {
		if slices.Contains(blackListedTypes, pokemonType.Slug) {
			continue
		}
		_, err := strategy.db.Exec(query, pokemonType.PokeapiId, pokemonType.Slug, pokemonType.DisplayName)
		if err != nil {
			return fmt.Errorf("TypeStrategy.load(): pokemonType=%v. %w", pokemonType, err)
		}
	}

	return nil
}

func (strategy TypeStrategy) Run() error {
	records, err := strategy.extract()
	if err != nil {
		return fmt.Errorf("TypeStrategy.Run(): %w", err)
	}

	pokemonTypes, err := strategy.transform(records)
	if err != nil {
		return fmt.Errorf("TypeStrategy.Run(): %w", err)
	}

	err = strategy.load(pokemonTypes)
	if err != nil {
		return fmt.Errorf("TypeStrategy.Run(): %w", err)
	}

	return nil
}
