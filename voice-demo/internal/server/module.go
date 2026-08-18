package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"voice-demo/internal/realtime"
)

// Dependencies 是 HTTP 服务需要的显式依赖。
type Dependencies struct {
	UpstreamURL string
	APIKey      string
	WebDir      string
	Logger      *log.Logger
}

// Server 是 HTTP 组合器：负责静态页面托管与 WebSocket 中转入口。
type Server struct {
	deps     Dependencies
	mux      *http.ServeMux
	upgrader websocket.Upgrader
}

// New 创建 HTTP 服务并注册路由。
func New(deps Dependencies) *Server {
	s := &Server{
		deps: deps,
		mux:  http.NewServeMux(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
	s.mux.HandleFunc("/ws", s.handleWebSocket)
	s.mux.Handle("/", http.FileServer(http.Dir(deps.WebDir)))
	return s
}

// Handler 返回可供 http.Server 使用的根处理器。
func (s *Server) Handler() http.Handler {
	return s.mux
}

// handleWebSocket 升级浏览器连接，并用 Relay 桥接官方服务。
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.deps.Logger.Printf("升级 WebSocket 失败: %v", err)
		return
	}

	relay := realtime.NewRelay(conn, realtime.RelayConfig{
		UpstreamURL: s.deps.UpstreamURL,
		APIKey:      s.deps.APIKey,
		Logger:      s.deps.Logger,
	})
	if err := relay.Run(r.Context()); err != nil {
		s.deps.Logger.Printf("中转结束: %v", err)
	}
}
