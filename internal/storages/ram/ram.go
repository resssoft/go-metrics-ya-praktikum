package ramstorage

import (
	"errors"
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

type simpleRAMStorage struct{}

func New() interfaces.Storage {
	return &simpleRAMStorage{}
}

func (s *simpleRAMStorage) SaveGuage(key string, val models.Gauge) {
	data.Lock()
	data.GaugeData[key] = val
	data.Unlock()
}

func (s *simpleRAMStorage) SaveCounter(key string, val models.Counter) {
	data.Lock()
	data.CounterData[key] = val
	data.Unlock()
}

func (s *simpleRAMStorage) GetGuages() map[string]models.Gauge {
	result := make(map[string]models.Gauge)
	data.Lock()
	for k, v := range data.GaugeData {
		result[k] = v
	}
	data.Unlock()
	return result
}

func (s *simpleRAMStorage) GetCounters() map[string]models.Counter {
	result := make(map[string]models.Counter)
	data.Lock()
	for k, v := range data.CounterData {
		result[k] = v
	}
	data.Unlock()
	return result
}

func (s *simpleRAMStorage) IncrementCounter(key string, val models.Counter) {
	data.Lock()
	if current, ok := data.CounterData[key]; ok {
		data.CounterData[key] = val + current
	} else {
		data.CounterData[key] = val
	}
	data.Unlock()
}

func (s *simpleRAMStorage) GetCounter(key string) (models.Counter, error) {
	var err error = nil
	data.Lock()
	value, ok := data.CounterData[key]
	if !ok {
		err = models.ErrorNotFound
	}
	data.Unlock()
	return value, err
}

func (s *simpleRAMStorage) GetGuage(key string) (models.Gauge, error) {
	var err error = nil
	data.Lock()
	value, ok := data.GaugeData[key]
	if !ok {
		err = errors.New("not found")
	}
	data.Unlock()
	return value, err
}
