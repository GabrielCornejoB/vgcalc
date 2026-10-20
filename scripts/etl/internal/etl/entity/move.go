package entity

import "fmt"

type MoveFlags struct {
	IsContact bool
	IsPunch   bool
	IsBite    bool
	IsSound   bool
	IsPulse   bool
	IsBullet  bool
	IsSlicing bool
	IsWind    bool
}

type Move struct {
	PokeapiID   int
	Slug        string
	DisplayName string
	ShortEffect string
	Effect      string
	TypeID      int
	Category    string
	Power       int
	Accuracy    int
	PP          int
	Priority    int
	Target      string
	MoveFlags
	MinHits *int
	MaxHits *int
}

func (m Move) Validate() error {
	if m.PokeapiID <= 0 {
		return fmt.Errorf("move %q: pokeapi_id must be positive, got %d", m.Slug, m.PokeapiID)
	}
	if m.Slug == "" {
		return fmt.Errorf("move pokeapi_id=%d: empty slug", m.PokeapiID)
	}
	if m.TypeID == 0 {
		return fmt.Errorf("move %q: missing type", m.Slug)
	}
	switch m.Category {
	case "physical", "special", "status":
	default:
		return fmt.Errorf("move %q: unknown category %q", m.Slug, m.Category)
	}
	if m.Power < 0 {
		return fmt.Errorf("move %q: power must be >= 0, got %d", m.Slug, m.Power)
	}
	if m.PP < 0 {
		return fmt.Errorf("move %q: pp must be >= 0, got %d", m.Slug, m.PP)
	}
	if m.Accuracy < 0 || m.Accuracy > 100 {
		return fmt.Errorf("move %q: accuracy must be within 0-100, got %d", m.Slug, m.Accuracy)
	}
	if m.MinHits != nil && *m.MinHits < 1 {
		return fmt.Errorf("move %q: min_hits must be >= 1, got %d", m.Slug, *m.MinHits)
	}
	if m.MaxHits != nil && *m.MaxHits < 1 {
		return fmt.Errorf("move %q: max_hits must be >= 1, got %d", m.Slug, *m.MaxHits)
	}
	if m.MinHits != nil && m.MaxHits != nil && *m.MaxHits < *m.MinHits {
		return fmt.Errorf("move %q: max_hits %d is below min_hits %d", m.Slug, *m.MaxHits, *m.MinHits)
	}

	return nil
}
