// Package models provides structures for working with metrics.
package models

type Metrics struct {
	Delta *int64   `json:"delta,omitempty"` // metric value in case of counter transfer
	Value *float64 `json:"value,omitempty"` // metric value in case of gauge transfer
	ID    string   `json:"id"`              // metric name
	MType string   `json:"type"`            // parameter taking the value gauge or counter
}

const Counter = "counter"
const Gauge = "gauge"
