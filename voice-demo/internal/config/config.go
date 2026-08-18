package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config 服务运行所需的全部配置。
type Config struct {
	APIKey string
	URL    string
	Model  string
	Voice  string
	Port   string
	WebDir string
}

// Load 从 .env 文件与系统环境变量读取配置。
// 系统环境变量优先于 .env 文件；缺失字段使用默认值。
func Load(envPath string) (Config, error) {
	cfg := Config{
		URL:    "wss://api.stepfun.com/v1/realtime",
		Model:  "stepaudio-2.5-realtime",
		Voice:  "linjiajiejie",
		Port:   "8080",
		WebDir: "web",
	}

	values, err := readEnvFile(envPath)
	if err != nil {
		return Config{}, err
	}

	// 系统环境变量覆盖 .env 文件。
	for _, key := range []string{
		"STEPFUN_API_KEY",
		"STEPFUN_URL",
		"MODEL",
		"VOICE",
		"PORT",
		"WEB_DIR",
	} {
		if v := os.Getenv(key); v != "" {
			values[key] = v
		}
	}

	cfg.APIKey = values["STEPFUN_API_KEY"]
	cfg.URL = firstNonEmpty(values["STEPFUN_URL"], cfg.URL)
	cfg.Model = firstNonEmpty(values["MODEL"], cfg.Model)
	cfg.Voice = firstNonEmpty(values["VOICE"], cfg.Voice)
	cfg.Port = firstNonEmpty(values["PORT"], cfg.Port)
	cfg.WebDir = firstNonEmpty(values["WEB_DIR"], cfg.WebDir)

	return cfg, nil
}

// readEnvFile 解析 KEY=VALUE 格式的 .env 文件，忽略空行与 # 注释。
func readEnvFile(path string) (map[string]string, error) {
	values := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return values, nil
		}
		return nil, fmt.Errorf("打开配置文件失败: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return values, scanner.Err()
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
