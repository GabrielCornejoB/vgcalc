package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/entity"
)

type TypeEfficacyStrategy struct {
	CSVPath string
	DB      *sql.DB
}

func NewTypeEfficacyStrategy(db *sql.DB, dataDir string) TypeEfficacyStrategy {
	return TypeEfficacyStrategy{DB: db, CSVPath: dataDir + "type_efficacy.csv"}
}

func (s TypeEfficacyStrategy) Run() error {
	src, err := s.extract()
	if err != nil {
		return err
	}
	typeEfficacies, err := transformTypeEfficacies(src)
	if err != nil {
		return err
	}

	return loadTypeEfficacies(s.DB, typeEfficacies)
}

type typeEfficacyCSVRow struct {
	DamageTypeID int `csv:"damage_type_id"`
	TargetTypeID int `csv:"target_type_id"`
	DamageFactor int `csv:"damage_factor"`
}

type typeEfficacySource struct {
	efficacyRows []typeEfficacyCSVRow
	typesMap     map[int]int
}

func (s TypeEfficacyStrategy) extract() (typeEfficacySource, error) {
	var src typeEfficacySource
	var err error

	if src.efficacyRows, err = csv.ReadFile[typeEfficacyCSVRow](s.CSVPath); err != nil {
		return src, err
	}
	if src.typesMap, err = queryPokeapiIDMap(s.DB, vgcalcTypes); err != nil {
		return src, err
	}

	return src, nil
}

func transformTypeEfficacies(src typeEfficacySource) ([]entity.TypeEfficacy, error) {
	typeEfficacies := make([]entity.TypeEfficacy, 0, len(src.efficacyRows))
	for _, row := range src.efficacyRows {
		attackID, ok := src.typesMap[row.DamageTypeID]
		if !ok {
			return nil, fmt.Errorf("unknown attack_type_id=%d", row.DamageTypeID)
		}
		targetID, ok := src.typesMap[row.TargetTypeID]
		if !ok {
			return nil, fmt.Errorf("unknown target_type_id=%d", row.TargetTypeID)
		}
		typeEfficacy := entity.TypeEfficacy{
			AttackTypeID: attackID,
			TargetTypeID: targetID,
			Multiplier:   float64(row.DamageFactor) / 100, // PokeAPI has 0, 50, 100, 200. New model has 0, 0.5, 1, 2.
		}
		if err := typeEfficacy.Validate(); err != nil {
			return nil, err
		}
		typeEfficacies = append(typeEfficacies, typeEfficacy)
	}

	return typeEfficacies, nil
}

func loadTypeEfficacies(db *sql.DB, typeEfficacies []entity.TypeEfficacy) error {
	const query = `INSERT INTO vgcalc.type_efficacy (attack_type_id, target_type_id, multiplier) VALUES ($1, $2, $3)
	ON CONFLICT (attack_type_id, target_type_id) DO UPDATE SET multiplier = EXCLUDED.multiplier`

	return inTx(db, func(tx *sql.Tx) error {
		for _, typeEfficacy := range typeEfficacies {
			_, err := tx.Exec(query, typeEfficacy.AttackTypeID, typeEfficacy.TargetTypeID, typeEfficacy.Multiplier)
			if err != nil {
				return fmt.Errorf("typeEfficacy=%v. %w", typeEfficacy, err)
			}
		}

		return nil
	})
}
