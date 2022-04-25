package handlers

import (
	"fmt"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/interfaces"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"net/http"
	"strconv"
	"strings"
)

type metricsSaver struct {
	storage interfaces.Storage
}

func NewMetricsSaver(storage interfaces.Storage) metricsSaver {
	return metricsSaver{
		storage: storage,
	}
}

func (ms *metricsSaver) SaveMetrics(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		rw.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(rw, "unsupported method")
		return
	}
	uriParts := strings.Split(req.URL.Path, "/")
	if len(uriParts) < 5 {
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rw, "Not found this route")
		return
	}
	dataType := uriParts[2]
	dataName := uriParts[3]
	dataValue := uriParts[4]
	//fmt.Println("data type,name,value: ", dataType, dataName, dataValue)
	switch dataType {
	case "counter":
		valueInt64, err := strconv.ParseInt(dataValue, 10, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rw, "value parsing error: %v", err.Error())
			return
		}
		ms.storage.SaveCounter(dataName, models.Counter(valueInt64))
	case "guage":
		valueFloat64, err := strconv.ParseFloat(dataValue, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rw, "value parsing error: %v", err.Error())
			return
		}
		ms.storage.SaveGuage(dataName, models.Gauge(valueFloat64))
	}
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, "Counters %v", ms.storage.GetCounters())
	fmt.Fprintf(rw, "Gauges %v", ms.storage.GetGuages())
}
