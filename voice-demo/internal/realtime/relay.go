package realtime

import (
	"context"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// RelayConfig 中转能力所需的配置与依赖。
type RelayConfig struct {
	UpstreamURL string
	APIKey      string
	Logger      *log.Logger
}

// Relay 负责浏览器连接与官方连接之间的双向透传。
type Relay struct {
	client   *websocket.Conn // 浏览器侧连接
	upstream *UpstreamClient // 官方侧连接
	config   RelayConfig
}

// NewRelay 创建中转对象。
// client 是已经完成升级的浏览器 WebSocket 连接。
func NewRelay(client *websocket.Conn, config RelayConfig) *Relay {
	return &Relay{
		client:   client,
		upstream: NewUpstreamClient(),
		config:   config,
	}
}

// Run 建立官方连接并启动双向转发，直到任一端断开。
func (r *Relay) Run(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	if err := r.upstream.Connect(ctx, r.config.UpstreamURL, r.config.APIKey); err != nil {
		return err
	}
	r.config.Logger.Println("已连接官方 Realtime 服务")

	results := make(chan error, 2)
	go func() { results <- r.forwardClientToUpstream() }()
	go func() { results <- r.forwardUpstreamToClient() }()

	// 任一方先结束，就收尾关闭两端。
	firstErr := <-results
	cancel()
	r.shutdown()
	<-results
	return firstErr
}

// forwardClientToUpstream 把浏览器发来的消息原样转发给官方。
func (r *Relay) forwardClientToUpstream() error {
	for {
		_, data, err := r.client.ReadMessage()
		if err != nil {
			return fmt.Errorf("读取浏览器消息失败: %w", err)
		}
		if err := r.upstream.SendMessage(data); err != nil {
			return err
		}
	}
}

// forwardUpstreamToClient 把官方发来的消息原样转发给浏览器。
func (r *Relay) forwardUpstreamToClient() error {
	for {
		data, err := r.upstream.ReadMessage()
		if err != nil {
			return fmt.Errorf("读取官方消息失败: %w", err)
		}
		if err := r.client.WriteMessage(websocket.TextMessage, data); err != nil {
			return fmt.Errorf("向浏览器写入失败: %w", err)
		}
	}
}

// shutdown 关闭两端连接，让另一个转发 goroutine 结束。
func (r *Relay) shutdown() {
	if r.client != nil {
		_ = r.client.Close()
	}
	if r.upstream != nil {
		_ = r.upstream.Close()
	}
}
