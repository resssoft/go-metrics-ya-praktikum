package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

type errResponse struct {
	Error string
}

type MetricsSaver struct {
	storage   structure.Storage
	cryptoKey string
	dbAddress string
}

func NewMetricsSaver(storage structure.Storage, cryptoKey, dbAddress string) MetricsSaver {
	return MetricsSaver{
		storage:   storage,
		cryptoKey: cryptoKey,
		dbAddress: dbAddress,
	}
}

func (ms *MetricsSaver) SaveGauge(rw http.ResponseWriter, req *http.Request) {
	log.Info().Msgf("SaveGauge path: %s", req.URL.Path)
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
	log.Info().Msgf("SaveCounter path: %s", req.URL.Path)
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
	log.Info().Msgf("GetGauge path: %s", req.URL.Path)
	name := chi.URLParam(req, "name")
	if name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	val, err := ms.storage.GetGauge(name)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
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
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, "%v", val)
}

func (ms *MetricsSaver) SaveValue(rw http.ResponseWriter, req *http.Request) {
	metrics := structure.Metrics{}
	respBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusInternalServerError)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info().Msgf("SaveValue path: %s, body %s", req.URL.Path, string(respBody))
	err = json.Unmarshal(respBody, &metrics)
	if err != nil {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info().Interface("metrics", metrics).Send()
	if metrics.Hash != "" && ms.cryptoKey != "" {
		sha := ms.getMetricsHash(metrics)
		decodedHex, err := hex.DecodeString(metrics.Hash)
		if err != nil {
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.Error(rw, "Hash can`t be decoded", http.StatusBadRequest)
			return
		}
		if !hmac.Equal(sha, decodedHex) {
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.Error(rw, "Hash sum is not true", http.StatusBadRequest)
			return
		}
	}
	switch metrics.MType {
	case structure.CounterType:
		if metrics.Delta == nil {
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			rw.WriteHeader(http.StatusBadRequest)
			http.Error(rw, "Delta is empty", http.StatusInternalServerError)
			return
		}
		ms.storage.IncrementCounter(metrics.ID, models.Counter(*metrics.Delta))
		rw.WriteHeader(http.StatusOK)
	case structure.GaugeType:
		if metrics.Value == nil {
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			rw.WriteHeader(http.StatusBadRequest)
			http.Error(rw, "Value is empty", http.StatusInternalServerError)
			return
		}
		ms.storage.SaveGauge(metrics.ID, models.Gauge(*metrics.Value))
		rw.WriteHeader(http.StatusOK)
	default:
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusForbidden)
		return
	}
}

func (ms *MetricsSaver) GetValue(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	metrics := structure.Metrics{}
	respBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, getErr(err.Error()), http.StatusInternalServerError)
		return
	}
	log.Info().Msgf("GetValue path: %s, body %s", req.URL.Path, string(respBody))
	err = json.Unmarshal(respBody, &metrics)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Header().Set("X-Content-Type-Options", "nosniff")
		fmt.Fprintln(rw, getErr(err.Error()))
		//http.Error(rw, getErr(err.Error()), http.StatusBadRequest)
		return
	}
	switch metrics.MType {
	case structure.CounterType:
		val, err := ms.storage.GetCounter(metrics.ID)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Header().Set("X-Content-Type-Options", "nosniff")
			fmt.Fprintln(rw, getErr(err.Error()))
			//http.Error(rw, getErr(err.Error()), http.StatusNotFound)
			return
		}
		intVal := int64(val)
		metrics.Delta = &intVal
		if ms.cryptoKey != "" {
			metrics.Hash = hex.EncodeToString(ms.getMetricsHash(metrics))
		}
		metricJSON, err := json.Marshal(metrics)
		if err != nil {
			rw.WriteHeader(http.StatusForbidden)
			rw.Header().Set("X-Content-Type-Options", "nosniff")
			fmt.Fprintln(rw, getErr(err.Error()))
			//http.Error(rw, getErr(err.Error()), http.StatusForbidden)
			return
		}
		fmt.Fprintf(rw, "%s", string(metricJSON))
	case structure.GaugeType:
		val, err := ms.storage.GetGauge(metrics.ID)
		floatVal := float64(val)
		metrics.Value = &floatVal
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Header().Set("X-Content-Type-Options", "nosniff")
			fmt.Fprintln(rw, getErr(err.Error()))
			//http.Error(rw, getErr(err.Error()), http.StatusForbidden)
			return
		}
		if ms.cryptoKey != "" {
			metrics.Hash = hex.EncodeToString(ms.getMetricsHash(metrics))
		}
		metricJSON, err := json.Marshal(metrics)
		if err != nil {
			rw.WriteHeader(http.StatusForbidden)
			rw.Header().Set("X-Content-Type-Options", "nosniff")
			fmt.Fprintln(rw, getErr(err.Error()))
			//http.Error(rw, getErr(err.Error()), http.StatusForbidden)
			return
		}
		fmt.Fprintf(rw, "%s", string(metricJSON))
	default:
		rw.WriteHeader(http.StatusForbidden)
		rw.Header().Set("X-Content-Type-Options", "nosniff")
		fmt.Fprintln(rw, getErr("unsupported type"))
		//http.Error(rw, getErr("unsupported type"), http.StatusForbidden)
		return
	}
}

func (ms *MetricsSaver) getMetricsHash(metrics structure.Metrics) []byte {
	var hashBody []byte
	switch metrics.MType {
	case structure.CounterType:
		hashBody = []byte(fmt.Sprintf("%s:counter:%d", metrics.ID, getSafelyDelta(metrics.Delta)))
	case structure.GaugeType:
		hashBody = []byte(fmt.Sprintf("%s:gauge:%f", metrics.ID, getSafelyValue(metrics.Value)))
	}
	h := hmac.New(sha256.New, []byte(ms.cryptoKey))
	h.Write(hashBody)
	return h.Sum(nil)
}

func (ms *MetricsSaver) SaveValues(rw http.ResponseWriter, req *http.Request) {
	var metricsList []structure.Metrics
	respBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusInternalServerError)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info().Msgf("SaveValue path: %s, body %s", req.URL.Path, string(respBody))
	err = json.Unmarshal(respBody, &metricsList)
	if err != nil {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info().Interface("metrics", metricsList).Send()
	for _, metrics := range metricsList {
		if metrics.Hash != "" && ms.cryptoKey != "" {
			sha := ms.getMetricsHash(metrics)
			decodedHex, err := hex.DecodeString(metrics.Hash)
			if err != nil {
				rw.Header().Set("Content-Type", "text/html; charset=utf-8")
				http.Error(rw, "Hash can`t be decoded", http.StatusBadRequest)
				return
			}
			if !hmac.Equal(sha, decodedHex) {
				rw.Header().Set("Content-Type", "text/html; charset=utf-8")
				http.Error(rw, "Hash sum is not true", http.StatusBadRequest)
				return
			}
		}
		switch metrics.MType {
		case structure.CounterType:
			if metrics.Delta == nil {
				rw.Header().Set("Content-Type", "text/html; charset=utf-8")
				rw.WriteHeader(http.StatusBadRequest)
				http.Error(rw, "Delta is empty", http.StatusInternalServerError)
				return
			}
			ms.storage.IncrementCounter(metrics.ID, models.Counter(*metrics.Delta))
			rw.WriteHeader(http.StatusOK)
		case structure.GaugeType:
			if metrics.Value == nil {
				rw.Header().Set("Content-Type", "text/html; charset=utf-8")
				rw.WriteHeader(http.StatusBadRequest)
				http.Error(rw, "Value is empty", http.StatusInternalServerError)
				return
			}
			ms.storage.SaveGauge(metrics.ID, models.Gauge(*metrics.Value))
			rw.WriteHeader(http.StatusOK)
		default:
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			rw.WriteHeader(http.StatusForbidden)
			return
		}
	}
}

func getSafelyDelta(link *int64) int64 {
	if link == nil {
		return 0
	}
	return *link
}

func getSafelyValue(link *float64) float64 {
	if link == nil {
		return 0.0
	}
	return *link
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
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(rw, "%v", tpl.String())
}

func (ms *MetricsSaver) DBPing(rw http.ResponseWriter, req *http.Request) {
	result := ms.storage.Ping()
	if result == "" {
		rw.WriteHeader(http.StatusOK)
	} else {
		http.Error(rw, result, http.StatusInternalServerError)
	}
}

func (ms *MetricsSaver) h501(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusNotImplemented)
}

func (ms *MetricsSaver) h404(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusNotFound)
}

func getErr(msg string) string {
	errObj := errResponse{
		Error: msg,
	}
	errObjJSON, err := json.Marshal(errObj)
	if err != nil {
		return fmt.Sprintf("{Error:\"%s\"}", err.Error())
	}
	return string(errObjJSON)
}
