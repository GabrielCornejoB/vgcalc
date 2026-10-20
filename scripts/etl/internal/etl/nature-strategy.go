package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/etl/common"
	"vgcalc/etl/internal/etl/entity"
)

type NatureStrategy struct {
	naturesCSVPath     string
	natureNamesCSVPath string
	DB                 *sql.DB
}

func NewNatureStrategy(db *sql.DB, dataDir string) NatureStrategy {
	return NatureStrategy{
		DB:                 db,
		naturesCSVPath:     dataDir + "natures.csv",
		natureNamesCSVPath: dataDir + "nature_names.csv",
	}
}

var statIDToSlug = map[int]string{
	common.StatAtk: "atk",
	common.StatDef: "def",
	common.StatSpA: "spa",
	common.StatSpD: "spd",
	common.StatSpe: "spe",
}

func (s NatureStrategy) Run() error {
	src, err := s.extract()
	if err != nil {
		return err
	}
	natures, err := transformNatures(src)
	if err != nil {
		return err
	}

	return loadNatures(s.DB, natures)
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

type natureSource struct {
	natureRows []natureCSVRow
	names      map[int]string
}

func (s NatureStrategy) extract() (natureSource, error) {
	var src natureSource
	var err error

	if src.natureRows, err = csv.ReadFile[natureCSVRow](s.naturesCSVPath); err != nil {
		return src, err
	}
	nameRows, err := csv.ReadFile[natureNameCSVRow](s.natureNamesCSVPath)
	if err != nil {
		return src, err
	}

	src.names = make(map[int]string, len(nameRows))
	for _, row := range nameRows {
		if row.LocalLanguageID == common.PokeapiLanguageID {
			src.names[row.NatureID] = row.Name
		}
	}

	return src, nil
}

func transformNatures(src natureSource) ([]entity.Nature, error) {
	natures := make([]entity.Nature, 0, len(src.natureRows))
	for _, row := range src.natureRows {
		minusStat, ok := statIDToSlug[row.DecreasedStatID]
		if !ok {
			return nil, fmt.Errorf("transformNatures(): unknown decreased_stat_id=%d", row.DecreasedStatID)
		}
		plusStat, ok := statIDToSlug[row.IncreasedStatID]
		if !ok {
			return nil, fmt.Errorf("transformNatures(): unknown increased_stat_id=%d", row.IncreasedStatID)
		}
		nature := entity.Nature{
			PokeapiID:   row.ID,
			Slug:        row.Identifier,
			DisplayName: src.names[row.ID],
			PlusStat:    plusStat,
			MinusStat:   minusStat,
		}
		if err := nature.Validate(); err != nil {
			return nil, err
		}
		natures = append(natures, nature)
	}

	return natures, nil
}

func loadNatures(db *sql.DB, natures []entity.Nature) error {
	const query = `INSERT INTO vgcalc.natures (pokeapi_id, slug, display_name, plus_stat, minus_stat)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (pokeapi_id) DO UPDATE SET
		slug = EXCLUDED.slug,
		display_name = EXCLUDED.display_name,
		plus_stat = EXCLUDED.plus_stat,
		minus_stat = EXCLUDED.minus_stat`

	return common.InTx(db, func(tx *sql.Tx) error {
		for _, nature := range natures {
			_, err := tx.Exec(query,
				nature.PokeapiID, nature.Slug, nature.DisplayName, nature.PlusStat, nature.MinusStat)
			if err != nil {
				return fmt.Errorf("error inserting nature %v. %w", nature, err)
			}
		}

		return nil
	})
}
