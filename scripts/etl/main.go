package main

import (
	"database/sql"
	"fmt"
	"os"
	"vgcalc/etl/internal/env"
	"vgcalc/etl/internal/etl"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TODO: Organize this error handling
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

	typeStrategy, err := etl.NewTypeStrategy(db)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = typeStrategy.Run()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	typeEfficacyStrategy, err := etl.NewTypeEfficacyStrategy(db)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = typeEfficacyStrategy.Run()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
