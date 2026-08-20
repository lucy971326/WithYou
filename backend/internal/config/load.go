package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type qwenEndpoints struct {
	plotBaseURL string
	realtimeURL string
	voiceAPIURL string
}

// Load 读环境变量，缺省补齐。可选加载当前目录或上级目录的 .env，不覆盖已有环境变量。
func Load() (Config, error) {
	loadDotEnv(".env")
	loadDotEnv("../.env")
	site, endpoints, err := loadQwenEndpoints()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Addr:              envOr("WITHYOU_ADDR", "127.0.0.1:8080"),
		QwenSite:          site,
		QwenAPIKey:        firstEnv("QWEN_API_KEY", "DASHSCOPE_API_KEY"),
		QwenPlotModel:     envOr("QWEN_PLOT_MODEL", DefaultQwenPlotModel),
		QwenPlotBaseURL:   endpoints.plotBaseURL,
		QwenRealtimeModel: envOr("QWEN_REALTIME_MODEL", DefaultQwenRealtimeModel),
		QwenRealtimeURL:   endpoints.realtimeURL,
		QwenRealtimeVoice: envOr("QWEN_REALTIME_VOICE", DefaultQwenRealtimeVoice),
		QwenVoiceAPIURL:   endpoints.voiceAPIURL,
	}
	return cfg, nil
}

func loadQwenEndpoints() (string, qwenEndpoints, error) {
	site := strings.ToLower(strings.TrimSpace(envOr("QWEN_SITE", DefaultQwenSite)))
	switch site {
	case "intl", "international":
		return "intl", qwenEndpoints{
			plotBaseURL: envOr("QWEN_PLOT_BASE_URL_INTL", DefaultQwenPlotBaseURLIntl),
			realtimeURL: envOr("QWEN_REALTIME_URL_INTL", DefaultQwenRealtimeURLIntl),
			voiceAPIURL: envOr("QWEN_VOICE_API_URL_INTL", DefaultQwenVoiceAPIURLIntl),
		}, nil
	case "cn", "china", "domestic":
		return "cn", qwenEndpoints{
			plotBaseURL: envOr("QWEN_PLOT_BASE_URL_CN", DefaultQwenPlotBaseURLCN),
			realtimeURL: envOr("QWEN_REALTIME_URL_CN", DefaultQwenRealtimeURLCN),
			voiceAPIURL: envOr("QWEN_VOICE_API_URL_CN", DefaultQwenVoiceAPIURLCN),
		}, nil
	default:
		return "", qwenEndpoints{}, fmt.Errorf("config: QWEN_SITE must be intl or cn, got %q", site)
	}
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" {
			continue
		}
		if _, exists := os.LookupEnv(k); exists {
			continue
		}
		_ = os.Setenv(k, v)
	}
}
