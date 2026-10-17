package etl

import (
	"database/sql"
	"fmt"
	"strconv"

	"vgcalc/etl/internal/csv"
)

type TypeEfficacyStrategy struct {
	CSVPath string
	DB    	*sql.DB
}

type TypeEfficacy struct {
	AttackTypeId int
	TargetTypeId int
	Multiplier   float64
}

func (s TypeEfficacyStrategy) Run() error {
	records, err := csv.ReadCSVFile(s.CSVPath)
	if err != nil {
		return fmt.Errorf("TypeEfficacyStrategy: error reading csv file. %w", err)
	}

	typeEfficacies, err := parseTypeEfficacies(records)
	if err != nil {
		return fmt.Errorf("TypeEfficacyStrategy: %w", err)
	}

	typesMap, err := loadTypesIDMap(s.DB)
	if err != nil {
		return fmt.Errorf("TypeEfficacyStrategy: %w", err)
	}

	const query = `INSERT INTO vgcalc.type_efficacy (attack_type_id, target_type_id, multiplier) VALUES ($1, $2, $3)
	ON CONFLICT (attack_type_id, target_type_id) DO UPDATE SET multiplier = EXCLUDED.multiplier`

	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("TypeEfficacyStrategy: error starting transaction. %w", err)
	}
	defer tx.Rollback()

	for _, typeEfficacy := range typeEfficacies {
		attackId, attackIdExists := typesMap[typeEfficacy.AttackTypeId]
		targetId, targetIdExists := typesMap[typeEfficacy.TargetTypeId]
		// If type doesn't matches then continue. This is because some types are excluded (unknown, shadow, stellar)
		if !attackIdExists || !targetIdExists {
			continue
		}
		_, err := tx.Exec(query, attackId, targetId, typeEfficacy.Multiplier)
		if err != nil {
			return fmt.Errorf("TypeEfficacyStrategy: typeEfficacy=%v. %w", typeEfficacy, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("TypeEfficacyStrategy: error committing transaction. %w", err)
	}
	return nil
}

func parseTypeEfficacies(records [][]string) ([]TypeEfficacy, error) {
	typeEfficacies := make([]TypeEfficacy, 0, len(records))
	for _, record := range records {
		if len(record) < 3 {
			return nil, fmt.Errorf("parseTypeEfficacies(): invalid record %v", record)
		}
		attackTypeId, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("parseTypeEfficacies(): invalid damage_type_id=%q. %w", record[0], err)
		}
		targetTypeId, err := strconv.Atoi(record[1])
		if err != nil {
			return nil, fmt.Errorf("parseTypeEfficacies(): invalid target_type_id=%q. %w", record[1], err)
		}
		damageFactor, err := strconv.Atoi(record[2])
		if err != nil {
			return nil, fmt.Errorf("parseTypeEfficacies(): invalid damage_factor=%q. %w", record[2], err)
		}
		typeEfficacies = append(typeEfficacies, TypeEfficacy{
			AttackTypeId: attackTypeId,
			TargetTypeId: targetTypeId,
			Multiplier:   float64(damageFactor) / 100, // PokeAPI has 0, 50, 100, 200. New model has 0, 0.5, 1, 2.
		})
	}

	return typeEfficacies, nil
}

func loadTypesIDMap(db *sql.DB) (map[int]int, error) {
	rows, err := db.Query("SELECT id, pokeapi_id FROM vgcalc.types")
	if err != nil {
		return nil, fmt.Errorf("loadTypesIDMap(): failed to query existing types. %w", err)
	}
	defer rows.Close()
	typesMap := make(map[int]int)
	for rows.Next() {
		var dbID, pokeapiID int
		if err := rows.Scan(&dbID, &pokeapiID); err != nil {
			return nil, fmt.Errorf("loadTypesIDMap(): %w", err)
		}
		typesMap[pokeapiID] = dbID
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("loadTypesIDMap(): %w", err)
	}

	return typesMap, nil
}
