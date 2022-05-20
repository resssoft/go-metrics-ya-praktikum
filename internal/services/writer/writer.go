package writer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/structure"
	"io/ioutil"
	"os"
	"time"
)

type storeData struct {
	Counters map[string]models.Counter
	Gauges   map[string]models.Gauge
}

type writer struct {
	duration time.Duration
	path     string
	storage  structure.Storage
	ticker   *time.Ticker
}

func New(
	duration time.Duration,
	path string,
	restoreFlag bool,
	storage structure.Storage) structure.Task {
	writerClient := &writer{
		duration: duration,
		storage:  storage,
		path:     path,
	}
	if restoreFlag {
		writerClient.restore()
	}
	return writerClient
}

func (w *writer) Start() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	w.ticker = time.NewTicker(w.duration)
	go w.poll(ctx)
	return cancel
}

func (w *writer) Stop(cancel context.CancelFunc) {
	cancel()
}

func (w *writer) store() {
	fmt.Println("run store data")
	data := storeData{
		Counters: w.storage.GetCounters(),
		Gauges:   w.storage.GetGauges(),
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("json marshal data error: " + err.Error())
	} else {
		err = ioutil.WriteFile(w.path, jsonData, os.ModePerm)
		if err != nil {
			fmt.Println("store data error: " + err.Error())
		}
	}
}

func (w *writer) restore() {
	fmt.Println("run restore data")
	dataJson, err := ioutil.ReadFile(w.path)
	if err != nil {
		fmt.Println("restore data error: " + err.Error())
	} else {
		data := storeData{}
		err = json.Unmarshal(dataJson, &data)
		if err != nil {
			fmt.Println("unmarshal data error: " + err.Error())
		} else {
			for key, val := range data.Counters {
				w.storage.SaveCounter(key, val)
			}
			for key, val := range data.Gauges {
				w.storage.SaveGauge(key, val)
			}
		}
	}
}

func (w *writer) poll(ctx context.Context) {
	fmt.Println("run store poll event spy")
	for {
		select {
		case <-w.ticker.C:
			w.store()
		case <-ctx.Done():
			fmt.Println("run store before exit")
			w.store()
			return
		}
	}
}
