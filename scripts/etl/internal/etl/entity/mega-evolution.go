package entity

import "fmt"

type MegaEvolution struct {
	PokeapiID      int
	BasePokemonID  int
	Slug           string
	RequiredItemID *int
	RequiredMoveID *int
	AbilityID      int
	Stats
	TypeSlots
	Weight float64
}

func (m MegaEvolution) Validate() error {
	if m.PokeapiID <= 0 {
		return fmt.Errorf("mega evolution %q: pokeapi_id must be positive, got %d", m.Slug, m.PokeapiID)
	}
	if m.Slug == "" {
		return fmt.Errorf("mega evolution pokeapi_id=%d: empty slug", m.PokeapiID)
	}
	if m.BasePokemonID == 0 {
		return fmt.Errorf("mega evolution %q: missing base pokemon", m.Slug)
	}
	if (m.RequiredItemID == nil) == (m.RequiredMoveID == nil) {
		return fmt.Errorf("mega evolution %q: exactly one of required item or required move must be set", m.Slug)
	}
	if m.AbilityID == 0 {
		return fmt.Errorf("mega evolution %q: missing ability", m.Slug)
	}
	if m.Primary == 0 {
		return fmt.Errorf("mega evolution %q: missing primary type", m.Slug)
	}
	if err := m.Stats.Validate(); err != nil {
		return fmt.Errorf("mega evolution %q: %w", m.Slug, err)
	}
	if m.Weight <= 0 {
		return fmt.Errorf("mega evolution %q: weight must be > 0, got %v", m.Slug, m.Weight)
	}

	return nil
}
