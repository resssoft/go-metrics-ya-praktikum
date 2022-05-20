package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/structure"
	"github.com/rs/zerolog/log"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
)

const tmpDefault = `
<html><body><table align="center">
<tr><th> Name </th><th> Value </th></tr>
{{range .Values}}
<tr><td> {{.Name}} </td><td> {{.Value}} </td></tr>
{{end}}
</table></body></html>
`

type MetricsSaver struct {
	storage structure.Storage
}

func NewMetricsSaver(storage structure.Storage) MetricsSaver {
	return MetricsSaver{
		storage: storage,
	}
}

func (ms *MetricsSaver) SaveGauge(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Path)
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")
	valueFloat64, err := strconv.ParseFloat(value, 64)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "value parsing error: %v", err.Error())
		return
	}
	ms.storage.SaveGauge(name, models.Gauge(valueFloat64))
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

func (ms *MetricsSaver) GetGauge(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Path)
	name := chi.URLParam(req, "name")
	if name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	val, err := ms.storage.GetGauge(name)
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

func (ms *MetricsSaver) SaveValue(rw http.ResponseWriter, req *http.Request) {
	metrics := structure.Metrics{}
	respBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "%v", err.Error())
		return
	}
	err = json.Unmarshal(respBody, &metrics)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "%v", err.Error())
		return
	}
	log.Info().Interface("metrics", metrics).Send()
	switch metrics.MType {
	case "counter":
		if metrics.Delta == nil {
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rw, "Delta is empty")
			return
		}
		ms.storage.IncrementCounter(metrics.ID, models.Counter(*metrics.Delta))
		rw.WriteHeader(http.StatusNoContent)
	case "gauge":
		if metrics.Value == nil {
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rw, "Value is empty")
			return
		}
		ms.storage.SaveGauge(metrics.ID, models.Gauge(*metrics.Value))
		rw.WriteHeader(http.StatusNoContent)
	default:
		rw.WriteHeader(http.StatusForbidden)
		return
	}
}

func (ms *MetricsSaver) GetValue(rw http.ResponseWriter, req *http.Request) {
	metrics := structure.Metrics{}
	respBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "%v", err.Error())
		return
	}
	err = json.Unmarshal(respBody, &metrics)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "%v", err.Error())
		return
	}
	switch metrics.MType {
	case "counter":
		val, err := ms.storage.GetCounter(metrics.ID)
		if err != nil {
			rw.WriteHeader(http.StatusOK)
			fmt.Fprintf(rw, "%v", err.Error())
			return
		}
		fmt.Fprintf(rw, "%v", val)
	case "gauge":
		val, err := ms.storage.GetGauge(metrics.ID)
		if err != nil {
			rw.WriteHeader(http.StatusOK)
			fmt.Fprintf(rw, "%v", err.Error())
			return
		}
		fmt.Fprintf(rw, "%v", val)
	default:
		rw.WriteHeader(http.StatusForbidden)
		return
	}
}

type tmpValue struct {
	Name  string
	Value string
}

type tmp struct {
	Values []tmpValue
}

func (ms *MetricsSaver) GetAll(rw http.ResponseWriter, req *http.Request) {
	gauges := ms.storage.GetGauges()
	counters := ms.storage.GetCounters()
	var values []tmpValue
	for key, value := range gauges {
		values = append(values, tmpValue{
			key,
			fmt.Sprintf("%v", value),
		})
	}
	for key, value := range counters {
		values = append(values, tmpValue{
			key,
			fmt.Sprintf("%v", value),
		})
	}
	tmpData := tmp{values}
	t := template.Must(template.New("").Parse(tmpDefault))
	var tpl bytes.Buffer
	err := t.Execute(&tpl, tmpData)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rw, "%v", err.Error())
		return
	}
	fmt.Fprintf(rw, "%v", tpl.String())
}
