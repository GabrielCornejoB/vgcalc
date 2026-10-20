package entity

import "fmt"

type Stats struct {
	HP, Atk, Def, SpA, SpD, Spe int
}

type TypeSlots struct {
	Primary   int
	Secondary *int
}

type AbilitySlots struct {
	Slot1  int
	Slot2  *int
	Hidden *int
}

func (s Stats) Validate() error {
	stats := []struct {
		name  string
		value int
	}{{"hp", s.HP}, {"atk", s.Atk}, {"def", s.Def}, {"spa", s.SpA}, {"spd", s.SpD}, {"spe", s.Spe}}
	for _, stat := range stats {
		if stat.value < 1 {
			return fmt.Errorf("%s must be >= 1, got %d", stat.name, stat.value)
		}
	}

	return nil
}
