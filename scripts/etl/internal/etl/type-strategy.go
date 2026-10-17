package etl

import (
	"database/sql"
	"fmt"
	"slices"
	"strconv"

	"vgcalc/etl/internal/csv"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TypeStrategy struct {
	CSVPath string
	DB      *sql.DB
}

type PokemonType struct {
	PokeapiID   int
	Slug        string
	DisplayName string
}

var excludedTypeSlugs = []string{"unknown", "shadow", "stellar"}

func (s TypeStrategy) Run() error {
	records, err := csv.ReadCSVFile(s.CSVPath)
	if err != nil {
		return fmt.Errorf("TypeStrategy: error reading csv file. %w", err)
	}
	if len(records) <= 1 {
		return fmt.Errorf("TypeStrategy: file doesn't has enough lines.")
	}

	records = records[1:]
	pokemonTypes := make([]PokemonType, 0)
	caser := cases.Title(language.English)
	for _, record := range records {
		if len(record) < 2 {
			return fmt.Errorf("TypeStrategy: invalid record=%v", record)
		}
		pokeapiID, err := strconv.Atoi(record[0])
		if err != nil {
			return fmt.Errorf("TypeStrategy: invalid pokeapi_id=%q. %w", record[0], err)
		}
		pokemonTypes = append(pokemonTypes, PokemonType{PokeapiID: pokeapiID, Slug: record[1], DisplayName: caser.String(record[1])})
	}

	const query = "INSERT INTO vgcalc.types (pokeapi_id, slug, display_name) VALUES ($1, $2, $3)"
	for _, pokemonType := range pokemonTypes {
		if slices.Contains(excludedTypeSlugs, pokemonType.Slug) {
			continue
		}
		_, err := s.DB.Exec(query, pokemonType.PokeapiID, pokemonType.Slug, pokemonType.DisplayName)
		if err != nil {
			return fmt.Errorf("TypeStrategy: error inserting pokemonType=%v. %w", pokemonType, err)
		}
	}

	return nil
}
