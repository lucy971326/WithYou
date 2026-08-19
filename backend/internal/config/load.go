package config

import (
	"bufio"
	"os"
	"strings"
)

// Load 读环境变量，缺省补齐。可选加载当前目录或上级目录的 .env，不覆盖已有环境变量。
func Load() (Config, error) {
	loadDotEnv(".env")
	loadDotEnv("../.env")

	cfg := Config{
		Addr:              envOr("WITHYOU_ADDR", "127.0.0.1:8080"),
		QwenAPIKey:        firstEnv("QWEN_API_KEY", "DASHSCOPE_API_KEY"),
		QwenPlotModel:     envOr("QWEN_PLOT_MODEL", DefaultQwenPlotModel),
		QwenRealtimeModel: envOr("QWEN_REALTIME_MODEL", DefaultQwenRealtimeModel),
		QwenRealtimeURL:   envOr("QWEN_REALTIME_URL", "wss://dashscope-intl.aliyuncs.com/api-ws/v1/realtime"),
	}
	return cfg, nil
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
