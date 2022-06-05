package structure

import (
	"context"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
)

type Task interface {
	Start() context.CancelFunc
	Stop(context.CancelFunc)
}

type Storage interface {
	SaveGauge(string, models.Gauge)
	SaveCounter(string, models.Counter)
	GetGauges() map[string]models.Gauge
	GetCounters() map[string]models.Counter
	IncrementCounter(string, models.Counter)
	GetCounter(string) (models.Counter, error)
	GetGauge(string) (models.Gauge, error)
}

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
