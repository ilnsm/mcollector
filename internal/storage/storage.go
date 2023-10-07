package storage

type Storager interface {
	InsertGauge(k string, v float64) error
	InsertCounter(k string, v int64) error
	//Update(k,v string) error
	//Delete(k,v string) error
	//Select(k,v string) error
}
