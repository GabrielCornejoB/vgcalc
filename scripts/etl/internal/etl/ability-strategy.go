package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
)

type AbilityStrategy struct {
	abilitiesCSVPath    string
	abilityNamesCSVPath string
	abilityProseCSVPath string
	DB                  *sql.DB
}

type Ability struct {
	PokeapiID   int
	Slug        string
	DisplayName string
	ShortEffect string
	Effect      string
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
	abilityRows, err := csv.LoadFile[abilityCSVRow](s.abilitiesCSVPath)
	if err != nil {
		return err
	}
	abilities, abilitiesMap := buildAbilities(abilityRows)

	nameRows, err := csv.LoadFile[abilityNameCSVRow](s.abilityNamesCSVPath)
	if err != nil {
		return err
	}
	applyAbilityNames(nameRows, abilities, abilitiesMap)

	proseRows, err := csv.LoadFile[abilityProseCSVRow](s.abilityProseCSVPath)
	if err != nil {
		return err
	}
	applyAbilityEffects(proseRows, abilities, abilitiesMap)

	const query = `INSERT INTO vgcalc.abilities (pokeapi_id, slug, display_name, short_effect, effect)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (pokeapi_id) DO UPDATE SET
		slug = EXCLUDED.slug,
		display_name = EXCLUDED.display_name,
		short_effect = EXCLUDED.short_effect,
		effect = EXCLUDED.effect`

	return inTx(s.DB, func(tx *sql.Tx) error {
		for _, ability := range abilities {
			_, err := tx.Exec(query, ability.PokeapiID, ability.Slug, ability.DisplayName, ability.ShortEffect, ability.Effect)
			if err != nil {
				return fmt.Errorf("error inserting ability %v. %w", ability, err)
			}
		}

		return nil
	})
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

func buildAbilities(rows []abilityCSVRow) ([]Ability, map[int]int) {
	abilities := make([]Ability, 0, len(rows))
	abilitiesMap := make(map[int]int, len(rows))
	for i, row := range rows {
		abilities = append(abilities, Ability{PokeapiID: row.ID, Slug: row.Identifier})
		abilitiesMap[row.ID] = i
	}

	return abilities, abilitiesMap
}

func applyAbilityNames(rows []abilityNameCSVRow, abilities []Ability, abilitiesMap map[int]int) {
	for _, row := range rows {
		// If language id is != 9 (english), then skip the record
		if row.LocalLanguageID != pokeapiLanguageID {
			continue
		}
		abilityIndex, exists := abilitiesMap[row.AbilityID]
		if !exists {
			continue
		}
		abilities[abilityIndex].DisplayName = row.Name
	}
}

func applyAbilityEffects(rows []abilityProseCSVRow, abilities []Ability, abilitiesMap map[int]int) {
	for _, row := range rows {
		// If language id is != 9 (english), then skip the record
		if row.LocalLanguageID != pokeapiLanguageID {
			continue
		}
		abilityIndex, exists := abilitiesMap[row.AbilityID]
		if !exists {
			continue
		}
		abilities[abilityIndex].ShortEffect = row.ShortEffect
		abilities[abilityIndex].Effect = row.Effect
	}
}
