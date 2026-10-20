package etl

import (
	"database/sql"
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/entity"
	"vgcalc/etl/internal/etl/common"
)

type TypeStrategy struct {
	CSVPath string
	DB      *sql.DB
}

func NewTypeStrategy(db *sql.DB, dataDir string) TypeStrategy {
	return TypeStrategy{DB: db, CSVPath: dataDir + "types.csv"}
}

func (s TypeStrategy) Run() error {
	src, err := s.extract()
	if err != nil {
		return err
	}
	pokemonTypes, err := transformTypes(src)
	if err != nil {
		return err
	}

	return loadTypes(s.DB, pokemonTypes)
}

type typeCSVRow struct {
	ID         int    `csv:"id"`
	Identifier string `csv:"identifier"`
}

type typeSource struct {
	typeRows []typeCSVRow
}

func (s TypeStrategy) extract() (typeSource, error) {
	rows, err := csv.ReadFile[typeCSVRow](s.CSVPath)
	if err != nil {
		return typeSource{}, err
	}

	return typeSource{typeRows: rows}, nil
}

func transformTypes(src typeSource) ([]entity.Type, error) {
	caser := cases.Title(language.English)
	pokemonTypes := make([]entity.Type, 0, len(src.typeRows))
	for _, row := range src.typeRows {
		pokemonType := entity.Type{
			PokeapiID:   row.ID,
			Slug:        row.Identifier,
			DisplayName: caser.String(row.Identifier),
		}
		if err := pokemonType.Validate(); err != nil {
			return nil, err
		}
		pokemonTypes = append(pokemonTypes, pokemonType)
	}

	return pokemonTypes, nil
}

func loadTypes(db *sql.DB, pokemonTypes []entity.Type) error {
	const query = `INSERT INTO vgcalc.types (pokeapi_id, slug, display_name) VALUES ($1, $2, $3)
	ON CONFLICT (pokeapi_id) DO UPDATE SET slug = EXCLUDED.slug, display_name = EXCLUDED.display_name`

	return common.InTx(db, func(tx *sql.Tx) error {
		for _, pokemonType := range pokemonTypes {
			_, err := tx.Exec(query, pokemonType.PokeapiID, pokemonType.Slug, pokemonType.DisplayName)
			if err != nil {
				return fmt.Errorf("error inserting pokemonType=%v. %w", pokemonType, err)
			}
		}

		return nil
	})
}
