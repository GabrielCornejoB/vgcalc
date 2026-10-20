package entity

import "fmt"

type TypeEfficacy struct {
	AttackTypeID int
	TargetTypeID int
	Multiplier   float64
}

func (t TypeEfficacy) Validate() error {
	if t.AttackTypeID == 0 {
		return fmt.Errorf("type efficacy: missing attack_type_id")
	}
	if t.TargetTypeID == 0 {
		return fmt.Errorf("type efficacy attack_type_id=%d: missing target_type_id", t.AttackTypeID)
	}
	switch t.Multiplier {
	case 0, 0.5, 1, 2:
		return nil
	}

	return fmt.Errorf("type efficacy attack_type_id=%d target_type_id=%d: invalid multiplier %v",
		t.AttackTypeID, t.TargetTypeID, t.Multiplier)
}
