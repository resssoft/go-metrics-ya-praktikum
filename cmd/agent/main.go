package main

import (
	"flag"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/services/poller"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/services/reporter"
	ramstorage "github.com/resssoft/go-metrics-ya-praktikum/internal/storages/ram"
	"github.com/resssoft/go-metrics-ya-praktikum/pkg/params"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel) // DebugLevel | InfoLevel
	addressFlag := flag.String("a", "127.0.0.1:8080", "server address")
	reportIntervalIntervalFlag := flag.Duration("r", time.Second*5, "agent report interval")
	pollIntervalIntervalFlag := flag.Duration("p", time.Second*2, "agent poll interval")
	cryptoKeyFlag := flag.String("k", "", "crypto key")
	flag.Parse()

	reportInterval := params.DurationByEnv(*reportIntervalIntervalFlag, "REPORT_INTERVAL")
	pollInterval := params.DurationByEnv(*pollIntervalIntervalFlag, "POLL_INTERVAL")
	address := params.StrByEnv(*addressFlag, "ADDRESS")
	cryptoKey := params.StrByEnv(*cryptoKeyFlag, "KEY")

	log.Info().Msgf(
		"Start agent with intervals for poll: %v, for report: %v and api address: %s Key [%s]\n",
		pollInterval,
		reportInterval,
		address,
		cryptoKey)
	storage := ramstorage.New()
	pollerService := poller.New(pollInterval, storage)
	reporterService := reporter.New(reportInterval, address, storage, cryptoKey)

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
			log.Info().Msg("Signal quit triggered.")
			pollerService.Stop(cancelPoller)
			reporterService.Stop(cancelReporter)
			os.Exit(0)
		default:
			log.Info().Msg("Unknown signal.")
		}
	}()
}

func TestMain(t *testing.T) {
	t.Skip()
}
