package main

import (
	"database/sql"
	"fmt"
	"os"
	"vgcalc/etl/internal/env"
	"vgcalc/etl/internal/etl"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	err := env.LoadEnvs()
	if err != nil {
		fmt.Println(err)
		return
	}

	connectionString, err := env.GetDbConnectionString()
	if err != nil {
		fmt.Println(err)
		return
	}

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	dataDir, err := env.GetEnvByKey("DATA_DIR")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	strategies := []etl.EtlStrategy{
		etl.NewTypeStrategy(db, dataDir),
		etl.NewTypeEfficacyStrategy(db, dataDir),
		etl.NewAbilityStrategy(db, dataDir),
		etl.NewNatureStrategy(db, dataDir),
		etl.NewItemStrategy(db, dataDir),
		etl.NewMoveStrategy(db, dataDir),
		etl.NewPokemonStrategy(db, dataDir),
		etl.NewMegaEvolutionStrategy(db, dataDir),
		etl.NewPokemonMovesStrategy(db, dataDir),
	}

	for _, strategy := range strategies {
		if err := strategy.Run(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

}
