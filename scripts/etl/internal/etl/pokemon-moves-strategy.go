package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/etl/common"
	"vgcalc/etl/internal/etl/entity"
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

func NewPokemonMovesStrategy(db *sql.DB, dataDir string) PokemonMovesStrategy {
	return PokemonMovesStrategy{
		DB:                  db,
		pokemonMovesCSVPath: dataDir + "pokemon_moves.csv",
	}
}

func (s PokemonMovesStrategy) Run() error {
	src, err := s.extract()
	if err != nil {
		return err
	}
	pokemonMoves, err := transformPokemonMoves(src)
	if err != nil {
		return err
	}

	return loadPokemonMoves(s.DB, pokemonMoves)
}

type pokemonMovesSource struct {
	pokemonMoveRows []pokemonMoveCSVRow
	pokemonMap      map[int]int
	movesMap        map[int]int
}

func (s PokemonMovesStrategy) extract() (pokemonMovesSource, error) {
	var src pokemonMovesSource
	var err error

	if src.pokemonMoveRows, err = csv.ReadFile[pokemonMoveCSVRow](s.pokemonMovesCSVPath); err != nil {
		return src, err
	}
	if src.pokemonMap, err = common.QueryPokeapiIDMap(s.DB, common.TablePokemon); err != nil {
		return src, err
	}
	if src.movesMap, err = common.QueryPokeapiIDMap(s.DB, common.TableMoves); err != nil {
		return src, err
	}

	return src, nil
}

func transformPokemonMoves(src pokemonMovesSource) ([]entity.PokemonMove, error) {
	pokemonMoves := make([]entity.PokemonMove, 0, len(src.pokemonMoveRows))
	for _, row := range src.pokemonMoveRows {
		if row.VersionGroupID != championsVersionGroupID || row.MethodID != trainMoveMethodID {
			continue
		}
		// Megas and primals live in mega_evolution, so their rows have no
		// vgcalc.pokemon match; they inherit the base pokemon's move list.
		pokemonID, ok := src.pokemonMap[row.PokemonID]
		if !ok {
			continue
		}
		moveID, ok := src.movesMap[row.MoveID]
		if !ok {
			return nil, fmt.Errorf("transformPokemonMoves(): unknown move_id=%d for pokemon_id=%d",
				row.MoveID, row.PokemonID)
		}
		pokemonMove := entity.PokemonMove{PokemonID: pokemonID, MoveID: moveID}
		if err := pokemonMove.Validate(); err != nil {
			return nil, fmt.Errorf("transformPokemonMoves(): %w", err)
		}
		pokemonMoves = append(pokemonMoves, pokemonMove)
	}

	return pokemonMoves, nil
}

func loadPokemonMoves(db *sql.DB, pokemonMoves []entity.PokemonMove) error {
	// DO NOTHING (not UPDATE): the row has no columns besides its key, and it
	// also absorbs any duplicate pair left in the source.
	const query = `INSERT INTO vgcalc.pokemon_moves (pokemon_id, move_id) VALUES ($1, $2)
	ON CONFLICT (pokemon_id, move_id) DO NOTHING`

	return common.InTx(db, func(tx *sql.Tx) error {
		for _, pokemonMove := range pokemonMoves {
			if _, err := tx.Exec(query, pokemonMove.PokemonID, pokemonMove.MoveID); err != nil {
				return fmt.Errorf("error inserting pokemon_move %v. %w", pokemonMove, err)
			}
		}

		return nil
	})
}
