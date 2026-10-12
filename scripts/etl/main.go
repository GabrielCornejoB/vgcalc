package main


import (
	"vgcalc/etl/internal/etlstrategy"
	"fmt"
)


func main() {
	typeStrategy := etlstrategy.NewTypeEtlStrategy()
	err := typeStrategy.Run()
	if err != nil {
		fmt.Printf(err.Error())
	}
}
