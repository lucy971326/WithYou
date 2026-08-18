package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"voice-demo/internal/config"
	"voice-demo/internal/server"
)

func main() {
	logger := log.New(os.Stdout, "[voice-demo] ", log.LstdFlags)

	cfg, err := config.Load(".env")
	if err != nil {
		logger.Fatalf("加载配置失败: %v", err)
	}
	if cfg.APIKey == "" {
		logger.Fatal("缺少 STEPFUN_API_KEY，请在 .env 中配置")
	}

	app := server.New(server.Dependencies{
		UpstreamURL: cfg.URL + "?model=" + cfg.Model,
		APIKey:      cfg.APIKey,
		WebDir:      cfg.WebDir,
		Logger:      logger,
	})

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: app.Handler(),
	}

	go func() {
		logger.Printf("服务已启动: http://localhost:%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("服务启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Println("服务关闭中...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}
