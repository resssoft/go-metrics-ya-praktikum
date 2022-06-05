package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/structure"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	reportURL  = "http://%s/update/" // New: http://%s/update/ Old:http://address:port/update/<type>/<name>/<value>
	apiAddress = "127.0.0.1"
)

type Reporter struct {
	Duration time.Duration
	ticker   *time.Ticker
	storage  structure.Storage
}

func New(
	duration time.Duration,
	address string,
	storage structure.Storage) structure.Task {
	apiAddress = address
	return &Reporter{
		Duration: duration,
		storage:  storage,
	}
}

func (r *Reporter) Start() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	r.ticker = time.NewTicker(r.Duration)
	go r.report(ctx)
	return cancel
}

func (r *Reporter) Stop(cancel context.CancelFunc) {
	cancel()
}

func (r *Reporter) report(ctx context.Context) {
	fmt.Println("run report event spy")
	for {
		select {
		case <-r.ticker.C:
			for name, value := range r.storage.GetGauges() {
				gaugeValue := float64(value)
				metric := structure.Metrics{
					ID:    name,
					MType: "gauge",
					Value: &gaugeValue,
				}
				metricJSON, err := json.Marshal(metric)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = r.send(metricJSON)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			for name, value := range r.storage.GetCounters() {
				deltaValue := int64(value)
				metric := structure.Metrics{
					ID:    name,
					MType: "counter",
					Delta: &deltaValue,
				}
				metricJSON, err := json.Marshal(metric)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = r.send(metricJSON)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		case <-ctx.Done():
			fmt.Println("break report")
			return
		}
	}
}

func (r *Reporter) send(data []byte) error {
	response, err := http.Post(fmt.Sprintf(
		reportURL,
		apiAddress), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = response.Body.Close()
	if err != nil {
		return err
	}
	return nil
}
