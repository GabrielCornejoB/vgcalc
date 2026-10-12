package etlstrategy

import (
	"context"
	"fmt"
	"slices"
	"vgcalc/etl/internal/utils"

	"github.com/jackc/pgx/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TypeEtlStrategy struct {
	filePathUrl string
	dbConnection *pgx.Conn
}

func NewTypeEtlStrategy(dbConnection *pgx.Conn) (TypeEtlStrategy, error) {
	filePath, err := utils.GetEnvByKey("PATH_TYPES_CSV")
	if err != nil {
		return TypeEtlStrategy{}, fmt.Errorf("NewTypeEtlStrategy(): %w", err)
	}
	return TypeEtlStrategy{filePath, dbConnection}, nil
}

type PokemonType struct {
	slug string
	displayName string
}

func (strategy TypeEtlStrategy) extract() ([][]string, error) {
	records, err := utils.ReadCsvFile(strategy.filePathUrl)
	if err != nil {
		return nil, fmt.Errorf("TypeEtlStrategy.extract(): %w", err)
	}

	return records, nil
}

func (strategy TypeEtlStrategy) transform(records [][]string) ([]PokemonType, error) {
	records = records[1:]
	pokemonTypes := make([]PokemonType, 0)
	caser := cases.Title(language.English)

	for _, record := range records {
		pokemonTypes = append(pokemonTypes, PokemonType{slug:record[1], displayName: caser.String(record[1])})
	}

	return pokemonTypes, nil
}

func (strategy TypeEtlStrategy) load(pokemonTypes []PokemonType) error {
	const sql = "INSERT INTO vgcalc.types (slug, display_name) VALUES ($1, $2)"
	var ctx = context.Background()
	var blackListedTypes = []string{"unknown", "shadow"}

	for _, pokemonType := range pokemonTypes {
		if slices.Contains(blackListedTypes, pokemonType.slug) {
			continue
		}
		_, err := strategy.dbConnection.Exec(ctx, sql, pokemonType.slug, pokemonType.displayName)
		if err != nil {
			return fmt.Errorf("TypeEtlStrategy.load(): pokemonType=%v. %w", pokemonType, err)
		}
	}

	return nil
}

func (strategy TypeEtlStrategy) Run() error {
	records, err := strategy.extract()
	if err != nil {
		return fmt.Errorf("TypeEtlStrategy.Run(): %w", err)
	}

	pokemonTypes, err := strategy.transform(records)
	if err != nil {
		return fmt.Errorf("TypeEtlStrategy.Run(): %w", err)
	}

	err = strategy.load(pokemonTypes)
	if err != nil {
		return fmt.Errorf("TypeEtlStrategy.Run(): %w", err)
	}

	return nil
}

