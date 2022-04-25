package main

import (
	"github.com/resssoft/go-metrics-ya-praktikum/internal/handlers"
	ramstorage "github.com/resssoft/go-metrics-ya-praktikum/internal/storages/ram"
	"log"
	"net/http"
)

func main() {
	storage := ramstorage.New()
	handler := handlers.NewMetricsSaver(storage)
	http.HandleFunc("/update/", handler.SaveMetrics)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
