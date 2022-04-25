package ramstorage

import (
	"github.com/resssoft/go-metrics-ya-praktikum/internal/interfaces"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"sync"
)

var data = struct {
	sync.RWMutex
	GaugeData   map[string]models.Gauge
	CounterData map[string]models.Counter
}{
	GaugeData:   make(map[string]models.Gauge),
	CounterData: make(map[string]models.Counter),
}

type simpleRamStorage struct{}

func New() interfaces.Storage {
	return &simpleRamStorage{}
}

func (s *simpleRamStorage) SaveGuage(key string, val models.Gauge) {
	data.Lock()
	data.GaugeData[key] = val
	data.Unlock()
}

func (s *simpleRamStorage) SaveCounter(key string, val models.Counter) {
	data.Lock()
	data.CounterData[key] = val
	data.Unlock()
}

func (s *simpleRamStorage) GetGuages() map[string]models.Gauge {
	result := make(map[string]models.Gauge)
	data.Lock()
	for k, v := range data.GaugeData {
		result[k] = v
	}
	data.Unlock()
	return result
}

func (s *simpleRamStorage) GetCounters() map[string]models.Counter {
	result := make(map[string]models.Counter)
	data.Lock()
	for k, v := range data.CounterData {
		result[k] = v
	}
	data.Unlock()
	return result
}
