package etl

import (
	"database/sql"
	"fmt"
	"strconv"
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

// statIDToSlug maps PokeAPI stat IDs to the model's inline stat enum.
// Only battle stats 2-6 appear in natures.csv; hp (1) never does.
var statIDToSlug = map[int]string{
	2: "atk",
	3: "def",
	4: "spa",
	5: "spd",
	6: "spe",
}

func (s NatureStrategy) Run() error {
	records, err := csv.ReadCSVFile(s.naturesCSVPath)
	if err != nil {
		return fmt.Errorf("NatureStrategy: error reading csv file. %w", err)
	}
	natures, naturesMap, err := loadNatures(records)
	if err != nil {
		return fmt.Errorf("NatureStrategy: %w", err)
	}

	records, err = csv.ReadCSVFile(s.natureNamesCSVPath)
	if err != nil {
		return fmt.Errorf("NatureStrategy: error reading csv file. %w", err)
	}
	err = loadNatureNames(records, natures, naturesMap)
	if err != nil {
		return fmt.Errorf("NatureStrategy: %w", err)
	}

	const query = `INSERT INTO vgcalc.natures (pokeapi_id, slug, display_name, plus_stat, minus_stat) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (pokeapi_id) DO UPDATE SET slug = EXCLUDED.slug, display_name = EXCLUDED.display_name, plus_stat = EXCLUDED.plus_stat, minus_stat = EXCLUDED.minus_stat`

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

func loadNatures(records [][]string) ([]Nature, map[int]int, error) {
	natures := make([]Nature, 0, len(records))
	naturesMap := make(map[int]int)
	for i, record := range records {
		if len(record) < 4 {
			return nil, nil, fmt.Errorf("loadNatures(): invalid record=%v", record)
		}
		pokeapiID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, nil, fmt.Errorf("loadNatures(): invalid pokeapi_id=%q. %w", record[0], err)
		}
		decreasedStatID, err := strconv.Atoi(record[2])
		if err != nil {
			return nil, nil, fmt.Errorf("loadNatures(): invalid decreased_stat_id=%q. %w", record[2], err)
		}
		increasedStatID, err := strconv.Atoi(record[3])
		if err != nil {
			return nil, nil, fmt.Errorf("loadNatures(): invalid increased_stat_id=%q. %w", record[3], err)
		}
		minusStat, ok := statIDToSlug[decreasedStatID]
		if !ok {
			return nil, nil, fmt.Errorf("loadNatures(): unknown decreased_stat_id=%q", record[2])
		}
		plusStat, ok := statIDToSlug[increasedStatID]
		if !ok {
			return nil, nil, fmt.Errorf("loadNatures(): unknown increased_stat_id=%q", record[3])
		}
		natures = append(natures, Nature{PokeapiID: pokeapiID, Slug: record[1], PlusStat: plusStat, MinusStat: minusStat})
		naturesMap[pokeapiID] = i
	}

	return natures, naturesMap, nil
}

func loadNatureNames(records [][]string, natures []Nature, naturesMap map[int]int) error {
	for _, record := range records {
		if len(record) < 3 {
			return fmt.Errorf("loadNatureNames(): invalid record=%v", record)
		}
		pokeapiID, err := strconv.Atoi(record[0])
		if err != nil {
			return fmt.Errorf("loadNatureNames(): invalid pokeapi_id=%q. %w", record[0], err)
		}
		localLanguageID, err := strconv.Atoi(record[1])
		if err != nil {
			return fmt.Errorf("loadNatureNames(): invalid local_language_id=%q. %w", record[1], err)
		}
		// If language id is != 9 (english), then skip the record
		if localLanguageID != pokeapiLanguageID {
			continue
		}
		natureIndex, exists := naturesMap[pokeapiID]
		if !exists {
			continue
		}
		natures[natureIndex].DisplayName = record[2]
	}
	return nil
}
