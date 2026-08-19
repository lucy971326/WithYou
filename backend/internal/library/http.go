package library

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// HTTP 把 library 能力挂到标准库 ServeMux 上。
type HTTP struct {
	picker   *Picker
	playable *Playable
	media    *Media
}

// Register 注册打开、当前文件、媒体 Range。
func (h *HTTP) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", h.handleHealth)
	mux.HandleFunc("POST /api/open", h.handleOpen)
	mux.HandleFunc("GET /api/current", h.handleCurrent)
	mux.HandleFunc("GET /media", h.media.Serve)
	mux.HandleFunc("HEAD /media", h.media.Serve)
}

func (h *HTTP) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HTTP) handleOpen(w http.ResponseWriter, r *http.Request) {
	file, err := h.picker.Pick(r.Context())
	if err != nil {
		if errors.Is(err, ErrCancelled) {
			writeJSON(w, http.StatusConflict, ErrorResponse{Error: "cancelled"})
			return
		}
		log.Printf("library open pick: %v", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "open failed"})
		return
	}
	file, err = h.playable.Ensure(r.Context(), file)
	if err != nil {
		log.Printf("library open prepare: %v", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "open failed"})
		return
	}
	writeJSON(w, http.StatusOK, toOpenResponse(file))
}

func (h *HTTP) handleCurrent(w http.ResponseWriter, _ *http.Request) {
	file, ok := h.media.Current()
	if !ok {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "no file open"})
		return
	}
	writeJSON(w, http.StatusOK, CurrentResponse(toOpenResponse(file)))
}

func toOpenResponse(file OpenedFile) OpenResponse {
	return OpenResponse{
		Name:        file.Name,
		Size:        file.Size,
		Codec:       file.Codec,
		PixelFormat: file.PixelFormat,
		BrowserSafe: file.BrowserSafe,
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
