package entity

import "fmt"

type Type struct {
	PokeapiID   int
	Slug        string
	DisplayName string
}

func (t Type) Validate() error {
	if t.PokeapiID <= 0 {
		return fmt.Errorf("type %q: pokeapi_id must be positive, got %d", t.Slug, t.PokeapiID)
	}
	if t.Slug == "" {
		return fmt.Errorf("type pokeapi_id=%d: empty slug", t.PokeapiID)
	}

	return nil
}
