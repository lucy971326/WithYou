package config

import "testing"

func TestLoadSelectsDomesticQwenEndpoints(t *testing.T) {
	t.Setenv("QWEN_SITE", "cn")
	t.Setenv("QWEN_PLOT_BASE_URL_CN", "https://plot.cn.example/v1")
	t.Setenv("QWEN_REALTIME_URL_CN", "wss://realtime.cn.example/ws")
	t.Setenv("QWEN_VOICE_API_URL_CN", "https://voice.cn.example/api")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.QwenSite != "cn" {
		t.Fatalf("site = %q, want cn", cfg.QwenSite)
	}
	if cfg.QwenPlotBaseURL != "https://plot.cn.example/v1" {
		t.Fatalf("plot base URL = %q", cfg.QwenPlotBaseURL)
	}
	if cfg.QwenRealtimeURL != "wss://realtime.cn.example/ws" {
		t.Fatalf("realtime URL = %q", cfg.QwenRealtimeURL)
	}
	if cfg.QwenVoiceAPIURL != "https://voice.cn.example/api" {
		t.Fatalf("voice API URL = %q", cfg.QwenVoiceAPIURL)
	}
}

func TestLoadRejectsUnknownQwenSite(t *testing.T) {
	t.Setenv("QWEN_SITE", "mars")
	if _, err := Load(); err == nil {
		t.Fatal("Load should reject an unknown QWEN_SITE")
	}
}
