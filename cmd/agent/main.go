package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type gauge float64
type counter int64

var (
	exitChan       chan int
	pollInterval   = time.Second * 22
	reportInterval = time.Second * 990
	reportUrl      = "http://%s:%s/update/%s/%s/%v" // http://address:port/update/<type>/<name>/<value>
	defaultAddress = "127.0.0.1"
	defaultPort    = "8080"
	data           = struct {
		sync.RWMutex
		GaugeData   map[string]gauge
		CounterData map[string]counter
		Iterator    int64
	}{
		GaugeData:   make(map[string]gauge),
		CounterData: make(map[string]counter),
	}
)

func main() {
	fmt.Println("Start agent")
	exitChan = make(chan int)
	ctx, cancel := context.WithCancel(context.Background())
	updateDataTicker := time.NewTicker(pollInterval)
	sendDataTicker := time.NewTicker(reportInterval)
	go poll(ctx, updateDataTicker)
	go report(ctx, sendDataTicker)

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)

	s := <-signalChanel
	switch s {
	case syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM:
		fmt.Println("Signal quit triggered.")
		cancel()
	default:
		fmt.Println("Unknown signal.")
	}
	<-exitChan
	<-exitChan
}

func poll(ctx context.Context, ticker *time.Ticker) {
	fmt.Println("run poll event spy")
	var m runtime.MemStats
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&m)
			data.Lock()
			data.GaugeData["Alloc"] = gauge(m.Alloc)
			data.GaugeData["BuckHashSys"] = gauge(m.BuckHashSys)
			data.GaugeData["Frees"] = gauge(m.Frees)
			data.GaugeData["GCCPUFraction"] = gauge(m.GCCPUFraction)
			data.GaugeData["GCSys"] = gauge(m.GCSys)
			data.GaugeData["HeapAlloc"] = gauge(m.HeapAlloc)
			data.GaugeData["HeapIdle"] = gauge(m.HeapIdle)
			data.GaugeData["HeapInuse"] = gauge(m.HeapInuse)
			data.GaugeData["HeapObjects"] = gauge(m.HeapObjects)
			data.GaugeData["HeapReleased"] = gauge(m.HeapReleased)
			data.GaugeData["HeapSys"] = gauge(m.HeapSys)
			data.GaugeData["LastGC"] = gauge(m.LastGC)
			data.GaugeData["Lookups"] = gauge(m.Lookups)
			data.GaugeData["MCacheInuse"] = gauge(m.MCacheInuse)
			data.GaugeData["MCacheSys"] = gauge(m.MCacheSys)
			data.GaugeData["MSpanInuse"] = gauge(m.MSpanInuse)
			data.GaugeData["MSpanSys"] = gauge(m.MSpanSys)
			data.GaugeData["Mallocs"] = gauge(m.Mallocs)
			data.GaugeData["NextGC"] = gauge(m.NextGC)
			data.GaugeData["NumForcedGC"] = gauge(m.NumForcedGC)
			data.GaugeData["NumGC"] = gauge(m.NumGC)
			data.GaugeData["OtherSys"] = gauge(m.OtherSys)
			data.GaugeData["PauseTotalNs"] = gauge(m.PauseTotalNs)
			data.GaugeData["StackInuse"] = gauge(m.StackInuse)
			data.GaugeData["StackSys"] = gauge(m.StackSys)
			data.GaugeData["Sys"] = gauge(m.Sys)
			data.GaugeData["TotalAlloc"] = gauge(m.TotalAlloc)

			data.Iterator++
			data.CounterData["PollCount"] = counter(data.Iterator)
			data.CounterData["RandomValue"] = counter(r1.Int())

			fmt.Println("poll data", data.CounterData["PollCount"])
			data.Unlock()

		case <-ctx.Done():
			fmt.Println("break poll")
			exitChan <- 1
			return
		}
	}
}

func report(ctx context.Context, ticker *time.Ticker) {
	fmt.Println("run report event spy")
	for {
		select {
		case <-ticker.C:
			data.Lock()
			for name, value := range data.GaugeData {
				response, err := http.Post(fmt.Sprintf(
					reportUrl,
					defaultAddress,
					defaultPort,
					"gauge",
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
			for name, value := range data.CounterData {
				fmt.Println(fmt.Sprintf(
					reportUrl,
					defaultAddress,
					defaultPort,
					"counter",
					name,
					value))
				response, err := http.Post(fmt.Sprintf(
					reportUrl,
					defaultAddress,
					defaultPort,
					"counter",
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
			fmt.Println("report data", data.CounterData["PollCount"])
			data.Unlock()
		case <-ctx.Done():
			fmt.Println("break report")
			exitChan <- 1
			return
		}
	}
}
