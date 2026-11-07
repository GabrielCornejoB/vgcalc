package etl

type EtlStrategy interface {
	Run() error
}
