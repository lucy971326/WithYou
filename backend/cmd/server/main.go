package main

import (
	"log"
	"net/http"

	"withyou/internal/config"
	"withyou/internal/library"
	"withyou/internal/plot"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	lib := library.New()
	plt := plot.New(plot.Dependencies{
		Media:          lib.Media,
		DeepSeekAPIKey: cfg.DeepSeekAPIKey,
		DeepSeekModel:  cfg.DeepSeekModel,
	})
	mux := http.NewServeMux()
	lib.HTTP.Register(mux)
	plt.HTTP.Register(mux)

	log.Printf("withyou listening on http://%s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Fatal(err)
	}
}
