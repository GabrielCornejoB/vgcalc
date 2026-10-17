package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
)

type TypeEfficacyStrategy struct {
	CSVPath string
	DB      *sql.DB
}

type TypeEfficacy struct {
	AttackTypeId int
	TargetTypeId int
	Multiplier   float64
}

func NewTypeEfficacyStrategy(db *sql.DB, dataDir string) TypeEfficacyStrategy {
	return TypeEfficacyStrategy{DB: db, CSVPath: dataDir + "type_efficacy.csv"}
}

func (s TypeEfficacyStrategy) Run() error {
	rows, err := csv.LoadFile[typeEfficacyCSVRow](s.CSVPath)
	if err != nil {
		return fmt.Errorf("TypeEfficacyStrategy: %w", err)
	}

	typeEfficacies := parseTypeEfficacies(rows)

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

type typeEfficacyCSVRow struct {
	DamageTypeID int `csv:"damage_type_id"`
	TargetTypeID int `csv:"target_type_id"`
	DamageFactor int `csv:"damage_factor"`
}

func parseTypeEfficacies(rows []typeEfficacyCSVRow) []TypeEfficacy {
	typeEfficacies := make([]TypeEfficacy, 0, len(rows))
	for _, row := range rows {
		typeEfficacies = append(typeEfficacies, TypeEfficacy{
			AttackTypeId: row.DamageTypeID,
			TargetTypeId: row.TargetTypeID,
			Multiplier:   float64(row.DamageFactor) / 100, // PokeAPI has 0, 50, 100, 200. New model has 0, 0.5, 1, 2.
		})
	}

	return typeEfficacies
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
