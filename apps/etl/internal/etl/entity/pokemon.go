package entity

import "fmt"

type Pokemon struct {
	PokeapiID   int
	Slug        string
	DisplayName string
	Stats
	TypeSlots
	AbilitySlots
	Weight      float64
	FixedGender *string
}

func (p Pokemon) Validate() error {
	if p.PokeapiID <= 0 {
		return fmt.Errorf("pokemon %q: pokeapi_id must be positive, got %d", p.Slug, p.PokeapiID)
	}
	if p.Slug == "" {
		return fmt.Errorf("pokemon pokeapi_id=%d: empty slug", p.PokeapiID)
	}
	if p.Primary == 0 {
		return fmt.Errorf("pokemon %q: missing primary type", p.Slug)
	}
	if p.Slot1 == 0 {
		return fmt.Errorf("pokemon %q: missing ability1", p.Slug)
	}
	if err := p.Stats.Validate(); err != nil {
		return fmt.Errorf("pokemon %q: %w", p.Slug, err)
	}
	if p.Weight <= 0 {
		return fmt.Errorf("pokemon %q: weight must be > 0, got %v", p.Slug, p.Weight)
	}
	if p.FixedGender != nil {
		switch *p.FixedGender {
		case "male", "female", "genderless":
		default:
			return fmt.Errorf("pokemon %q: unknown fixed_gender %q", p.Slug, *p.FixedGender)
		}
	}

	return nil
}
