package entity

import "fmt"

type Ability struct {
	PokeapiID   int
	Slug        string
	DisplayName string
	ShortEffect string
	Effect      string
}

func (a Ability) Validate() error {
	if a.PokeapiID <= 0 {
		return fmt.Errorf("ability %q: pokeapi_id must be positive, got %d", a.Slug, a.PokeapiID)
	}
	if a.Slug == "" {
		return fmt.Errorf("ability pokeapi_id=%d: empty slug", a.PokeapiID)
	}

	return nil
}
