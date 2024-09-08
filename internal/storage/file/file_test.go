package file

import (
	"context"
	"testing"
	"time"

	"github.com/ospiem/mcollector/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewFileStorageWithValidParameters(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test_file"
	restore := true
	storeInterval := time.Second * 5

	_, err := New(ctx, fileStoragePath, restore, storeInterval)

	assert.NoError(t, err)
}

func TestNewFileStorageWithZeroStoreInterval(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test_file"
	restore := true
	storeInterval := time.Second * 0

	_, err := New(ctx, fileStoragePath, restore, storeInterval)

	assert.NoError(t, err)
}

func TestInsertGaugeWithValidParameters(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test_file"
	restore := true
	storeInterval := time.Second * 5

	fs, _ := New(ctx, fileStoragePath, restore, storeInterval)

	err := fs.InsertGauge(ctx, "test_key", 1.0)

	assert.NoError(t, err)
}

func TestInsertCounterWithValidParameters(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test_file"
	restore := true
	storeInterval := time.Second * 5

	fs, _ := New(ctx, fileStoragePath, restore, storeInterval)

	err := fs.InsertCounter(ctx, "test_key", 1)

	assert.NoError(t, err)
}

func TestSelectGaugeWithNonExistingKey(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test_file"
	restore := true
	storeInterval := time.Second * 5

	fs, _ := New(ctx, fileStoragePath, restore, storeInterval)

	_, err := fs.SelectGauge(ctx, "non_existing_key")

	assert.Error(t, err)
}

func TestSelectCounterWithNonExistingKey(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test_file"
	restore := true
	storeInterval := time.Second * 5

	fs, _ := New(ctx, fileStoragePath, restore, storeInterval)

	_, err := fs.SelectCounter(ctx, "non_existing_key")

	assert.Error(t, err)
}

func TestGetCounters(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test_file"
	restore := true
	storeInterval := time.Second * 5

	fs, _ := New(ctx, fileStoragePath, restore, storeInterval)

	_, err := fs.GetCounters(ctx)

	assert.NoError(t, err)
}

func TestGetGauges(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test_file"
	restore := true
	storeInterval := time.Second * 5

	fs, _ := New(ctx, fileStoragePath, restore, storeInterval)

	_, err := fs.GetGauges(ctx)

	assert.NoError(t, err)
}

func TestInsertBatch(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test_file"
	restore := true
	storeInterval := time.Second * 5

	fs, _ := New(ctx, fileStoragePath, restore, storeInterval)

	metrics := []models.Metrics{
		{
			MType: "counter",
			ID:    "test_counter",
			Delta: new(int64),
		},
		{
			MType: "gauge",
			ID:    "test_gauge",
			Value: new(float64),
		},
	}

	err := fs.InsertBatch(ctx, metrics)

	assert.NoError(t, err)
}

func TestPing(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test_file"
	restore := true
	storeInterval := time.Second * 5

	fs, _ := New(ctx, fileStoragePath, restore, storeInterval)

	err := fs.Ping(ctx)

	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test_file"
	restore := true
	storeInterval := time.Second * 5

	fs, _ := New(ctx, fileStoragePath, restore, storeInterval)

	err := fs.Close(ctx)

	assert.NoError(t, err)
}
