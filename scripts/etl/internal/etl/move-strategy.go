package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/entity"
)

type MoveStrategy struct {
	movesCSVPath           string
	moveNamesCSVPath       string
	moveEffectProseCSVPath string
	moveFlagMapCSVPath     string
	moveMetaCSVPath        string
	DB                     *sql.DB
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
	src, err := s.extract()
	if err != nil {
		return err
	}
	moves, err := transformMoves(src)
	if err != nil {
		return err
	}

	return loadMoves(s.DB, moves)
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

type moveSource struct {
	moveRows []moveCSVRow
	typesMap map[int]int
	names    map[int]string
	effects  map[int]moveEffectCSVRow
	flags    map[int]entity.MoveFlags
	hits     map[int]moveMetaCSVRow
}

func (s MoveStrategy) extract() (moveSource, error) {
	var src moveSource
	var err error

	if src.moveRows, err = csv.ReadFile[moveCSVRow](s.movesCSVPath); err != nil {
		return src, err
	}
	if src.typesMap, err = queryPokeapiIDMap(s.DB, vgcalcTypes); err != nil {
		return src, err
	}
	nameRows, err := csv.ReadFile[moveNameCSVRow](s.moveNamesCSVPath)
	if err != nil {
		return src, err
	}
	effectRows, err := csv.ReadFile[moveEffectCSVRow](s.moveEffectProseCSVPath)
	if err != nil {
		return src, err
	}
	flagRows, err := csv.ReadFile[moveFlagMapCSVRow](s.moveFlagMapCSVPath)
	if err != nil {
		return src, err
	}
	metaRows, err := csv.ReadFile[moveMetaCSVRow](s.moveMetaCSVPath)
	if err != nil {
		return src, err
	}

	src.names = make(map[int]string, len(nameRows))
	for _, row := range nameRows {
		if row.LanguageID == pokeapiLanguageID {
			src.names[row.MoveID] = row.Name
		}
	}
	src.effects = make(map[int]moveEffectCSVRow, len(effectRows))
	for _, row := range effectRows {
		if row.LanguageID == pokeapiLanguageID {
			src.effects[row.EffectID] = row
		}
	}
	src.flags = make(map[int]entity.MoveFlags, len(flagRows))
	for _, row := range flagRows {
		flags := src.flags[row.MoveID]
		switch row.FlagID {
		case 1:
			flags.IsContact = true
		case 8:
			flags.IsPunch = true
		case 16:
			flags.IsBite = true
		case 9:
			flags.IsSound = true
		case 17:
			flags.IsPulse = true
		case 18:
			flags.IsBullet = true
		}
		src.flags[row.MoveID] = flags
	}
	src.hits = make(map[int]moveMetaCSVRow, len(metaRows))
	for _, row := range metaRows {
		src.hits[row.MoveID] = row
	}

	return src, nil
}

func transformMoves(src moveSource) ([]entity.Move, error) {
	moves := make([]entity.Move, 0, len(src.moveRows))
	for _, row := range src.moveRows {
		category, ok := damageClassIDToSlug[row.DamageClassID]
		if !ok {
			return nil, fmt.Errorf("transformMoves(): unknown damage_class_id=%d", row.DamageClassID)
		}
		target, ok := moveTargetIDToSlug[row.TargetID]
		if !ok {
			return nil, fmt.Errorf("transformMoves(): unknown target_id=%d", row.TargetID)
		}
		typeID, ok := src.typesMap[row.TypeID]
		if !ok {
			return nil, fmt.Errorf("transformMoves(): unknown type_id=%d", row.TypeID)
		}

		effect := src.effects[row.EffectID]
		hits := src.hits[row.ID]
		move := entity.Move{
			PokeapiID:   row.ID,
			Slug:        row.Identifier,
			DisplayName: src.names[row.ID],
			ShortEffect: effect.ShortEffect,
			Effect:      effect.Effect,
			TypeID:      typeID,
			Category:    category,
			Power:       row.Power,
			Accuracy:    row.Accuracy,
			PP:          row.PP,
			Priority:    row.Priority,
			Target:      target,
			MoveFlags:   src.flags[row.ID],
			MinHits:     hits.MinHits,
			MaxHits:     hits.MaxHits,
		}
		if err := move.Validate(); err != nil {
			return nil, fmt.Errorf("transformMoves(): %w", err)
		}
		moves = append(moves, move)
	}

	return moves, nil
}

func loadMoves(db *sql.DB, moves []entity.Move) error {
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
