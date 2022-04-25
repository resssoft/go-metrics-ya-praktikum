package main

import (
	"fmt"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/server"
	ramstorage "github.com/resssoft/go-metrics-ya-praktikum/internal/storages/ram"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Start server")
	storage := ramstorage.New()
	log.Fatal(http.ListenAndServe(":8080", server.Router(storage)))
}
