package etl

import (
	"database/sql"
	"fmt"
)

// vgcalcTable restricts loadPokeapiIDMap to known vgcalc tables.
// SQL identifiers cannot be query parameters, so the table name is concatenated;
// the closed set of constants below is the only way to construct one.
type vgcalcTable string

const (
	vgcalcTypes     vgcalcTable = "types"
	vgcalcAbilities vgcalcTable = "abilities"
	vgcalcItems     vgcalcTable = "items"
	vgcalcMoves     vgcalcTable = "moves"
	vgcalcPokemon   vgcalcTable = "pokemon"
)

// loadPokeapiIDMap maps every source pokeapi_id in a table to its surrogate id.
func loadPokeapiIDMap(db *sql.DB, table vgcalcTable) (map[int]int, error) {
	rows, err := db.Query("SELECT id, pokeapi_id FROM vgcalc." + string(table))
	if err != nil {
		return nil, fmt.Errorf("loadPokeapiIDMap(%s): failed to query existing rows. %w", table, err)
	}
	defer rows.Close()

	idMap := make(map[int]int)
	for rows.Next() {
		var dbID, pokeapiID int
		if err := rows.Scan(&dbID, &pokeapiID); err != nil {
			return nil, fmt.Errorf("loadPokeapiIDMap(%s): %w", table, err)
		}
		idMap[pokeapiID] = dbID
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("loadPokeapiIDMap(%s): %w", table, err)
	}

	return idMap, nil
}
