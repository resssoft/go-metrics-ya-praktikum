package reporter

import (
	"context"
	"fmt"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/interfaces"
	"io/ioutil"
	"net/http"
	"time"
)

const workers = 1

var (
	reportURL      = "http://%s:%s/update/%s/%s/%v" // http://address:port/update/<type>/<name>/<value>
	defaultAddress = "127.0.0.1"
	defaultPort    = "8080"
)

type Reporter struct {
	Duration time.Duration
	ticker   *time.Ticker
	ctx      context.Context
	cancel   context.CancelFunc
	storage  interfaces.Storage
	exitChan chan int
}

func New(
	duration time.Duration,
	storage interfaces.Storage) interfaces.Task {
	ctx, cancel := context.WithCancel(context.Background())
	return &Reporter{
		ctx:      ctx,
		Duration: duration,
		cancel:   cancel,
		storage:  storage,
		exitChan: make(chan int),
	}
}

func (r *Reporter) Start() {
	r.ticker = time.NewTicker(r.Duration)
	for i := 0; i < workers; i++ {
		go r.report()
	}

}

func (r *Reporter) Stop() {
	r.cancel()
	for i := 0; i < workers; i++ {
		<-r.exitChan
	}
}

func (r *Reporter) report() {
	fmt.Println("run report event spy")
	for {
		select {
		case <-r.ticker.C:
			for name, value := range r.storage.GetGuages() {
				response, err := http.Post(fmt.Sprintf(
					reportURL,
					defaultAddress,
					defaultPort,
					"guage",
					name,
					value), "text/plain", nil)
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
		case <-r.ctx.Done():
			fmt.Println("break report")
			r.exitChan <- 1
			return
		}
	}
}
