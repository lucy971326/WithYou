package main

import (
	"log"
	"net/http"

	"withyou/internal/config"
	"withyou/internal/library"
	"withyou/internal/plot"
	"withyou/internal/realtime"
	"withyou/internal/transcript"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	lib := library.New()
	plt := plot.New(plot.Dependencies{
		Media:      lib.Media,
		QwenAPIKey: cfg.QwenAPIKey,
		QwenModel:  cfg.QwenPlotModel,
	})
	rt := realtime.New(realtime.Dependencies{
		APIKey:      cfg.QwenAPIKey,
		Model:       cfg.QwenRealtimeModel,
		URL:         cfg.QwenRealtimeURL,
		Voice:       cfg.QwenRealtimeVoice,
		VoiceAPIURL: cfg.QwenVoiceAPIURL,
	})
	tr := transcript.New(transcript.Dependencies{Media: lib.Media})
	mux := http.NewServeMux()
	lib.HTTP.Register(mux)
	plt.HTTP.Register(mux)
	rt.HTTP.Register(mux)
	tr.HTTP.Register(mux)
	if cfg.QwenAPIKey == "" {
		log.Printf("realtime: QWEN_API_KEY empty, /api/realtime will 503")
	}

	log.Printf("withyou listening on http://%s", cfg.Addr)
	err = http.ListenAndServe(cfg.Addr, mux)
	if err != nil {
		log.Fatal(err)
	}
}
