package etl

type EtlStrategy interface {
	extract() error
	transform() error
	load() error
	Run() error
}

