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

const workers = 1

var (
	reportURL      = "http://%s:%s/update/" // New: http://%s:%s/update/ Old:http://address:port/update/<type>/<name>/<value>
	defaultAddress = "127.0.0.1"
	defaultPort    = "8080"
)

type Reporter struct {
	Duration time.Duration
	ticker   *time.Ticker
	storage  structure.Storage
	exitChan chan int
}

func New(
	duration time.Duration,
	storage structure.Storage) structure.Task {
	return &Reporter{
		Duration: duration,
		storage:  storage,
		exitChan: make(chan int),
	}
}

func (r *Reporter) Start() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	r.ticker = time.NewTicker(r.Duration)
	for i := 0; i < workers; i++ {
		go r.report(ctx)
	}
	return cancel
}

func (r *Reporter) Stop(cancel context.CancelFunc) {
	cancel()
	for i := 0; i < workers; i++ {
		<-r.exitChan
	}
}

func (r *Reporter) report(ctx context.Context) {
	fmt.Println("run report event spy")
	for {
		select {
		case <-r.ticker.C:
			for name, value := range r.storage.GetGuages() {
				guageValue := float64(value)
				metric := structure.Metrics{
					ID:    name,
					MType: "gauge",
					Value: &guageValue,
				}
				metricJson, err := json.Marshal(metric)
				if err != nil {
					fmt.Println(err)
					return
				}
				response, err := http.Post(fmt.Sprintf(
					reportURL,
					defaultAddress,
					defaultPort), "application/json", bytes.NewBuffer(metricJson))
				if err != nil {
					fmt.Println(err)
					return
				}
				_, err = ioutil.ReadAll(response.Body)
				if err != nil {
					return
				}
				response.Body.Close()
			}

			for name, value := range r.storage.GetCounters() {
				response, err := http.Post(fmt.Sprintf(
					reportURL,
					defaultAddress,
					defaultPort,
					"counter",
					name,
					value), "text/plain", nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("report data", name, value) //TODO: remove after check
				_, err = ioutil.ReadAll(response.Body)
				if err != nil {
					return
				}
				response.Body.Close()
			}
		case <-ctx.Done():
			fmt.Println("break report")
			r.exitChan <- 1
			return
		}
	}
}
