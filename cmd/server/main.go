package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	http.HandleFunc("/", save)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func save(rw http.ResponseWriter, req *http.Request) {
	uriParts := strings.Split(req.URL.Path, "/")
	if len(uriParts) < 5 {
		rw.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(rw, "unsupported method")
		return
	}
	dataType := uriParts[2]
	dataName := uriParts[3]
	dataValue := uriParts[4]
	rw.WriteHeader(http.StatusNoContent)
	fmt.Println("data type,name,value: ", dataType, dataName, dataValue)
}
