package main

import (
	"flag"
	"fmt"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/server"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/services/writer"
	ramstorage "github.com/resssoft/go-metrics-ya-praktikum/internal/storages/ram"
	"github.com/resssoft/go-metrics-ya-praktikum/pkg/params"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func main() {
	restoreFlag := flag.Bool("r", true, "restore flag")
	addressFlag := flag.String("a", ":8080", "server address")
	storePathFlag := flag.String("f", "/tmp/devops-metrics-db.json", "server store file path")
	storeIntervalFlag := flag.Duration("i", time.Second*300, "server store interval")
	flag.Parse()

	address := params.StrByEnv(*addressFlag, "ADDRESS")
	storeInterval := params.DurationByEnv(*storeIntervalFlag, "STORE_INTERVAL")
	storePath := params.StrByEnv(*storePathFlag, "STORE_FILE")
	restore := params.BoolByEnv(*restoreFlag, "RESTORE")

	fmt.Printf(
		"Start server by address: %s store duration: %v restore flag: %v and store file: %s \n",
		address,
		storeInterval,
		restore,
		storePath)
	storage := ramstorage.New()
	writerService := writer.New(storeInterval, storePath, restore, storage)
	cansel := writerService.Start()

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)
	go func() {
		s := <-signalChanel
		fmt.Printf("New OS signal: %v \n", s)
		switch s {
		case syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM:
			fmt.Println("Signal quit triggered.")
			writerService.Stop(cansel)
			os.Exit(0)
		default:
			fmt.Println("Unknown signal.")
		}
	}()

	log.Fatal(http.ListenAndServe(address, server.Router(storage)))
}

func Test_main(t *testing.T) {
	t.Skip()
}
