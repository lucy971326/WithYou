package server

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func serveTestServer(t *testing.T, upstreamURL string) *httptest.Server {
	t.Helper()

	webDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(webDir, "index.html"), []byte("voice-demo index"), 0o644); err != nil {
		t.Fatalf("写 index.html 失败: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(webDir, "assets"), 0o755); err != nil {
		t.Fatalf("建 assets 目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(webDir, "assets", "app.js"), []byte("console.log('app')"), 0o644); err != nil {
		t.Fatalf("写 app.js 失败: %v", err)
	}

	app := New(Dependencies{
		UpstreamURL: upstreamURL,
		APIKey:      "server-key",
		WebDir:      webDir,
		Logger:      log.New(io.Discard, "", 0),
	})
	return httptest.NewServer(app.Handler())
}

func mockUpstreamEcho(t *testing.T) *httptest.Server {
	t.Helper()

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			_ = conn.WriteMessage(websocket.TextMessage, data)
		}
	}))
	return server
}

func wsURL(server *httptest.Server, path string) string {
	return "ws" + strings.TrimPrefix(server.URL+path, "http")
}

func TestServeIndexAndAssets(t *testing.T) {
	upstream := mockUpstreamEcho(t)
	defer upstream.Close()

	app := serveTestServer(t, wsURL(upstream, ""))
	defer app.Close()

	resp, err := http.Get(app.URL + "/")
	if err != nil {
		t.Fatalf("GET / 失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "voice-demo index") {
		t.Errorf("页面内容不符: %s", body)
	}

	resp2, err := http.Get(app.URL + "/assets/app.js")
	if err != nil {
		t.Fatalf("GET /assets/app.js 失败: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("app.js 状态码错误: %d", resp2.StatusCode)
	}
}

func TestServeWebSocketRelay(t *testing.T) {
	upstream := mockUpstreamEcho(t)
	defer upstream.Close()

	app := serveTestServer(t, wsURL(upstream, ""))
	defer app.Close()

	dialer := websocket.Dialer{HandshakeTimeout: 3 * time.Second}
	conn, _, err := dialer.Dial(wsURL(app, "/ws"), nil)
	if err != nil {
		t.Fatalf("连接 /ws 失败: %v", err)
	}
	defer conn.Close()

	payload := `{"type":"session.update","session":{}}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("读取回显失败: %v", err)
	}
	if string(data) != payload {
		t.Errorf("期望回显 %q，实际 %q", payload, string(data))
	}
}
