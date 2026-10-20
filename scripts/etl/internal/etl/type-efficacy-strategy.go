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
		return err
	}

	typeEfficacies := parseTypeEfficacies(rows)

	typesMap, err := loadPokeapiIDMap(s.DB, vgcalcTypes)
	if err != nil {
		return err
	}

	const query = `INSERT INTO vgcalc.type_efficacy (attack_type_id, target_type_id, multiplier) VALUES ($1, $2, $3)
	ON CONFLICT (attack_type_id, target_type_id) DO UPDATE SET multiplier = EXCLUDED.multiplier`

	return inTx(s.DB, func(tx *sql.Tx) error {
		for _, typeEfficacy := range typeEfficacies {
			attackId, attackIdExists := typesMap[typeEfficacy.AttackTypeId]
			if !attackIdExists {
				return fmt.Errorf("unknown attack_type_id=%d", typeEfficacy.AttackTypeId)
			}
			targetId, targetIdExists := typesMap[typeEfficacy.TargetTypeId]
			if !targetIdExists {
				return fmt.Errorf("unknown target_type_id=%d", typeEfficacy.TargetTypeId)
			}
			_, err := tx.Exec(query, attackId, targetId, typeEfficacy.Multiplier)
			if err != nil {
				return fmt.Errorf("typeEfficacy=%v. %w", typeEfficacy, err)
			}
		}

		return nil
	})
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
