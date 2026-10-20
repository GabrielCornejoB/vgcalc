package etl

import (
	"database/sql"
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"vgcalc/etl/internal/csv"
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
	rows, err := csv.LoadFile[typeCSVRow](s.CSVPath)
	if err != nil {
		return err
	}

	pokemonTypes := loadTypes(rows)

	const query = `INSERT INTO vgcalc.types (pokeapi_id, slug, display_name) VALUES ($1, $2, $3)
	ON CONFLICT (pokeapi_id) DO UPDATE SET slug = EXCLUDED.slug, display_name = EXCLUDED.display_name`

	return inTx(s.DB, func(tx *sql.Tx) error {
		for _, pokemonType := range pokemonTypes {
			_, err := tx.Exec(query, pokemonType.PokeapiID, pokemonType.Slug, pokemonType.DisplayName)
			if err != nil {
				return fmt.Errorf("error inserting pokemonType=%v. %w", pokemonType, err)
			}
		}

		return nil
	})
}

type typeCSVRow struct {
	ID         int    `csv:"id"`
	Identifier string `csv:"identifier"`
}

func loadTypes(rows []typeCSVRow) []PokemonType {
	caser := cases.Title(language.English)
	pokemonTypes := make([]PokemonType, 0, len(rows))
	for _, row := range rows {
		pokemonTypes = append(pokemonTypes, PokemonType{
			PokeapiID:   row.ID,
			Slug:        row.Identifier,
			DisplayName: caser.String(row.Identifier),
		})
	}

	return pokemonTypes
}
