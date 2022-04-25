package main

import (
	"fmt"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/services/poller"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/services/reporter"
	ramstorage "github.com/resssoft/go-metrics-ya-praktikum/internal/storages/ram"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	exitChan       chan int
	pollInterval   = time.Second * 2
	reportInterval = time.Second * 10
)

func main() {
	fmt.Println("Start agent")
	exitChan = make(chan int)
	storage := ramstorage.New()
	poller := poller.New(pollInterval, storage)
	reporter := reporter.New(reportInterval, storage)

	poller.Start()
	reporter.Start()

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)

	go func() {
		s := <-signalChanel
		switch s {
		case syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM:
			fmt.Println("Signal quit triggered.")
			poller.Stop()
			reporter.Stop()
			exitChan <- 1
		default:
			fmt.Println("Unknown signal.")
		}
	}()
	<-exitChan
}
