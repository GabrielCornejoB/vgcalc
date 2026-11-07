package entity

import "fmt"

type PokemonMove struct {
	PokemonID int
	MoveID    int
}

func (p PokemonMove) Validate() error {
	if p.PokemonID == 0 {
		return fmt.Errorf("pokemon move move_id=%d: missing pokemon_id", p.MoveID)
	}
	if p.MoveID == 0 {
		return fmt.Errorf("pokemon move pokemon_id=%d: missing move_id", p.PokemonID)
	}

	return nil
}
