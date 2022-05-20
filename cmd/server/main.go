package main

import (
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
	"time"
)

func main() {
	address := params.StrByEnv(":8080", "ADDRESS")
	storeInterval := params.DurationByEnv(time.Second*300, "STORE_INTERVAL")
	storePath := params.StrByEnv("/tmp/devops-metrics-db.json", "STORE_FILE")
	restoreFlag := params.BoolByEnv(true, "RESTORE")
	fmt.Println("Start server by address: " + address)
	storage := ramstorage.New()
	writerService := writer.New(storeInterval, storePath, restoreFlag, storage)
	cansel := writerService.Start()

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
			writerService.Stop(cansel)
			os.Exit(0)
		default:
			fmt.Println("Unknown signal.")
		}
	}()

	log.Fatal(http.ListenAndServe(address, server.Router(storage)))
}
