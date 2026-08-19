package realtime

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
)

// HTTP 把 Realtime WS 挂到 ServeMux。
type HTTP struct {
	relay *Relay
}

// Register 注册 /api/realtime。
func (h *HTTP) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/realtime", h.handleWS)
}

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

	if err := h.relay.Serve(r.Context(), browser); err != nil {
		log.Printf("realtime serve: %v", err)
	}
}
