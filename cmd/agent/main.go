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

const (
	pollInterval   = time.Second * 2
	reportInterval = time.Second * 10
)

var (
	exitChan chan int
)

func main() {
	fmt.Println("Start agent")
	exitChan = make(chan int)
	storage := ramstorage.New()
	pollerService := poller.New(pollInterval, storage)
	reporterService := reporter.New(reportInterval, storage)

	cancelPoller := pollerService.Start()
	cancelReporter := reporterService.Start()

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
			pollerService.Stop(cancelPoller)
			reporterService.Stop(cancelReporter)
			exitChan <- 1
		default:
			fmt.Println("Unknown signal.")
		}
	}()
	<-exitChan
}
