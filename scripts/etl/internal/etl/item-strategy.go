package etl

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/jszwec/csvutil"
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
	FlingPower  sql.NullInt64
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
	itemsData, err := os.ReadFile(s.itemsCSVPath)
	if err != nil {
		return fmt.Errorf("ItemStrategy: error reading csv file. %w", err)
	}
	items, itemsMap, err := loadItems(itemsData)
	if err != nil {
		return fmt.Errorf("ItemStrategy: %w", err)
	}

	namesData, err := os.ReadFile(s.itemNamesCSVPath)
	if err != nil {
		return fmt.Errorf("ItemStrategy: error reading csv file. %w", err)
	}
	if err := loadItemNames(namesData, items, itemsMap); err != nil {
		return fmt.Errorf("ItemStrategy: %w", err)
	}

	proseData, err := os.ReadFile(s.itemProseCSVPath)
	if err != nil {
		return fmt.Errorf("ItemStrategy: error reading csv file. %w", err)
	}
	if err := loadItemEffects(proseData, items, itemsMap); err != nil {
		return fmt.Errorf("ItemStrategy: %w", err)
	}

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

// NullInt64 maps an empty CSV field to NULL (Valid=false).
// csvutil calls it per field via the Unmarshaler interface, so the
// "empty means unflingable" rule lives in one place.
type NullInt64 struct {
	sql.NullInt64
}

func (n *NullInt64) UnmarshalCSV(data []byte) error {
	if len(data) == 0 {
		n.Int64, n.Valid = 0, false
		return nil
	}
	v, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("invalid fling_power=%q. %w", string(data), err)
	}
	n.Int64, n.Valid = int64(v), true
	return nil
}

// Raw CSV rows. Field mapping is declarative via `csv` tags (header names),
// so column indexes and per-field Atoi boilerplate disappear.
type itemCSVRow struct {
	ID         int       `csv:"id"`
	Identifier string    `csv:"identifier"`
	FlingPower NullInt64 `csv:"fling_power"`
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

func loadItems(data []byte) ([]Item, map[int]int, error) {
	var rows []itemCSVRow
	if err := csvutil.Unmarshal(data, &rows); err != nil {
		return nil, nil, fmt.Errorf("loadItems(): %w", err)
	}

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
			FlingPower: row.FlingPower.NullInt64,
		})
		itemsMap[row.ID] = len(items) - 1
		seenSlugs[row.Identifier] = struct{}{}
	}

	return items, itemsMap, nil
}

func loadItemNames(data []byte, items []Item, itemsMap map[int]int) error {
	var rows []itemNameCSVRow
	if err := csvutil.Unmarshal(data, &rows); err != nil {
		return fmt.Errorf("loadItemNames(): %w", err)
	}
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
	return nil
}

func loadItemEffects(data []byte, items []Item, itemsMap map[int]int) error {
	var rows []itemProseCSVRow
	if err := csvutil.Unmarshal(data, &rows); err != nil {
		return fmt.Errorf("loadItemEffects(): %w", err)
	}
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
	return nil
}
