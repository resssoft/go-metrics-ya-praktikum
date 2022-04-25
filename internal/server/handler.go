package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/interfaces"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"net/http"
	"strconv"
)

type MetricsSaver struct {
	storage interfaces.Storage
}

func NewMetricsSaver(storage interfaces.Storage) MetricsSaver {
	return MetricsSaver{
		storage: storage,
	}
}

func (ms *MetricsSaver) SaveGuage(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Path)
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")
	valueFloat64, err := strconv.ParseFloat(value, 64)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "value parsing error: %v", err.Error())
		return
	}
	ms.storage.SaveGuage(name, models.Gauge(valueFloat64))
}

func (ms *MetricsSaver) SaveCounter(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Path)
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")
	valueInt64, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "value parsing error: %v", err.Error())
		return
	}
	ms.storage.IncrementCounter(name, models.Counter(valueInt64))
}

func (ms *MetricsSaver) GetGuage(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Path)
	name := chi.URLParam(req, "name")
	if name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	val, err := ms.storage.GetGuage(name)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rw, "%v", err.Error())
		return
	}
	fmt.Fprintf(rw, "%v", val)
}

func (ms *MetricsSaver) GetCounter(rw http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")
	if name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	val, err := ms.storage.GetCounter(name)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rw, "%v", err.Error())
		return
	}
	fmt.Fprintf(rw, "%v", val)
}

func (ms *MetricsSaver) GetAll(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "%v", 1)
}
