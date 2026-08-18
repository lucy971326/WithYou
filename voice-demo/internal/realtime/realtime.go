package realtime

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// UpstreamClient 是与官方 Realtime 服务通信的能力对象。
// 它负责建立带鉴权的连接，以及发送客户端事件、读取服务端消息。
type UpstreamClient struct {
	conn    *websocket.Conn
	writeMu sync.Mutex
}

// NewUpstreamClient 创建上游客户端。
func NewUpstreamClient() *UpstreamClient {
	return &UpstreamClient{}
}

// Connect 建立到官方服务的 WebSocket 连接，并在握手时带上 API Key。
func (c *UpstreamClient) Connect(ctx context.Context, url, apiKey string) error {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+apiKey)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, header)
	if err != nil {
		return fmt.Errorf("连接官方服务失败: %w", err)
	}
	c.conn = conn
	return nil
}

// SendJSON 把任意客户端事件序列化为 JSON 并发送。
func (c *UpstreamClient) SendJSON(v any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("尚未连接官方服务")
	}
	if err := c.conn.WriteJSON(v); err != nil {
		return fmt.Errorf("发送事件失败: %w", err)
	}
	return nil
}

// SendMessage 发送一条原始文本消息（JSON 字节）。
func (c *UpstreamClient) SendMessage(data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("尚未连接官方服务")
	}
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}
	return nil
}

// ReadMessage 读取一条服务端消息的原始字节。
func (c *UpstreamClient) ReadMessage() ([]byte, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("尚未连接官方服务")
	}
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Close 关闭与官方服务的连接。
func (c *UpstreamClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
