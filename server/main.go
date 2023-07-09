package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"realm/handler"
)

//go:embed assets/*
var assets embed.FS

func main() {

	assetsRFS, _ := fs.Sub(assets, "assets")
	var assetsFS = http.FS(assetsRFS)

	fs := http.FileServer(assetsFS)
	log.Print("Serving on :8888")

	mux := http.NewServeMux()

	mux.HandleFunc("/realm/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(w, r)
	})

	mux.HandleFunc("/ws", handler.Websocket)
	mux.HandleFunc("/forum/", handler.Forum)

	s := &http.Server{
		Handler:        mux,
		Addr:           ":8888",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
