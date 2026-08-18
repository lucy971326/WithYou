package realtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// mockUpstream 是一个假的官方服务：收到 session.update 后回一个 session.created。
func mockUpstream(t *testing.T) *httptest.Server {
	t.Helper()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("升级连接失败: %v", err)
			return
		}
		defer conn.Close()

		// 校验鉴权头存在。
		if got := r.Header.Get("Authorization"); got == "" {
			t.Error("请求头缺少 Authorization")
		}

		// 读一条客户端事件。
		var raw map[string]any
		if err := conn.ReadJSON(&raw); err != nil {
			t.Errorf("读取客户端事件失败: %v", err)
			return
		}
		if raw["type"] != string(EventSessionUpdate) {
			t.Errorf("期望 session.update，实际 %v", raw["type"])
		}

		// 回复 session.created。
		reply := `{"type":"session.created","session":{"model":"stepaudio-2.5-realtime"}}`
		if err := conn.WriteMessage(websocket.TextMessage, []byte(reply)); err != nil {
			t.Errorf("回复 session.created 失败: %v", err)
		}
	}))
	return server
}

func wsAddr(server *httptest.Server) string {
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func TestUpstreamClientConnectAndExchange(t *testing.T) {
	server := mockUpstream(t)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := NewUpstreamClient()
	defer client.Close()

	if err := client.Connect(ctx, wsAddr(server), "fake-key"); err != nil {
		t.Fatalf("Connect 失败: %v", err)
	}

	session := SessionUpdateEvent{
		Type: EventSessionUpdate,
		Session: SessionConfig{
			Modalities: []string{"text", "audio"},
		},
	}
	if err := client.SendJSON(session); err != nil {
		t.Fatalf("SendJSON 失败: %v", err)
	}

	data, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage 失败: %v", err)
	}
	event, err := ParseServerEvent(data)
	if err != nil {
		t.Fatalf("解析服务端事件失败: %v", err)
	}
	if _, ok := event.(SessionCreatedEvent); !ok {
		t.Fatalf("期望 SessionCreatedEvent，实际 %T", event)
	}
}

func TestUpstreamClientSendBeforeConnect(t *testing.T) {
	client := NewUpstreamClient()
	if err := client.SendJSON(SessionUpdateEvent{Type: EventSessionUpdate}); err == nil {
		t.Fatal("未连接时发送应返回错误")
	}
	if err := client.Close(); err != nil {
		t.Fatalf("关闭未连接客户端应无错误，实际 %v", err)
	}
}
