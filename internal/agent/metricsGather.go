package agent

import (
	"runtime"
	"strconv"
)

func GetMetrics(ms *runtime.MemStats) (map[string]string, error) {
	runtime.ReadMemStats(ms)
	mtr := make(map[string]string)
	mtr["Alloc"] = string(ms.Alloc)
	mtr["BuckHashSys"] = string(ms.BuckHashSys)
	mtr["Frees"] = string(ms.Frees)
	mtr["GCCPUFraction"] = strconv.FormatFloat(ms.GCCPUFraction, 'f', -1, 64)
	mtr["GCSys"] = string(ms.GCSys)
	mtr["HeapAlloc"] = string(ms.HeapAlloc)
	mtr["HeapIdle"] = string(ms.HeapIdle)
	mtr["HeapInuse"] = string(ms.HeapInuse)
	mtr["HeapObjects"] = string(ms.HeapObjects)
	mtr["HeapReleased"] = string(ms.HeapReleased)
	mtr["HeapSys"] = string(ms.HeapSys)
	mtr["LastGC"] = string(ms.LastGC)
	mtr["Lookups"] = string(ms.Lookups)
	mtr["MCacheInuse"] = string(ms.MCacheInuse)
	mtr["MCacheSys"] = string(ms.MCacheSys)
	mtr["MSpanInuse"] = string(ms.MSpanInuse)
	mtr["MSpanSys"] = string(ms.MSpanSys)
	mtr["Mallocs"] = string(ms.Mallocs)
	mtr["NextGC"] = string(ms.NextGC)
	mtr["NumForcedGC"] = string(ms.NumForcedGC)
	mtr["NumGC"] = string(ms.NumGC)
	mtr["OtherSys"] = string(ms.OtherSys)
	mtr["PauseTotalNs"] = string(ms.PauseTotalNs)
	mtr["StackInuse"] = string(ms.StackInuse)
	mtr["StackSys"] = string(ms.StackSys)
	mtr["Sys"] = string(ms.Sys)
	mtr["TotalAlloc"] = string(ms.TotalAlloc)
	return mtr, nil
}
