package etl

import (
	"database/sql"
	"fmt"
	"strconv"
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
	records, err := csv.ReadCSVFile(s.itemsCSVPath)
	if err != nil {
		return fmt.Errorf("ItemStrategy: error reading csv file. %w", err)
	}
	items, itemsMap, err := loadItems(records)
	if err != nil {
		return fmt.Errorf("ItemStrategy: %w", err)
	}

	records, err = csv.ReadCSVFile(s.itemNamesCSVPath)
	if err != nil {
		return fmt.Errorf("ItemStrategy: error reading csv file. %w", err)
	}
	err = loadItemNames(records, items, itemsMap)
	if err != nil {
		return fmt.Errorf("ItemStrategy: %w", err)
	}

	records, err = csv.ReadCSVFile(s.itemProseCSVPath)
	if err != nil {
		return fmt.Errorf("ItemStrategy: error reading csv file. %w", err)
	}
	err = loadItemEffects(records, items, itemsMap)
	if err != nil {
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

func loadItems(records [][]string) ([]Item, map[int]int, error) {
	items := make([]Item, 0, len(records))
	itemsMap := make(map[int]int)
	seenSlugs := make(map[string]struct{})
	for _, record := range records {
		if len(record) < 5 {
			return nil, nil, fmt.Errorf("loadItems(): invalid record=%v", record)
		}
		// PokeAPI contains duplicate identifiers (e.g. roseli-berry as both
		// 723 and 2279 — same data in our columns, newer row has no EN text).
		// Keep the first occurrence (lowest pokeapi_id), skip the rest.
		// Generic rule, no per-row exceptions.
		if _, seen := seenSlugs[record[1]]; seen {
			continue
		}
		pokeapiID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, nil, fmt.Errorf("loadItems(): invalid pokeapi_id=%q. %w", record[0], err)
		}
		// Empty fling_power means unflingable -> NULL.
		var flingPower sql.NullInt64
		if record[4] != "" {
			flingPowerValue, err := strconv.Atoi(record[4])
			if err != nil {
				return nil, nil, fmt.Errorf("loadItems(): invalid fling_power=%q. %w", record[4], err)
			}
			flingPower = sql.NullInt64{Int64: int64(flingPowerValue), Valid: true}
		}
		items = append(items, Item{PokeapiID: pokeapiID, Slug: record[1], FlingPower: flingPower})
		itemsMap[pokeapiID] = len(items) - 1
		seenSlugs[record[1]] = struct{}{}
	}

	return items, itemsMap, nil
}

func loadItemNames(records [][]string, items []Item, itemsMap map[int]int) error {
	for _, record := range records {
		if len(record) < 3 {
			return fmt.Errorf("loadItemNames(): invalid record=%v", record)
		}
		pokeapiID, err := strconv.Atoi(record[0])
		if err != nil {
			return fmt.Errorf("loadItemNames(): invalid pokeapi_id=%q. %w", record[0], err)
		}
		localLanguageID, err := strconv.Atoi(record[1])
		if err != nil {
			return fmt.Errorf("loadItemNames(): invalid local_language_id=%q. %w", record[1], err)
		}
		// If language id is != 9 (english), then skip the record
		if localLanguageID != pokeapiLanguageID {
			continue
		}
		itemIndex, exists := itemsMap[pokeapiID]
		if !exists {
			continue
		}
		items[itemIndex].DisplayName = record[2]
	}
	return nil
}

func loadItemEffects(records [][]string, items []Item, itemsMap map[int]int) error {
	for _, record := range records {
		if len(record) < 4 {
			return fmt.Errorf("loadItemEffects(): invalid record=%v", record)
		}
		pokeapiID, err := strconv.Atoi(record[0])
		if err != nil {
			return fmt.Errorf("loadItemEffects(): invalid pokeapi_id=%q. %w", record[0], err)
		}
		localLanguageID, err := strconv.Atoi(record[1])
		if err != nil {
			return fmt.Errorf("loadItemEffects(): invalid local_language_id=%q. %w", record[1], err)
		}
		// If language id is != 9 (english), then skip the record
		if localLanguageID != pokeapiLanguageID {
			continue
		}
		itemIndex, exists := itemsMap[pokeapiID]
		if !exists {
			continue
		}
		items[itemIndex].ShortEffect = record[2]
		items[itemIndex].Effect = record[3]
	}
	return nil
}
