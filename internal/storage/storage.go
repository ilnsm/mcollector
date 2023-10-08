package storage

type Storager interface {
	InsertGauge(k string, v float64) error
	InsertCounter(k string, v int64) error
	//Update(k,v string) error
	//Delete(k,v string) error
	SelectGauge(k string) (float64, error)
	SelectCounter(k string) (int64, error)
}
