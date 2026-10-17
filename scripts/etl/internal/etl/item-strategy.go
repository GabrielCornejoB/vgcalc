package etl

import (
	"database/sql"
	"fmt"

	"vgcalc/etl/internal/csv"
)

type ItemStrategy struct {
	itemsCSVPath     string
	itemNamesCSVPath string
	itemProseCSVPath string
	DB               *sql.DB
}

type Item struct {
	PokeapiID   int
	Slug        string
	DisplayName string
	ShortEffect string
	Effect      string
	FlingPower  int
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
	itemRows, err := csv.LoadFile[itemCSVRow](s.itemsCSVPath)
	if err != nil {
		return fmt.Errorf("ItemStrategy: %w", err)
	}
	items, itemsMap := buildItems(itemRows)

	nameRows, err := csv.LoadFile[itemNameCSVRow](s.itemNamesCSVPath)
	if err != nil {
		return fmt.Errorf("ItemStrategy: %w", err)
	}
	applyItemNames(nameRows, items, itemsMap)

	proseRows, err := csv.LoadFile[itemProseCSVRow](s.itemProseCSVPath)
	if err != nil {
		return fmt.Errorf("ItemStrategy: %w", err)
	}
	applyItemEffects(proseRows, items, itemsMap)

	const query = `INSERT INTO vgcalc.items (pokeapi_id, slug, display_name, short_effect, effect, fling_power) VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (pokeapi_id) DO UPDATE SET slug = EXCLUDED.slug, display_name = EXCLUDED.display_name, short_effect = EXCLUDED.short_effect, effect = EXCLUDED.effect, fling_power = EXCLUDED.fling_power`

	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("ItemStrategy: error starting transaction. %w", err)
	}
	defer tx.Rollback()

	for _, item := range items {
		_, err := tx.Exec(query, item.PokeapiID, item.Slug, item.DisplayName, item.ShortEffect, item.Effect, item.FlingPower)
		if err != nil {
			return fmt.Errorf("ItemStrategy: error inserting item %v. %w", item, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ItemStrategy: error committing transaction. %w", err)
	}

	return nil
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

func buildItems(rows []itemCSVRow) ([]Item, map[int]int) {
	items := make([]Item, 0, len(rows))
	itemsMap := make(map[int]int, len(rows))
	seenSlugs := make(map[string]struct{})
	for _, row := range rows {
		// PokeAPI contains duplicate identifiers (e.g. roseli-berry as both
		// 723 and 2279 — same data in our columns, newer row has no EN text).
		// Keep the first occurrence (lowest pokeapi_id), skip the rest.
		// Generic rule, no per-row exceptions.
		if _, seen := seenSlugs[row.Identifier]; seen {
			continue
		}
		items = append(items, Item{
			PokeapiID:  row.ID,
			Slug:       row.Identifier,
			FlingPower: row.FlingPower,
		})
		itemsMap[row.ID] = len(items) - 1
		seenSlugs[row.Identifier] = struct{}{}
	}

	return items, itemsMap
}

func applyItemNames(rows []itemNameCSVRow, items []Item, itemsMap map[int]int) {
	for _, row := range rows {
		// If language id is != 9 (english), then skip the record
		if row.LocalLanguageID != pokeapiLanguageID {
			continue
		}
		itemIndex, exists := itemsMap[row.ItemID]
		if !exists {
			continue
		}
		items[itemIndex].DisplayName = row.Name
	}
}

func applyItemEffects(rows []itemProseCSVRow, items []Item, itemsMap map[int]int) {
	for _, row := range rows {
		// If language id is != 9 (english), then skip the record
		if row.LocalLanguageID != pokeapiLanguageID {
			continue
		}
		itemIndex, exists := itemsMap[row.ItemID]
		if !exists {
			continue
		}
		items[itemIndex].ShortEffect = row.ShortEffect
		items[itemIndex].Effect = row.Effect
	}
}
