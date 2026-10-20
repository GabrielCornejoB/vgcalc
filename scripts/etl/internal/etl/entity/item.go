package entity

import "fmt"

type Item struct {
	PokeapiID   int
	Slug        string
	DisplayName string
	ShortEffect string
	Effect      string
	FlingPower  int
}

func (i Item) Validate() error {
	if i.PokeapiID <= 0 {
		return fmt.Errorf("item %q: pokeapi_id must be positive, got %d", i.Slug, i.PokeapiID)
	}
	if i.Slug == "" {
		return fmt.Errorf("item pokeapi_id=%d: empty slug", i.PokeapiID)
	}
	if i.FlingPower < 0 {
		return fmt.Errorf("item %q: fling_power must be >= 0, got %d", i.Slug, i.FlingPower)
	}

	return nil
}
