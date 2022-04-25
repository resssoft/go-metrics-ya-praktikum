package interfaces

import "github.com/resssoft/go-metrics-ya-praktikum/internal/models"

type Task interface {
	Start()
	Stop()
}

type Storage interface {
	SaveGuage(string, models.Gauge)
	SaveCounter(string, models.Counter)
	GetGuages() map[string]models.Gauge
	GetCounters() map[string]models.Counter
}
