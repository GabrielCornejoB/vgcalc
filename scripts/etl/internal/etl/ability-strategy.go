package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/etl/common"
	"vgcalc/etl/internal/etl/entity"
)

type AbilityStrategy struct {
	abilitiesCSVPath    string
	abilityNamesCSVPath string
	abilityProseCSVPath string
	DB                  *sql.DB
}

func NewAbilityStrategy(db *sql.DB, dataDir string) AbilityStrategy {
	return AbilityStrategy{
		DB:                  db,
		abilitiesCSVPath:    dataDir + "abilities.csv",
		abilityNamesCSVPath: dataDir + "ability_names.csv",
		abilityProseCSVPath: dataDir + "ability_prose.csv",
	}
}

func (s AbilityStrategy) Run() error {
	src, err := s.extract()
	if err != nil {
		return err
	}
	abilities, err := transformAbilities(src)
	if err != nil {
		return err
	}

	return loadAbilities(s.DB, abilities)
}

type abilityCSVRow struct {
	ID         int    `csv:"id"`
	Identifier string `csv:"identifier"`
}

type abilityNameCSVRow struct {
	AbilityID       int    `csv:"ability_id"`
	LocalLanguageID int    `csv:"local_language_id"`
	Name            string `csv:"name"`
}

type abilityProseCSVRow struct {
	AbilityID       int    `csv:"ability_id"`
	LocalLanguageID int    `csv:"local_language_id"`
	ShortEffect     string `csv:"short_effect"`
	Effect          string `csv:"effect"`
}

type abilitySource struct {
	abilityRows []abilityCSVRow
	names       map[int]string
	prose       map[int]abilityProseCSVRow
}

func (s AbilityStrategy) extract() (abilitySource, error) {
	var src abilitySource
	var err error

	if src.abilityRows, err = csv.ReadFile[abilityCSVRow](s.abilitiesCSVPath); err != nil {
		return src, err
	}
	nameRows, err := csv.ReadFile[abilityNameCSVRow](s.abilityNamesCSVPath)
	if err != nil {
		return src, err
	}
	proseRows, err := csv.ReadFile[abilityProseCSVRow](s.abilityProseCSVPath)
	if err != nil {
		return src, err
	}

	src.names = make(map[int]string, len(nameRows))
	for _, row := range nameRows {
		if row.LocalLanguageID == common.PokeapiLanguageID {
			src.names[row.AbilityID] = row.Name
		}
	}
	src.prose = make(map[int]abilityProseCSVRow, len(proseRows))
	for _, row := range proseRows {
		if row.LocalLanguageID == common.PokeapiLanguageID {
			src.prose[row.AbilityID] = row
		}
	}

	return src, nil
}

func transformAbilities(src abilitySource) ([]entity.Ability, error) {
	abilities := make([]entity.Ability, 0, len(src.abilityRows))
	for _, row := range src.abilityRows {
		prose := src.prose[row.ID]
		ability := entity.Ability{
			PokeapiID:   row.ID,
			Slug:        row.Identifier,
			DisplayName: src.names[row.ID],
			ShortEffect: prose.ShortEffect,
			Effect:      prose.Effect,
		}
		if err := ability.Validate(); err != nil {
			return nil, err
		}
		abilities = append(abilities, ability)
	}

	return abilities, nil
}

func loadAbilities(db *sql.DB, abilities []entity.Ability) error {
	const query = `INSERT INTO vgcalc.abilities (pokeapi_id, slug, display_name, short_effect, effect)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (pokeapi_id) DO UPDATE SET
		slug = EXCLUDED.slug,
		display_name = EXCLUDED.display_name,
		short_effect = EXCLUDED.short_effect,
		effect = EXCLUDED.effect`

	return common.InTx(db, func(tx *sql.Tx) error {
		for _, ability := range abilities {
			_, err := tx.Exec(query,
				ability.PokeapiID, ability.Slug, ability.DisplayName,
				ability.ShortEffect, ability.Effect)
			if err != nil {
				return fmt.Errorf("error inserting ability %v. %w", ability, err)
			}
		}

		return nil
	})
}
