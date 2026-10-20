package entity

import "fmt"

type Nature struct {
	PokeapiID   int
	Slug        string
	DisplayName string
	PlusStat    string
	MinusStat   string
}

func (n Nature) Validate() error {
	if n.PokeapiID <= 0 {
		return fmt.Errorf("nature %q: pokeapi_id must be positive, got %d", n.Slug, n.PokeapiID)
	}
	if n.Slug == "" {
		return fmt.Errorf("nature pokeapi_id=%d: empty slug", n.PokeapiID)
	}
	if !isStatSlug(n.PlusStat) {
		return fmt.Errorf("nature %q: invalid plus_stat %q", n.Slug, n.PlusStat)
	}
	if !isStatSlug(n.MinusStat) {
		return fmt.Errorf("nature %q: invalid minus_stat %q", n.Slug, n.MinusStat)
	}

	return nil
}

func isStatSlug(s string) bool {
	switch s {
	case "atk", "def", "spa", "spd", "spe":
		return true
	}

	return false
}
