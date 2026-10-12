package etlstrategy

type EtlStrategy interface {
	extract() error
	transform() error
	load() error
	Run() error
}

