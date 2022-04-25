package server

import (
	"bytes"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/interfaces"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"html/template"
	"net/http"
	"strconv"
)

const tmpDefault = `
<html><body><table align="center">
{{range .Values}}
<tr><td> {{.Name}} </td><td> {{.Value}} </td></tr>
{{end}}
</table></body></html>
`

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

type tmpValue struct {
	Name  string
	Value string
}

type tmp struct {
	Values []tmpValue
}

func (ms *MetricsSaver) GetAll(rw http.ResponseWriter, req *http.Request) {
	guages := ms.storage.GetGuages()
	counters := ms.storage.GetCounters()
	var values []tmpValue
	for key, value := range guages {
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
