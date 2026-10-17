package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
)

type NatureStrategy struct {
	naturesCSVPath     string
	natureNamesCSVPath string
	DB                 *sql.DB
}

type Nature struct {
	PokeapiID   int
	Slug        string
	DisplayName string
	PlusStat    string
	MinusStat   string
}

func NewNatureStrategy(db *sql.DB, dataDir string) NatureStrategy {
	return NatureStrategy{
		DB:                 db,
		naturesCSVPath:     dataDir + "natures.csv",
		natureNamesCSVPath: dataDir + "nature_names.csv",
	}
}

var statIDToSlug = map[int]string{
	2: "atk",
	3: "def",
	4: "spa",
	5: "spd",
	6: "spe",
}

func (s NatureStrategy) Run() error {
	natureRows, err := csv.LoadFile[natureCSVRow](s.naturesCSVPath)
	if err != nil {
		return fmt.Errorf("NatureStrategy: %w", err)
	}
	natures, naturesMap, err := buildNatures(natureRows)
	if err != nil {
		return fmt.Errorf("NatureStrategy: %w", err)
	}

	nameRows, err := csv.LoadFile[natureNameCSVRow](s.natureNamesCSVPath)
	if err != nil {
		return fmt.Errorf("NatureStrategy: %w", err)
	}
	applyNatureNames(nameRows, natures, naturesMap)

	const query = `INSERT INTO vgcalc.natures (pokeapi_id, slug, display_name, plus_stat, minus_stat)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (pokeapi_id) DO UPDATE SET
		slug = EXCLUDED.slug,
		display_name = EXCLUDED.display_name,
		plus_stat = EXCLUDED.plus_stat,
		minus_stat = EXCLUDED.minus_stat`

	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("NatureStrategy: error starting transaction. %w", err)
	}
	defer tx.Rollback()

	for _, nature := range natures {
		_, err := tx.Exec(query, nature.PokeapiID, nature.Slug, nature.DisplayName, nature.PlusStat, nature.MinusStat)
		if err != nil {
			return fmt.Errorf("NatureStrategy: error inserting nature %v. %w", nature, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("NatureStrategy: error committing transaction. %w", err)
	}

	return nil
}

type natureCSVRow struct {
	ID              int    `csv:"id"`
	Identifier      string `csv:"identifier"`
	DecreasedStatID int    `csv:"decreased_stat_id"`
	IncreasedStatID int    `csv:"increased_stat_id"`
}

type natureNameCSVRow struct {
	NatureID        int    `csv:"nature_id"`
	LocalLanguageID int    `csv:"local_language_id"`
	Name            string `csv:"name"`
}

func buildNatures(rows []natureCSVRow) ([]Nature, map[int]int, error) {
	natures := make([]Nature, 0, len(rows))
	naturesMap := make(map[int]int, len(rows))
	for i, row := range rows {
		minusStat, ok := statIDToSlug[row.DecreasedStatID]
		if !ok {
			return nil, nil, fmt.Errorf("buildNatures(): unknown decreased_stat_id=%d", row.DecreasedStatID)
		}
		plusStat, ok := statIDToSlug[row.IncreasedStatID]
		if !ok {
			return nil, nil, fmt.Errorf("buildNatures(): unknown increased_stat_id=%d", row.IncreasedStatID)
		}
		natures = append(natures, Nature{PokeapiID: row.ID, Slug: row.Identifier, PlusStat: plusStat, MinusStat: minusStat})
		naturesMap[row.ID] = i
	}

	return natures, naturesMap, nil
}

func applyNatureNames(rows []natureNameCSVRow, natures []Nature, naturesMap map[int]int) {
	for _, row := range rows {
		// If language id is != 9 (english), then skip the record
		if row.LocalLanguageID != pokeapiLanguageID {
			continue
		}
		natureIndex, exists := naturesMap[row.NatureID]
		if !exists {
			continue
		}
		natures[natureIndex].DisplayName = row.Name
	}
}
