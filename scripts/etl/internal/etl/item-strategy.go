package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
	"vgcalc/etl/internal/entity"
)

type ItemStrategy struct {
	itemsCSVPath     string
	itemNamesCSVPath string
	itemProseCSVPath string
	DB               *sql.DB
}

func NewItemStrategy(db *sql.DB, dataDir string) ItemStrategy {
	return ItemStrategy{
		DB:               db,
		itemsCSVPath:     dataDir + "items.csv",
		itemNamesCSVPath: dataDir + "item_names.csv",
		itemProseCSVPath: dataDir + "item_prose.csv",
	}
}

func (s ItemStrategy) Run() error {
	src, err := s.extract()
	if err != nil {
		return err
	}
	items, err := transformItems(src)
	if err != nil {
		return err
	}

	return loadItems(s.DB, items)
}

type itemCSVRow struct {
	ID         int    `csv:"id"`
	Identifier string `csv:"identifier"`
	FlingPower int    `csv:"fling_power,omitempty"`
}

type itemNameCSVRow struct {
	ItemID          int    `csv:"item_id"`
	LocalLanguageID int    `csv:"local_language_id"`
	Name            string `csv:"name"`
}

type itemProseCSVRow struct {
	ItemID          int    `csv:"item_id"`
	LocalLanguageID int    `csv:"local_language_id"`
	ShortEffect     string `csv:"short_effect"`
	Effect          string `csv:"effect"`
}

type itemSource struct {
	itemRows []itemCSVRow
	names    map[int]string
	prose    map[int]itemProseCSVRow
}

func (s ItemStrategy) extract() (itemSource, error) {
	var src itemSource
	var err error

	if src.itemRows, err = csv.ReadFile[itemCSVRow](s.itemsCSVPath); err != nil {
		return src, err
	}
	nameRows, err := csv.ReadFile[itemNameCSVRow](s.itemNamesCSVPath)
	if err != nil {
		return src, err
	}
	proseRows, err := csv.ReadFile[itemProseCSVRow](s.itemProseCSVPath)
	if err != nil {
		return src, err
	}

	src.names = make(map[int]string, len(nameRows))
	for _, row := range nameRows {
		if row.LocalLanguageID == pokeapiLanguageID {
			src.names[row.ItemID] = row.Name
		}
	}
	src.prose = make(map[int]itemProseCSVRow, len(proseRows))
	for _, row := range proseRows {
		if row.LocalLanguageID == pokeapiLanguageID {
			src.prose[row.ItemID] = row
		}
	}

	return src, nil
}

func transformItems(src itemSource) ([]entity.Item, error) {
	items := make([]entity.Item, 0, len(src.itemRows))
	seenSlugs := make(map[string]struct{}, len(src.itemRows))
	for _, row := range src.itemRows {
		// PokeAPI contains duplicate identifiers (e.g. roseli-berry as both
		// 723 and 2279 — same data in our columns, newer row has no EN text).
		// Keep the first occurrence (lowest pokeapi_id), skip the rest.
		// Generic rule, no per-row exceptions.
		if _, seen := seenSlugs[row.Identifier]; seen {
			continue
		}
		seenSlugs[row.Identifier] = struct{}{}

		prose := src.prose[row.ID]
		item := entity.Item{
			PokeapiID:   row.ID,
			Slug:        row.Identifier,
			DisplayName: src.names[row.ID],
			ShortEffect: prose.ShortEffect,
			Effect:      prose.Effect,
			FlingPower:  row.FlingPower,
		}
		if err := item.Validate(); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func loadItems(db *sql.DB, items []entity.Item) error {
	const query = `INSERT INTO vgcalc.items (pokeapi_id, slug, display_name, short_effect, effect, fling_power)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (pokeapi_id) DO UPDATE SET
		slug = EXCLUDED.slug,
		display_name = EXCLUDED.display_name,
		short_effect = EXCLUDED.short_effect,
		effect = EXCLUDED.effect,
		fling_power = EXCLUDED.fling_power`

	return inTx(db, func(tx *sql.Tx) error {
		for _, item := range items {
			_, err := tx.Exec(query,
				item.PokeapiID, item.Slug, item.DisplayName,
				item.ShortEffect, item.Effect, item.FlingPower)
			if err != nil {
				return fmt.Errorf("error inserting item %v. %w", item, err)
			}
		}

		return nil
	})
}
