package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
)

const (
	// version_group_id 32 is "champions" (25 is "scarlet-violet").
	championsVersionGroupID = 32
	// pokemon_move_method_id 12 is "train", the only method present in that
	// version group (older groups use 1 level-up, 2 egg, 3 tutor, 4 machine).
	trainMoveMethodID = 12
)

type PokemonMovesStrategy struct {
	pokemonMovesCSVPath string
	DB                  *sql.DB
}

type pokemonMoveCSVRow struct {
	PokemonID      int `csv:"pokemon_id"`
	VersionGroupID int `csv:"version_group_id"`
	MoveID         int `csv:"move_id"`
	MethodID       int `csv:"pokemon_move_method_id"`
}

type PokemonMove struct {
	PokemonID int
	MoveID    int
}

func NewPokemonMovesStrategy(db *sql.DB, dataDir string) PokemonMovesStrategy {
	return PokemonMovesStrategy{
		DB:                  db,
		pokemonMovesCSVPath: dataDir + "pokemon_moves.csv",
	}
}

func (s PokemonMovesStrategy) Run() error {
	rows, err := csv.ReadFile[pokemonMoveCSVRow](s.pokemonMovesCSVPath)
	if err != nil {
		return err
	}

	pokemonMap, err := queryPokeapiIDMap(s.DB, vgcalcPokemon)
	if err != nil {
		return err
	}
	movesMap, err := queryPokeapiIDMap(s.DB, vgcalcMoves)
	if err != nil {
		return err
	}

	pokemonMoves, err := buildPokemonMoves(rows, pokemonMap, movesMap)
	if err != nil {
		return err
	}

	// DO NOTHING (not UPDATE): the row has no columns besides its key, and it
	// also absorbs any duplicate pair left in the source.
	const query = `INSERT INTO vgcalc.pokemon_moves (pokemon_id, move_id) VALUES ($1, $2)
	ON CONFLICT (pokemon_id, move_id) DO NOTHING`

	return inTx(s.DB, func(tx *sql.Tx) error {
		for _, pokemonMove := range pokemonMoves {
			if _, err := tx.Exec(query, pokemonMove.PokemonID, pokemonMove.MoveID); err != nil {
				return fmt.Errorf("error inserting pokemon_move %v. %w", pokemonMove, err)
			}
		}

		return nil
	})
}

func buildPokemonMoves(rows []pokemonMoveCSVRow, pokemonMap map[int]int, movesMap map[int]int) ([]PokemonMove, error) {
	pokemonMoves := make([]PokemonMove, 0, len(rows))
	for _, row := range rows {
		if row.VersionGroupID != championsVersionGroupID || row.MethodID != trainMoveMethodID {
			continue
		}
		// Megas and primals live in mega_evolution, so their rows have no
		// vgcalc.pokemon match; they inherit the base pokemon's move list.
		pokemonID, ok := pokemonMap[row.PokemonID]
		if !ok {
			continue
		}
		moveID, ok := movesMap[row.MoveID]
		if !ok {
			return nil, fmt.Errorf("buildPokemonMoves(): unknown move_id=%d for pokemon_id=%d", row.MoveID, row.PokemonID)
		}
		pokemonMoves = append(pokemonMoves, PokemonMove{PokemonID: pokemonID, MoveID: moveID})
	}

	return pokemonMoves, nil
}
