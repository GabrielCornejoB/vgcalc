package main

import (
	"context"
	"fmt"
	"os"
	"vgcalc/etl/internal/etl"
	"vgcalc/etl/internal/env"

	"github.com/jackc/pgx/v5"
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

	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	typeStrategy, err := etl.NewTypeStrategy(conn)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = typeStrategy.Run()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
