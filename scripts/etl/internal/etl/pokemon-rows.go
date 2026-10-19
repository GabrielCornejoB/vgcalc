package etl

// Row structs for the pokemon*.csv family, shared by every strategy that
// consumes them (pokemon, mega_evolution). Field sets are the union of what
// each strategy reads; csvutil ignores columns a strategy does not use.

type PokemonCSVRow struct {
	ID         int    `csv:"id"`
	Identifier string `csv:"identifier"`
	SpeciesID  int    `csv:"species_id"`
	Weight     int    `csv:"weight"`
	IsDefault  int    `csv:"is_default"`
}

type PokemonStatCSVRow struct {
	PokemonID int `csv:"pokemon_id"`
	StatID    int `csv:"stat_id"`
	BaseStat  int `csv:"base_stat"`
}

type PokemonTypeCSVRow struct {
	PokemonID int `csv:"pokemon_id"`
	TypeID    int `csv:"type_id"`
	Slot      int `csv:"slot"`
}

type PokemonAbilityCSVRow struct {
	PokemonID int `csv:"pokemon_id"`
	AbilityID int `csv:"ability_id"`
	IsHidden  int `csv:"is_hidden"`
	Slot      int `csv:"slot"`
}

type PokemonFormCSVRow struct {
	ID             int    `csv:"id"`
	PokemonID      int    `csv:"pokemon_id"`
	IsMega         int    `csv:"is_mega"`
	FormIdentifier string `csv:"form_identifier"`
	IsBattleOnly   int    `csv:"is_battle_only"`
}

type PokemonSpeciesCSVRow struct {
	ID         int `csv:"id"`
	GenderRate int `csv:"gender_rate"`
}
