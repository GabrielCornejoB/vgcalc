package common

import (
	"database/sql"
	"fmt"
)

// Table restricts QueryPokeapiIDMap to known vgcalc tables.
// SQL identifiers cannot be query parameters, so the table name is concatenated;
// the closed set of constants below is the only way to construct one.
type Table string

const (
	TableTypes     Table = "types"
	TableAbilities Table = "abilities"
	TableItems     Table = "items"
	TableMoves     Table = "moves"
	TablePokemon   Table = "pokemon"
)

// QueryPokeapiIDMap maps every source pokeapi_id in a table to its surrogate id.
func QueryPokeapiIDMap(db *sql.DB, table Table) (map[int]int, error) {
	rows, err := db.Query("SELECT id, pokeapi_id FROM vgcalc." + string(table))
	if err != nil {
		return nil, fmt.Errorf("QueryPokeapiIDMap(%s): failed to query existing rows. %w", table, err)
	}
	defer rows.Close()

	idMap := make(map[int]int)
	for rows.Next() {
		var dbID, pokeapiID int
		if err := rows.Scan(&dbID, &pokeapiID); err != nil {
			return nil, fmt.Errorf("QueryPokeapiIDMap(%s): %w", table, err)
		}
		idMap[pokeapiID] = dbID
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("QueryPokeapiIDMap(%s): %w", table, err)
	}

	return idMap, nil
}
