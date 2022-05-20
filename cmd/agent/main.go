package main

import (
	"flag"
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

func main() {
	addressFlag := flag.String("a", "127.0.0.1:8080", "server address")
	reportIntervalIntervalFlag := flag.Duration("r", time.Second*10, "agent report interval")
	pollIntervalIntervalFlag := flag.Duration("p", time.Second*10, "agent poll interval")
	flag.Parse()

	reportInterval := params.DurationByEnv(*reportIntervalIntervalFlag, "REPORT_INTERVAL")
	pollInterval := params.DurationByEnv(*pollIntervalIntervalFlag, "POLL_INTERVAL")
	address := params.StrByEnv(*addressFlag, "ADDRESS")

	fmt.Println(fmt.Sprintf(
		"Start agent with intervals for poll: %v, for report: %v and api address: %s \n",
		pollInterval,
		reportInterval,
		address))
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

	func() {
		s := <-signalChanel
		switch s {
		case syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM:
			fmt.Println("Signal quit triggered.")
			pollerService.Stop(cancelPoller)
			reporterService.Stop(cancelReporter)
			os.Exit(0)
		default:
			fmt.Println("Unknown signal.")
		}
	}()
}
