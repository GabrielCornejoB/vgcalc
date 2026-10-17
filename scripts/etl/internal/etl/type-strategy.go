package etl

import (
	"database/sql"
	"fmt"
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

func NewTypeStrategy(db *sql.DB, dataDir string) TypeStrategy {
	return TypeStrategy{DB: db, CSVPath: dataDir + "types.csv"}
}

func (s TypeStrategy) Run() error {
	records, err := csv.ReadCSVFile(s.CSVPath)
	if err != nil {
		return fmt.Errorf("TypeStrategy: error reading csv file. %w", err)
	}

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

	const query = `INSERT INTO vgcalc.types (pokeapi_id, slug, display_name) VALUES ($1, $2, $3)
	ON CONFLICT (pokeapi_id) DO UPDATE SET slug = EXCLUDED.slug, display_name = EXCLUDED.display_name`

	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("TypeStrategy: error starting transaction. %w", err)
	}
	defer tx.Rollback()

	for _, pokemonType := range pokemonTypes {
		_, err := tx.Exec(query, pokemonType.PokeapiID, pokemonType.Slug, pokemonType.DisplayName)
		if err != nil {
			return fmt.Errorf("TypeStrategy: error inserting pokemonType=%v. %w", pokemonType, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("TypeStrategy: error committing transaction. %w", err)
	}

	return nil
}
