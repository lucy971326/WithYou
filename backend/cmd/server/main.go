package main

import (
	"log"
	"net/http"

	"withyou/internal/config"
	"withyou/internal/library"
	"withyou/internal/plot"
	"withyou/internal/realtime"
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
	rt := realtime.New(realtime.Dependencies{
		APIKey: cfg.QwenAPIKey,
		Model:  cfg.QwenRealtimeModel,
		URL:    cfg.QwenRealtimeURL,
	})
	mux := http.NewServeMux()
	lib.HTTP.Register(mux)
	plt.HTTP.Register(mux)
	rt.HTTP.Register(mux)
	if cfg.QwenAPIKey == "" {
		log.Printf("realtime: QWEN_API_KEY empty, /api/realtime will 503")
	}

	log.Printf("withyou listening on http://%s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Fatal(err)
	}
}
