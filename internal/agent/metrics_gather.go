package agent

import (
	"runtime"
	"strconv"
)

func GetMetricsOld(ms *runtime.MemStats, mtr map[string]string) error {
	runtime.ReadMemStats(ms)
	mtr["Alloc"] = strconv.FormatUint(ms.Alloc, 10)
	mtr["BuckHashSys"] = strconv.FormatUint(ms.BuckHashSys, 10)
	mtr["Frees"] = strconv.FormatUint(ms.Frees, 10)
	mtr["GCCPUFraction"] = strconv.FormatFloat(ms.GCCPUFraction, 'f', -1, 64)
	mtr["GCSys"] = strconv.FormatUint(ms.GCSys, 10)
	mtr["HeapAlloc"] = strconv.FormatUint(ms.HeapAlloc, 10)
	mtr["HeapIdle"] = strconv.FormatUint(ms.HeapIdle, 10)
	mtr["HeapInuse"] = strconv.FormatUint(ms.HeapInuse, 10)
	mtr["HeapObjects"] = strconv.FormatUint(ms.HeapObjects, 10)
	mtr["HeapReleased"] = strconv.FormatUint(ms.HeapReleased, 10)
	mtr["HeapSys"] = strconv.FormatUint(ms.HeapSys, 10)
	mtr["LastGC"] = strconv.FormatUint(ms.LastGC, 10)
	mtr["Lookups"] = strconv.FormatUint(ms.Lookups, 10)
	mtr["MCacheInuse"] = strconv.FormatUint(ms.MCacheInuse, 10)
	mtr["MCacheSys"] = strconv.FormatUint(ms.MCacheSys, 10)
	mtr["MSpanInuse"] = strconv.FormatUint(ms.MSpanInuse, 10)
	mtr["MSpanSys"] = strconv.FormatUint(ms.MSpanSys, 10)
	mtr["Mallocs"] = strconv.FormatUint(ms.Mallocs, 10)
	mtr["NextGC"] = strconv.FormatUint(ms.NextGC, 10)
	mtr["NumForcedGC"] = strconv.Itoa(int(ms.NumForcedGC))
	mtr["NumGC"] = strconv.Itoa(int(ms.NumGC))
	mtr["OtherSys"] = strconv.FormatUint(ms.OtherSys, 10)
	mtr["PauseTotalNs"] = strconv.FormatUint(ms.PauseTotalNs, 10)
	mtr["StackInuse"] = strconv.FormatUint(ms.StackInuse, 10)
	mtr["StackSys"] = strconv.FormatUint(ms.StackSys, 10)
	mtr["Sys"] = strconv.FormatUint(ms.Sys, 10)
	mtr["TotalAlloc"] = strconv.FormatUint(ms.TotalAlloc, 10)
	return nil
}

func GetMetrics() (map[string]string, error) {
	var ms *runtime.MemStats
	mtr := make(map[string]string)
	runtime.ReadMemStats(ms)
	mtr["Alloc"] = strconv.FormatUint(ms.Alloc, 10)
	mtr["BuckHashSys"] = strconv.FormatUint(ms.BuckHashSys, 10)
	mtr["Frees"] = strconv.FormatUint(ms.Frees, 10)
	mtr["GCCPUFraction"] = strconv.FormatFloat(ms.GCCPUFraction, 'f', -1, 64)
	mtr["GCSys"] = strconv.FormatUint(ms.GCSys, 10)
	mtr["HeapAlloc"] = strconv.FormatUint(ms.HeapAlloc, 10)
	mtr["HeapIdle"] = strconv.FormatUint(ms.HeapIdle, 10)
	mtr["HeapInuse"] = strconv.FormatUint(ms.HeapInuse, 10)
	mtr["HeapObjects"] = strconv.FormatUint(ms.HeapObjects, 10)
	mtr["HeapReleased"] = strconv.FormatUint(ms.HeapReleased, 10)
	mtr["HeapSys"] = strconv.FormatUint(ms.HeapSys, 10)
	mtr["LastGC"] = strconv.FormatUint(ms.LastGC, 10)
	mtr["Lookups"] = strconv.FormatUint(ms.Lookups, 10)
	mtr["MCacheInuse"] = strconv.FormatUint(ms.MCacheInuse, 10)
	mtr["MCacheSys"] = strconv.FormatUint(ms.MCacheSys, 10)
	mtr["MSpanInuse"] = strconv.FormatUint(ms.MSpanInuse, 10)
	mtr["MSpanSys"] = strconv.FormatUint(ms.MSpanSys, 10)
	mtr["Mallocs"] = strconv.FormatUint(ms.Mallocs, 10)
	mtr["NextGC"] = strconv.FormatUint(ms.NextGC, 10)
	mtr["NumForcedGC"] = strconv.Itoa(int(ms.NumForcedGC))
	mtr["NumGC"] = strconv.Itoa(int(ms.NumGC))
	mtr["OtherSys"] = strconv.FormatUint(ms.OtherSys, 10)
	mtr["PauseTotalNs"] = strconv.FormatUint(ms.PauseTotalNs, 10)
	mtr["StackInuse"] = strconv.FormatUint(ms.StackInuse, 10)
	mtr["StackSys"] = strconv.FormatUint(ms.StackSys, 10)
	mtr["Sys"] = strconv.FormatUint(ms.Sys, 10)
	mtr["TotalAlloc"] = strconv.FormatUint(ms.TotalAlloc, 10)
	return mtr, nil
}
