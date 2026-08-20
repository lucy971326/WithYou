package realtime

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/coder/websocket"
)

// HTTP 是门，不是会话。只做升级，对话交给 Relay。
//
//	浏览器 GET /api/realtime → 本机 WS → Relay.Serve
type HTTP struct {
	relay       *Relay
	voice       string
	voiceAPIURL string
}

// Register 挂 HTTP 入口。
func (h *HTTP) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/realtime", h.handleWS)
	mux.HandleFunc("GET /api/voices", h.handleVoices)
}

// handleVoices 返回可选项：官方预置音色 + 账号下的克隆音色。
func (h *HTTP) handleVoices(w http.ResponseWriter, r *http.Request) {
	resp := voicesResponse{
		DefaultVoice: h.voice,
		Preset:       presetVoices,
		Custom:       []Voice{},
	}
	custom, err := listClonedVoices(r.Context(), h.relay.apiKey, h.voiceAPIURL)
	if err != nil {
		resp.CustomError = err.Error()
	} else {
		resp.Custom = custom
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
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
