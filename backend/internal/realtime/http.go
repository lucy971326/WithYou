package realtime

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
)

// HTTP 是门，不是会话。只做升级，对话交给 Relay。
//
//	浏览器 GET /api/realtime → 本机 WS → Relay.Serve
type HTTP struct {
	relay *Relay
}

// Register 挂升级入口。
func (h *HTTP) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/realtime", h.handleWS)
}

// handleWS：验 key → 升 WS（仅本机源）→ 阻塞在 Serve，直到这条会话死。
func (h *HTTP) handleWS(w http.ResponseWriter, r *http.Request) {
	if h.relay.apiKey == "" {
		log.Printf("realtime reject: missing QWEN_API_KEY")
		http.Error(w, `{"error":"missing QWEN_API_KEY"}`, http.StatusServiceUnavailable)
		return
	}
	log.Printf("realtime browser upgrade %s", r.RemoteAddr)
	browser, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"127.0.0.1:*", "localhost:*"},
	})
	if err != nil {
		log.Printf("realtime accept: %v", err)
		return
	}
	defer browser.Close(websocket.StatusNormalClosure, "done")

	err = h.relay.Serve(r.Context(), browser)
	if err != nil {
		log.Printf("realtime serve: %v", err)
	}
}
