package main

import (
	"log"
	"net/http"

	"withyou/internal/config"
	"withyou/internal/library"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	lib := library.New()
	mux := http.NewServeMux()
	lib.HTTP.Register(mux)

	log.Printf("withyou listening on http://%s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Fatal(err)
	}
}
