package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
)

type MoveStrategy struct {
	movesCSVPath           string
	moveNamesCSVPath       string
	moveEffectProseCSVPath string
	moveFlagMapCSVPath     string
	moveMetaCSVPath        string
	DB                     *sql.DB
}

type Move struct {
	PokeapiID   int
	Slug        string
	DisplayName string
	EffectID    int
	ShortEffect string
	Effect      string
	TypeID      int
	Category    string
	Power       int
	Accuracy    int
	PP          int
	Priority    int
	Target      string
	IsContact   bool
	IsPunch     bool
	IsBite      bool
	IsSound     bool
	IsPulse     bool
	IsBullet    bool
	IsSlicing   bool
	IsWind      bool
	MinHits     *int
	MaxHits     *int
}

func NewMoveStrategy(db *sql.DB, dataDir string) MoveStrategy {
	return MoveStrategy{
		DB:                     db,
		movesCSVPath:           dataDir + "moves.csv",
		moveNamesCSVPath:       dataDir + "move_names.csv",
		moveEffectProseCSVPath: dataDir + "move_effect_prose.csv",
		moveFlagMapCSVPath:     dataDir + "move_flag_map.csv",
		moveMetaCSVPath:        dataDir + "move_meta.csv",
	}
}

var damageClassIDToSlug = map[int]string{
	1: "status",
	2: "physical",
	3: "special",
}

var moveTargetIDToSlug = map[int]string{
	1:  "specific-move",
	2:  "selected-pokemon-me-first",
	3:  "ally",
	4:  "users-field",
	5:  "user-or-ally",
	6:  "opponents-field",
	7:  "user",
	8:  "random-opponent",
	9:  "all-other-pokemon",
	10: "selected-pokemon",
	11: "all-opponents",
	12: "entire-field",
	13: "user-and-allies",
	14: "all-pokemon",
	15: "all-allies",
	16: "fainting-pokemon",
}

func (s MoveStrategy) Run() error {
	rows, err := csv.ReadFile[moveCSVRow](s.movesCSVPath)
	if err != nil {
		return err
	}

	typesMap, err := queryPokeapiIDMap(s.DB, vgcalcTypes)
	if err != nil {
		return err
	}

	moves, movesMap, movesPerEffect, err := buildMoves(rows, typesMap)
	if err != nil {
		return err
	}

	nameRows, err := csv.ReadFile[moveNameCSVRow](s.moveNamesCSVPath)
	if err != nil {
		return err
	}

	applyMoveNames(nameRows, moves, movesMap)

	effectRows, err := csv.ReadFile[moveEffectCSVRow](s.moveEffectProseCSVPath)
	if err != nil {
		return err
	}

	applyMoveEffects(effectRows, moves, movesPerEffect)

	flagRows, err := csv.ReadFile[moveFlagMapCSVRow](s.moveFlagMapCSVPath)
	if err != nil {
		return err
	}

	applyMoveFlags(flagRows, moves, movesMap)

	metaRows, err := csv.ReadFile[moveMetaCSVRow](s.moveMetaCSVPath)
	if err != nil {
		return err
	}

	applyMoveMeta(metaRows, moves, movesMap)

	return upsertMoves(s.DB, moves)
}

func upsertMoves(db *sql.DB, moves []Move) error {
	const query = `INSERT INTO vgcalc.moves (
		pokeapi_id, slug, display_name, short_effect, effect,
		type_id, category, power, accuracy, pp, priority, target,
		is_contact, is_punch, is_bite, is_sound, is_pulse, is_bullet, is_slicing, is_wind,
		min_hits, max_hits
	)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
		$13, $14, $15, $16, $17, $18, $19, $20, $21, $22
	)
	ON CONFLICT (pokeapi_id) DO UPDATE SET
		slug = EXCLUDED.slug,
		display_name = EXCLUDED.display_name,
		short_effect = EXCLUDED.short_effect,
		effect = EXCLUDED.effect,
		type_id = EXCLUDED.type_id,
		category = EXCLUDED.category,
		power = EXCLUDED.power,
		accuracy = EXCLUDED.accuracy,
		pp = EXCLUDED.pp,
		priority = EXCLUDED.priority,
		target = EXCLUDED.target,
		is_contact = EXCLUDED.is_contact,
		is_punch = EXCLUDED.is_punch,
		is_bite = EXCLUDED.is_bite,
		is_sound = EXCLUDED.is_sound,
		is_pulse = EXCLUDED.is_pulse,
		is_bullet = EXCLUDED.is_bullet,
		is_slicing = EXCLUDED.is_slicing,
		is_wind = EXCLUDED.is_wind,
		min_hits = EXCLUDED.min_hits,
		max_hits = EXCLUDED.max_hits`

	return inTx(db, func(tx *sql.Tx) error {
		for _, move := range moves {
			_, err := tx.Exec(query,
				move.PokeapiID, move.Slug, move.DisplayName, move.ShortEffect, move.Effect,
				move.TypeID, move.Category, move.Power, move.Accuracy, move.PP, move.Priority,
				move.Target, move.IsContact, move.IsPunch, move.IsBite, move.IsSound,
				move.IsPulse, move.IsBullet, move.IsSlicing, move.IsWind, move.MinHits, move.MaxHits)
			if err != nil {
				return fmt.Errorf("error inserting move %v. %w", move, err)
			}
		}

		return nil
	})
}

type moveCSVRow struct {
	ID            int    `csv:"id"`
	Identifier    string `csv:"identifier"`
	EffectID      int    `csv:"effect_id,omitempty"`
	TypeID        int    `csv:"type_id"`
	Power         int    `csv:"power,omitempty"`
	PP            int    `csv:"pp,omitempty"`
	Accuracy      int    `csv:"accuracy,omitempty"`
	Priority      int    `csv:"priority"`
	TargetID      int    `csv:"target_id"`
	DamageClassID int    `csv:"damage_class_id"`
}

type moveNameCSVRow struct {
	MoveID     int    `csv:"move_id"`
	LanguageID int    `csv:"local_language_id"`
	Name       string `csv:"name"`
}

type moveEffectCSVRow struct {
	EffectID    int    `csv:"move_effect_id"`
	LanguageID  int    `csv:"local_language_id"`
	ShortEffect string `csv:"short_effect"`
	Effect      string `csv:"effect"`
}

type moveFlagMapCSVRow struct {
	MoveID int `csv:"move_id"`
	FlagID int `csv:"move_flag_id"`
}

type moveMetaCSVRow struct {
	MoveID  int  `csv:"move_id"`
	MinHits *int `csv:"min_hits,omitempty"`
	MaxHits *int `csv:"max_hits,omitempty"`
}

func buildMoves(rows []moveCSVRow, typesMap map[int]int) ([]Move, map[int]int, map[int][]int, error) {
	moves := make([]Move, 0, len(rows))
	movesMap := make(map[int]int, len(rows))
	movesPerEffect := make(map[int][]int)
	for _, row := range rows {
		category, ok := damageClassIDToSlug[row.DamageClassID]
		if !ok {
			return nil, nil, nil, fmt.Errorf("buildMoves(): unknown damage_class_id=%d", row.DamageClassID)
		}
		target, ok := moveTargetIDToSlug[row.TargetID]
		if !ok {
			return nil, nil, nil, fmt.Errorf("buildMoves(): unknown target_id=%d", row.TargetID)
		}
		typeID, ok := typesMap[row.TypeID]
		if !ok {
			return nil, nil, nil, fmt.Errorf("buildMoves(): unknown type_id=%d", row.TypeID)
		}
		moves = append(moves, Move{
			PokeapiID: row.ID,
			Slug:      row.Identifier,
			EffectID:  row.EffectID,
			TypeID:    typeID,
			Power:     row.Power,
			PP:        row.PP,
			Accuracy:  row.Accuracy,
			Priority:  row.Priority,
			Target:    target,
			Category:  category,
		})
		movesMap[row.ID] = len(moves) - 1
		if row.EffectID != 0 {
			movesPerEffect[row.EffectID] = append(movesPerEffect[row.EffectID], len(moves)-1)
		}
	}

	return moves, movesMap, movesPerEffect, nil
}

func applyMoveNames(rows []moveNameCSVRow, moves []Move, movesMap map[int]int) {
	for _, row := range rows {
		if row.LanguageID != pokeapiLanguageID {
			continue
		}
		moveIndex, ok := movesMap[row.MoveID]
		if !ok {
			continue
		}
		moves[moveIndex].DisplayName = row.Name
	}
}

func applyMoveEffects(rows []moveEffectCSVRow, moves []Move, movesPerEffect map[int][]int) {
	for _, row := range rows {
		if row.LanguageID != pokeapiLanguageID {
			continue
		}
		for _, moveIndex := range movesPerEffect[row.EffectID] {
			moves[moveIndex].Effect = row.Effect
			moves[moveIndex].ShortEffect = row.ShortEffect
		}
	}
}

func applyMoveFlags(rows []moveFlagMapCSVRow, moves []Move, movesMap map[int]int) {
	for _, row := range rows {
		moveIndex, ok := movesMap[row.MoveID]
		if !ok {
			continue
		}
		switch row.FlagID {
		case 1:
			moves[moveIndex].IsContact = true
		case 8:
			moves[moveIndex].IsPunch = true
		case 16:
			moves[moveIndex].IsBite = true
		case 9:
			moves[moveIndex].IsSound = true
		case 17:
			moves[moveIndex].IsPulse = true
		case 18:
			moves[moveIndex].IsBullet = true
		}
	}
}

func applyMoveMeta(rows []moveMetaCSVRow, moves []Move, movesMap map[int]int) {
	for _, row := range rows {
		moveIndex, ok := movesMap[row.MoveID]
		if !ok {
			continue
		}
		moves[moveIndex].MinHits = row.MinHits
		moves[moveIndex].MaxHits = row.MaxHits
	}
}
