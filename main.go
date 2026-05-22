package main

import (
	_ "embed"
	"log"
	"net/http"
)

//go:embed index.html
var indexHTML []byte

var globalHub = NewHub()

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", serveIndex)
	mux.HandleFunc("GET /stream", globalHub.ServeStream)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}
