package etl

import (
	"database/sql"
	"fmt"
	"strconv"
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

const pokeapiLanguageID = 9

func (s AbilityStrategy) Run() error {
	records, err := csv.ReadCSVFile(s.abilitiesCSVPath)
	if err != nil {
		return fmt.Errorf("AbilityStrategy: error reading csv file. %w", err)
	}
	abilities, abilitiesMap, err := loadAbilities(records)
	if err != nil {
		return fmt.Errorf("AbilityStrategy: %w", err)
	}

	records, err = csv.ReadCSVFile(s.abilityNamesCSVPath)
	if err != nil {
		return fmt.Errorf("AbilityStrategy: error reading csv file. %w", err)
	}
	err = loadAbilityNames(records, abilities, abilitiesMap)
	if err != nil {
		return fmt.Errorf("AbilityStrategy: %w", err)
	}

	records, err = csv.ReadCSVFile(s.abilityProseCSVPath)
	if err != nil {
		return fmt.Errorf("AbilityStrategy: error reading csv file. %w", err)
	}
	err = loadAbilityEffects(records, abilities, abilitiesMap)
	if err != nil {
		return fmt.Errorf("AbilityStrategy: %w", err)
	}

	const query = `INSERT INTO vgcalc.abilities (pokeapi_id, slug, display_name, short_effect, effect) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (pokeapi_id) DO UPDATE SET slug = EXCLUDED.slug, display_name = EXCLUDED.display_name, short_effect = EXCLUDED.short_effect, effect = EXCLUDED.effect`

	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("AbilityStrategy: error starting transaction. %w", err)
	}
	defer tx.Rollback()

	for _, ability := range abilities {
		_, err := tx.Exec(query, ability.PokeapiID, ability.Slug, ability.DisplayName, ability.ShortEffect, ability.Effect)
		if err != nil {
			return fmt.Errorf("AbilityStrategy: error inserting ability %v. %w", ability, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("AbilityStrategy: error commiting transaction. %w", err)
	}

	return nil
}

func loadAbilities(records [][]string) ([]Ability, map[int]int, error) {
	abilities := make([]Ability, 0, len(records))
	abilitiesMap := make(map[int]int)
	for i, record := range records {
		if len(record) < 2 {
			return nil, nil, fmt.Errorf("loadAbilities(): invalid record=%v", record)
		}
		pokeapiID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, nil, fmt.Errorf("loadAbilities(): invalid pokeapi_id=%q. %w", record[0], err)
		}
		abilities = append(abilities, Ability{PokeapiID: pokeapiID, Slug: record[1]})
		abilitiesMap[pokeapiID] = i
	}

	return abilities, abilitiesMap, nil
}

func loadAbilityNames(records [][]string, abilities []Ability, abilitiesMap map[int]int) error {
	for _, record := range records {
		if len(record) < 3 {
			return fmt.Errorf("loadAbilityNames(): invalid record=%v", record)
		}
		pokeapiID, err := strconv.Atoi(record[0])
		if err != nil {
			return fmt.Errorf("loadAbilityNames(): invalid pokeapi_id=%q. %w", record[0], err)
		}
		localLanguageID, err := strconv.Atoi(record[1])
		if err != nil {
			return fmt.Errorf("loadAbilityNames(): invalid local_language_id=%q. %w", record[1], err)
		}
		// If language id is != 9 (english), then skip the record
		if localLanguageID != pokeapiLanguageID {
			continue
		}
		abilityIndex, exists := abilitiesMap[pokeapiID]
		if !exists {
			continue
		}
		abilities[abilityIndex].DisplayName = record[2]
	}
	return nil
}

func loadAbilityEffects(records [][]string, abilities []Ability, abilitiesMap map[int]int) error {
	for _, record := range records {
		if len(record) < 4 {
			return fmt.Errorf("AbilityStrategy: invalid record=%v", record)
		}
		pokeapiID, err := strconv.Atoi(record[0])
		if err != nil {
			return fmt.Errorf("AbilityStrategy: invalid pokeapi_id=%q. %w", record[0], err)
		}
		localLanguageID, err := strconv.Atoi(record[1])
		if err != nil {
			return fmt.Errorf("AbilityStrategy: invalid local_language_id=%q. %w", record[1], err)
		}
		// If language id is != 9 (english), then skip the record
		if localLanguageID != pokeapiLanguageID {
			continue
		}
		abilityIndex, exists := abilitiesMap[pokeapiID]
		if !exists {
			continue
		}
		abilities[abilityIndex].ShortEffect = record[2]
		abilities[abilityIndex].Effect = record[3]
	}
	return nil
}
