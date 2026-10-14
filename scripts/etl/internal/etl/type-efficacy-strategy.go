package etl

import (
	"database/sql"
	"fmt"
	"strconv"

	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/env"
)

type TypeEfficacyStrategy struct {
	filePathUrl string
	db          *sql.DB
}

func NewTypeEfficacyStrategy(db *sql.DB) (TypeEfficacyStrategy, error) {
	filePath, err := env.GetEnvByKey("PATH_TYPES_EFFICACY_CSV")
	if err != nil {
		return TypeEfficacyStrategy{}, fmt.Errorf("NewTypeEfficacyStrategy(): %w", err)
	}
	return TypeEfficacyStrategy{filePath, db}, nil
}

type TypeEfficacy struct {
	AttackTypeId int
	TargetTypeId int
	Multiplier   float64
}

func (strategy TypeEfficacyStrategy) extract() ([][]string, error) {
	records, err := csv.ReadCsvFile(strategy.filePathUrl)
	if err != nil {
		return nil, fmt.Errorf("TypeEfficacyStrategy.extract(): %w", err)
	}

	return records, nil
}

func (strategy TypeEfficacyStrategy) transform(records [][]string) ([]TypeEfficacy, error) {
	records = records[1:]
	typeEfficacies := make([]TypeEfficacy, 0, len(records))

	for _, record := range records {
		attackTypeId, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("TypeEfficacyStrategy.transform(): invalid damage_type_id=%q. %w", record[0], err)
		}

		targetTypeId, err := strconv.Atoi(record[1])
		if err != nil {
			return nil, fmt.Errorf("TypeEfficacyStrategy.transform(): invalid target_type_id=%q. %w", record[1], err)
		}

		damageFactor, err := strconv.Atoi(record[2])
		if err != nil {
			return nil, fmt.Errorf("TypeEfficacyStrategy.transform(): invalid damage_factor=%q. %w", record[2], err)
		}

		typeEfficacies = append(typeEfficacies, TypeEfficacy{
			AttackTypeId: attackTypeId,
			TargetTypeId: targetTypeId,
			Multiplier:   float64(damageFactor) / 100,
		})
	}

	return typeEfficacies, nil
}

func (strategy TypeEfficacyStrategy) load(typeEfficacies []TypeEfficacy) error {
	rows, err := strategy.db.Query("SELECT id, pokeapi_id FROM vgcalc.types")
	if err != nil {
		return fmt.Errorf("TypeEfficacyStrategy.load(): failed to query existing types. %w", err)
	}
	defer rows.Close()
	typesMap := make(map[int]int)
	for rows.Next() {
		var dbId, pokeapiId int
		if err := rows.Scan(&dbId, &pokeapiId); err != nil {
			return fmt.Errorf("TypeEfficacyStrategy.load(): %w", err)
		}
		typesMap[pokeapiId] = dbId
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("TypeEfficacyStrategy.load(): %w", err)
	}

	const query = "INSERT INTO vgcalc.type_efficacy (attack_type_id, target_type_id, multiplier) VALUES ($1, $2, $3)"

	for _, typeEfficacy := range typeEfficacies {
		attackId, attackIdExists := typesMap[typeEfficacy.AttackTypeId]
		targetId, targetIdExists := typesMap[typeEfficacy.TargetTypeId]
		if !attackIdExists || !targetIdExists {
			continue
		}
		_, err := strategy.db.Exec(query, attackId, targetId, typeEfficacy.Multiplier)
		if err != nil {
			return fmt.Errorf("TypeEfficacyStrategy.load(): typeEfficacy=%v. %w", typeEfficacy, err)
		}
	}

	return nil
}

func (strategy TypeEfficacyStrategy) Run() error {
	records, err := strategy.extract()
	if err != nil {
		return fmt.Errorf("TypeEfficacyStrategy.Run(): %w", err)
	}

	typeEfficacies, err := strategy.transform(records)
	if err != nil {
		return fmt.Errorf("TypeEfficacyStrategy.Run(): %w", err)
	}

	err = strategy.load(typeEfficacies)
	if err != nil {
		return fmt.Errorf("TypeEfficacyStrategy.Run(): %w", err)
	}

	return nil
}
