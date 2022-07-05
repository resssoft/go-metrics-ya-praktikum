package ramstorage

import (
	"errors"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"sync"
)

type memStorage struct {
	sync.RWMutex
	GaugeData   map[string]models.Gauge
	CounterData map[string]models.Counter
}

type simpleRAMStorage struct {
	storage memStorage
}

func New() *simpleRAMStorage {
	return &simpleRAMStorage{
		storage: memStorage{
			GaugeData:   make(map[string]models.Gauge),
			CounterData: make(map[string]models.Counter),
		},
	}
}

func (s *simpleRAMStorage) SaveGauge(key string, val models.Gauge) {
	s.storage.Lock()
	s.storage.GaugeData[key] = val
	s.storage.Unlock()
}

func (s *simpleRAMStorage) SaveCounter(key string, val models.Counter) {
	s.storage.Lock()
	s.storage.CounterData[key] = val
	s.storage.Unlock()
}

func (s *simpleRAMStorage) GetGauges() map[string]models.Gauge {
	result := make(map[string]models.Gauge)
	s.storage.Lock()
	for k, v := range s.storage.GaugeData {
		result[k] = v
	}
	s.storage.Unlock()
	return result
}

func (s *simpleRAMStorage) GetCounters() map[string]models.Counter {
	result := make(map[string]models.Counter)
	s.storage.Lock()
	for k, v := range s.storage.CounterData {
		result[k] = v
	}
	s.storage.Unlock()
	return result
}

func (s *simpleRAMStorage) IncrementCounter(key string, val models.Counter) {
	s.storage.Lock()
	if current, ok := s.storage.CounterData[key]; ok {
		s.storage.CounterData[key] = val + current
	} else {
		s.storage.CounterData[key] = val
	}
	s.storage.Unlock()
}

func (s *simpleRAMStorage) GetCounter(key string) (models.Counter, error) {
	var err error
	s.storage.Lock()
	value, ok := s.storage.CounterData[key]
	if !ok {
		err = models.ErrNotFound
	}
	s.storage.Unlock()
	return value, err
}

func (s *simpleRAMStorage) GetGauge(key string) (models.Gauge, error) {
	var err error
	s.storage.Lock()
	value, ok := s.storage.GaugeData[key]
	if !ok {
		err = errors.New("not found")
	}
	s.storage.Unlock()
	return value, err
}

func (s *simpleRAMStorage) Ping() string {
	return "ram"
}
