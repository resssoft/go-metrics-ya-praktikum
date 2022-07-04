package poller

import (
	"context"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/structure"
	"github.com/rs/zerolog/log"
	"math/rand"
	"runtime"
	"time"
)

type Poller struct {
	Duration time.Duration
	ticker   *time.Ticker
	iterator int
	storage  structure.Storage
	randoms  *rand.Rand
}

func New(
	duration time.Duration,
	storage structure.Storage) structure.Task {
	s1 := rand.NewSource(time.Now().UnixNano())
	return &Poller{
		randoms:  rand.New(s1),
		Duration: duration,
		storage:  storage,
	}
}

func (p *Poller) Start() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	p.ticker = time.NewTicker(p.Duration)
	go p.pollHandler(ctx)
	p.poll()
	return cancel
}

func (p *Poller) Stop(cancel context.CancelFunc) {
	cancel()
}

func (p *Poller) pollHandler(ctx context.Context) {
	log.Info().Msg("run poll event spy")
	for {
		select {
		case <-p.ticker.C:
			p.poll()
			log.Debug().Msg("poll iterator")
		case <-ctx.Done():
			log.Info().Msg("break poll")
			return
		}
	}
}

func (p *Poller) poll() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	p.storage.SaveGauge("Alloc", models.Gauge(m.Alloc))
	p.storage.SaveGauge("Alloc", models.Gauge(m.Alloc))
	p.storage.SaveGauge("BuckHashSys", models.Gauge(m.BuckHashSys))
	p.storage.SaveGauge("Frees", models.Gauge(m.Frees))
	p.storage.SaveGauge("GCCPUFraction", models.Gauge(m.GCCPUFraction))
	p.storage.SaveGauge("GCSys", models.Gauge(m.GCSys))
	p.storage.SaveGauge("HeapAlloc", models.Gauge(m.HeapAlloc))
	p.storage.SaveGauge("HeapIdle", models.Gauge(m.HeapIdle))
	p.storage.SaveGauge("HeapInuse", models.Gauge(m.HeapInuse))
	p.storage.SaveGauge("HeapObjects", models.Gauge(m.HeapObjects))
	p.storage.SaveGauge("HeapReleased", models.Gauge(m.HeapReleased))
	p.storage.SaveGauge("HeapSys", models.Gauge(m.HeapSys))
	p.storage.SaveGauge("LastGC", models.Gauge(m.LastGC))
	p.storage.SaveGauge("Lookups", models.Gauge(m.Lookups))
	p.storage.SaveGauge("MCacheInuse", models.Gauge(m.MCacheInuse))
	p.storage.SaveGauge("MCacheSys", models.Gauge(m.MCacheSys))
	p.storage.SaveGauge("MSpanInuse", models.Gauge(m.MSpanInuse))
	p.storage.SaveGauge("MSpanSys", models.Gauge(m.MSpanSys))
	p.storage.SaveGauge("Mallocs", models.Gauge(m.Mallocs))
	p.storage.SaveGauge("NextGC", models.Gauge(m.NextGC))
	p.storage.SaveGauge("NumForcedGC", models.Gauge(m.NumForcedGC))
	p.storage.SaveGauge("NumGC", models.Gauge(m.NumGC))
	p.storage.SaveGauge("OtherSys", models.Gauge(m.OtherSys))
	p.storage.SaveGauge("PauseTotalNs", models.Gauge(m.PauseTotalNs))
	p.storage.SaveGauge("StackInuse", models.Gauge(m.StackInuse))
	p.storage.SaveGauge("StackSys", models.Gauge(m.StackSys))
	p.storage.SaveGauge("Sys", models.Gauge(m.Sys))
	p.storage.SaveGauge("TotalAlloc", models.Gauge(m.TotalAlloc))
	p.storage.SaveGauge("RandomValue", models.Gauge(p.randoms.Float64()))

	p.iterator++
	p.storage.SaveCounter("PollCount", models.Counter(p.iterator))
}
