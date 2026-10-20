package etl

// pokeapiLanguageID is PokeAPI's EN language (language_id 9 everywhere).
const pokeapiLanguageID = 9

// PokeAPI stat ids (stats.csv). Ids 7 and 8 (accuracy, evasion) are not battle stats.
const (
	statHP  = 1
	statAtk = 2
	statDef = 3
	statSpA = 4
	statSpD = 5
	statSpe = 6
)
