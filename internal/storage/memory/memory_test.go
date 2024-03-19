package memorystorage

import (
	"context"
	"testing"

	"github.com/ospiem/mcollector/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestMemStorage(t *testing.T) {
	mem := New()

	t.Run("InsertGauge", func(t *testing.T) {
		err := mem.InsertGauge(context.Background(), "test", 1.0)
		assert.NoError(t, err)
	})

	t.Run("InsertCounter", func(t *testing.T) {
		err := mem.InsertCounter(context.Background(), "test", 1)
		assert.NoError(t, err)
	})

	t.Run("SelectGauge", func(t *testing.T) {
		value, err := mem.SelectGauge(context.Background(), "test")
		assert.NoError(t, err)
		assert.Equal(t, 1.0, value)
	})

	t.Run("SelectCounter", func(t *testing.T) {
		value, err := mem.SelectCounter(context.Background(), "test")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), value)
	})

	t.Run("SelectNonExistingGauge", func(t *testing.T) {
		_, err := mem.SelectGauge(context.Background(), "non_existing")
		assert.Error(t, err)
	})

	t.Run("SelectNonExistingCounter", func(t *testing.T) {
		_, err := mem.SelectCounter(context.Background(), "non_existing")
		assert.Error(t, err)
	})

	t.Run("GetCounters", func(t *testing.T) {
		counters, err := mem.GetCounters(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(1), counters["test"])
	})

	t.Run("GetGauges", func(t *testing.T) {
		gauges, err := mem.GetGauges(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1.0, gauges["test"])
	})

	t.Run("InsertBatch", func(t *testing.T) {
		err := mem.InsertBatch(context.Background(), []models.Metrics{
			{ID: "test2", MType: "counter", Delta: new(int64)},
			{ID: "test2", MType: "gauge", Value: new(float64)},
		})
		assert.NoError(t, err)
	})

	t.Run("Ping", func(t *testing.T) {
		err := mem.Ping(context.Background())
		assert.NoError(t, err)
	})
}
