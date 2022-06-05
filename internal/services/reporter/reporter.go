package reporter

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/structure"
	"io/ioutil"
	"net/http"
	"time"

	"crypto/sha256"
)

var (
	reportURL  = "http://%s/update/" // New: http://%s/update/ Old:http://address:port/update/<type>/<name>/<value>
	apiAddress = "127.0.0.1"
)

type Reporter struct {
	Duration  time.Duration
	ticker    *time.Ticker
	storage   structure.Storage
	cryptoKey string
}

func New(
	duration time.Duration,
	address string,
	storage structure.Storage,
	cryptoKey string) structure.Task {
	apiAddress = address
	return &Reporter{
		Duration:  duration,
		storage:   storage,
		cryptoKey: cryptoKey,
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
				err := r.send(metric)
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
				err := r.send(metric)
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

func (r *Reporter) send(metric structure.Metrics) error {
	if r.cryptoKey != "" {
		var hashBody []byte
		switch metric.MType {
		case "counter":
			hashBody = []byte(fmt.Sprintf("%s:counter:%d", metric.ID, metric.Delta))
		case "gauge":
			hashBody = []byte(fmt.Sprintf("%s:gauge:%d", metric.ID, metric.Value))
		}
		h := hmac.New(sha256.New, []byte(r.cryptoKey))
		h.Write(hashBody)
		sha := hex.EncodeToString(h.Sum(nil))
		metric.Hash = sha
	}
	metricJSON, err := json.Marshal(metric)
	if err != nil {
		return err
	}
	fmt.Println(string(metricJSON))
	response, err := http.Post(fmt.Sprintf(
		reportURL,
		apiAddress), "application/json", bytes.NewBuffer(metricJSON))
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
