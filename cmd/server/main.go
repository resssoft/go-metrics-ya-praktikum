package main

import (
	"fmt"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/server"
	ramstorage "github.com/resssoft/go-metrics-ya-praktikum/internal/storages/ram"
	"github.com/resssoft/go-metrics-ya-praktikum/pkg/params"
	"log"
	"net/http"
)

func main() {
	address := params.StrByEnv(":8080", "ADDRESS")
	fmt.Println("Start server by address: " + address)
	storage := ramstorage.New()
	log.Fatal(http.ListenAndServe(address, server.Router(storage)))
}
