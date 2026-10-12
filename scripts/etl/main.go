package main

import (
	"context"
	"fmt"
	"os"
	"vgcalc/etl/internal/etlstrategy"
	"vgcalc/etl/internal/utils"

	"github.com/jackc/pgx/v5"
)


// TODO: Organize this error handling
func main() {
	connectionString, err := utils.GetDbConnectionString()
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

	typeStrategy, err := etlstrategy.NewTypeEtlStrategy(conn)

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
