package main

import (
	"fmt"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/services/poller"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/services/reporter"
	ramstorage "github.com/resssoft/go-metrics-ya-praktikum/internal/storages/ram"
	"github.com/resssoft/go-metrics-ya-praktikum/pkg/params"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	pollInterval   = time.Second * 2
	reportInterval = time.Second * 10
	address        = "127.0.0.1:8080"
)

var (
	exitChan chan int
)

func main() {
	reportInterval = params.DurationByEnv(reportInterval, "REPORT_INTERVAL")
	pollInterval = params.DurationByEnv(pollInterval, "POLL_INTERVAL")
	address = params.StrByEnv(address, "ADDRESS")

	fmt.Println(fmt.Sprintf(
		"Start agent with intervals for poll: %v, for report: %v and api address: %s",
		pollInterval,
		reportInterval,
		address))
	exitChan = make(chan int)
	storage := ramstorage.New()
	pollerService := poller.New(pollInterval, storage)
	reporterService := reporter.New(reportInterval, address, storage)

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
