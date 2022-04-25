package poller

import (
	"context"
	"fmt"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/interfaces"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"math/rand"
	"runtime"
	"time"
)

const workers = 1

type Poller struct {
	Duration time.Duration
	ticker   *time.Ticker
	ctx      context.Context
	cancel   context.CancelFunc
	iterator int
	storage  interfaces.Storage
	exitChan chan int
}

func New(
	duration time.Duration,
	storage interfaces.Storage) interfaces.Task {
	ctx, cancel := context.WithCancel(context.Background())
	return &Poller{
		ctx:      ctx,
		Duration: duration,
		cancel:   cancel,
		storage:  storage,
		exitChan: make(chan int),
	}
}

func (p *Poller) Start() {
	p.ticker = time.NewTicker(p.Duration)
	for i := 0; i < workers; i++ {
		go p.poll()
	}

}

func (p *Poller) Stop() {
	p.cancel()
	for i := 0; i < workers; i++ {
		<-p.exitChan
	}
}

func (p *Poller) poll() {
	fmt.Println("run poll event spy")
	var m runtime.MemStats
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	for {
		select {
		case <-p.ticker.C:
			runtime.ReadMemStats(&m)
			p.storage.SaveGuage("Alloc", models.Gauge(m.Alloc))
			p.storage.SaveGuage("Alloc", models.Gauge(m.Alloc))
			p.storage.SaveGuage("BuckHashSys", models.Gauge(m.BuckHashSys))
			p.storage.SaveGuage("Frees", models.Gauge(m.Frees))
			p.storage.SaveGuage("GCCPUFraction", models.Gauge(m.GCCPUFraction))
			p.storage.SaveGuage("GCSys", models.Gauge(m.GCSys))
			p.storage.SaveGuage("HeapAlloc", models.Gauge(m.HeapAlloc))
			p.storage.SaveGuage("HeapIdle", models.Gauge(m.HeapIdle))
			p.storage.SaveGuage("HeapInuse", models.Gauge(m.HeapInuse))
			p.storage.SaveGuage("HeapObjects", models.Gauge(m.HeapObjects))
			p.storage.SaveGuage("HeapReleased", models.Gauge(m.HeapReleased))
			p.storage.SaveGuage("HeapSys", models.Gauge(m.HeapSys))
			p.storage.SaveGuage("LastGC", models.Gauge(m.LastGC))
			p.storage.SaveGuage("Lookups", models.Gauge(m.Lookups))
			p.storage.SaveGuage("MCacheInuse", models.Gauge(m.MCacheInuse))
			p.storage.SaveGuage("MCacheSys", models.Gauge(m.MCacheSys))
			p.storage.SaveGuage("MSpanInuse", models.Gauge(m.MSpanInuse))
			p.storage.SaveGuage("MSpanSys", models.Gauge(m.MSpanSys))
			p.storage.SaveGuage("Mallocs", models.Gauge(m.Mallocs))
			p.storage.SaveGuage("NextGC", models.Gauge(m.NextGC))
			p.storage.SaveGuage("NumForcedGC", models.Gauge(m.NumForcedGC))
			p.storage.SaveGuage("NumGC", models.Gauge(m.NumGC))
			p.storage.SaveGuage("OtherSys", models.Gauge(m.OtherSys))
			p.storage.SaveGuage("PauseTotalNs", models.Gauge(m.PauseTotalNs))
			p.storage.SaveGuage("StackInuse", models.Gauge(m.StackInuse))
			p.storage.SaveGuage("StackSys", models.Gauge(m.StackSys))
			p.storage.SaveGuage("Sys", models.Gauge(m.Sys))
			p.storage.SaveGuage("TotalAlloc", models.Gauge(m.TotalAlloc))

			p.iterator++
			p.storage.SaveCounter("PollCount", models.Counter(p.iterator))
			p.storage.SaveCounter("RandomValue", models.Counter(r1.Int()))

			fmt.Println("poll iterator", p.iterator)

		case <-p.ctx.Done():
			fmt.Println("break poll")
			p.exitChan <- 1
			return
		}
	}
}