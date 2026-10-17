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

	typesCSVPath, err := env.GetEnvByKey("PATH_TYPES_CSV")
	if err != nil {
		fmt.Println(os.Stderr, err.Error())
		os.Exit(1)
	}

	typeStrategy := etl.TypeStrategy{DB: db, CSVPath: typesCSVPath}
	if err := typeStrategy.Run(); err != nil {
		fmt.Println(os.Stderr, err.Error())
		os.Exit(1)
	}

}
