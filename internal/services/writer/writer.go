package writer

import (
	"context"
	"encoding/json"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/structure"
	"github.com/rs/zerolog/log"
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
	log.Debug().Msg("run store data")
	data := storeData{
		Counters: w.storage.GetCounters(),
		Gauges:   w.storage.GetGauges(),
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Info().AnErr("unmarshal data error", err).Send()
	} else {
		err = ioutil.WriteFile(w.path, jsonData, os.ModePerm)
		if err != nil {
			log.Info().AnErr("restore data error", err).Send()
		}
	}
}

func (w *writer) restore() {
	log.Debug().Msg("run restore data")
	dataJSON, err := ioutil.ReadFile(w.path)
	if err != nil {
		log.Info().AnErr("restore data error", err).Send()
	} else {
		data := storeData{}
		err = json.Unmarshal(dataJSON, &data)
		if err != nil {
			log.Info().AnErr("unmarshal data error", err).Send()
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
	log.Debug().Msg("run store poll event spy")
	for {
		select {
		case <-w.ticker.C:
			w.store()
		case <-ctx.Done():
			log.Debug().Msg("run store before exit")
			w.store()
			return
		}
	}
}
