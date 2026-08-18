package realtime

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// mockEchoUpstream 假官方服务：校验鉴权头，并把收到的每条消息原样回显。
func mockEchoUpstream(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer relay-key" {
			t.Errorf("期望 Authorization=Bearer relay-key，实际 %q", got)
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("升级上游连接失败: %v", err)
			return
		}
		defer conn.Close()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}))
	return server
}

func TestRelayRoundTrip(t *testing.T) {
	upstream := mockEchoUpstream(t)
	defer upstream.Close()

	// 浏览器侧中转服务器：升级后用 Relay 桥接到 mock 上游。
	browserServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("升级浏览器连接失败: %v", err)
			return
		}
		relay := NewRelay(conn, RelayConfig{
			UpstreamURL: wsAddr(upstream),
			APIKey:      "relay-key",
			Logger:      log.New(io.Discard, "", 0),
		})
		go func() { _ = relay.Run(context.Background()) }()
	}))
	defer browserServer.Close()

	dialer := websocket.Dialer{HandshakeTimeout: 3 * time.Second}
	browser, _, err := dialer.Dial(wsAddr(browserServer), nil)
	if err != nil {
		t.Fatalf("浏览器连接失败: %v", err)
	}
	defer browser.Close()

	payload := `{"type":"response.create"}`
	if err := browser.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	_, data, err := browser.ReadMessage()
	if err != nil {
		t.Fatalf("读取回显失败: %v", err)
	}
	if string(data) != payload {
		t.Errorf("期望回显 %q，实际 %q", payload, string(data))
	}
}

func TestRelayEndsWhenBrowserCloses(t *testing.T) {
	upstream := mockEchoUpstream(t)
	defer upstream.Close()

	type done struct{}
	relayDone := make(chan done, 1)

	browserServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("升级浏览器连接失败: %v", err)
			return
		}
		relay := NewRelay(conn, RelayConfig{
			UpstreamURL: wsAddr(upstream),
			APIKey:      "relay-key",
			Logger:      log.New(io.Discard, "", 0),
		})
		go func() {
			_ = relay.Run(context.Background())
			relayDone <- done{}
		}()
	}))
	defer browserServer.Close()

	dialer := websocket.Dialer{HandshakeTimeout: 3 * time.Second}
	browser, _, err := dialer.Dial(wsAddr(browserServer), nil)
	if err != nil {
		t.Fatalf("浏览器连接失败: %v", err)
	}

	browser.Close()

	select {
	case <-relayDone:
	case <-time.After(3 * time.Second):
		t.Fatal("浏览器关闭后 Relay 未按预期结束")
	}
}
