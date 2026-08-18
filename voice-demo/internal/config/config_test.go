package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "STEPFUN_API_KEY=file-key\nMODEL=step-1o-audio\nPORT=9090\n"
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("写 .env 失败: %v", err)
	}
	t.Setenv("STEPFUN_API_KEY", "")

	cfg, err := Load(envPath)
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}
	if cfg.APIKey != "file-key" {
		t.Errorf("APIKey 错误: %q", cfg.APIKey)
	}
	if cfg.Model != "step-1o-audio" {
		t.Errorf("Model 错误: %q", cfg.Model)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port 错误: %q", cfg.Port)
	}
	if cfg.Voice != "linjiajiejie" {
		t.Errorf("默认音色错误: %q", cfg.Voice)
	}
}

func TestLoadOsEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("PORT=7000\n"), 0o644); err != nil {
		t.Fatalf("写 .env 失败: %v", err)
	}
	t.Setenv("PORT", "8000")

	cfg, err := Load(envPath)
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}
	if cfg.Port != "8000" {
		t.Errorf("系统环境变量应覆盖 .env，期望 8000，实际 %q", cfg.Port)
	}
}

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.env"))
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}
	if cfg.URL != "wss://api.stepfun.com/v1/realtime" {
		t.Errorf("默认 URL 错误: %q", cfg.URL)
	}
	if cfg.Model != "stepaudio-2.5-realtime" {
		t.Errorf("默认模型错误: %q", cfg.Model)
	}
}
