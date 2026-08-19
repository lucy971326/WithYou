package plot

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// HTTP 把 plot 抽轨能力挂到 ServeMux 上。
type HTTP struct {
	extractor *Extractor
	state     *state
}

// Register 注册抽字幕接口。
func (h *HTTP) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/plot/subtitles", h.handleExtract)
	mux.HandleFunc("GET /api/plot/subtitles", h.handleGet)
}

func (h *HTTP) handleExtract(w http.ResponseWriter, r *http.Request) {
	doc, err := h.extractor.Extract(r.Context())
	if err != nil {
		status := http.StatusInternalServerError
		msg := "extract failed"
		switch {
		case errors.Is(err, ErrNoFile):
			status = http.StatusConflict
			msg = "no file open"
		case errors.Is(err, ErrNoSubtitle):
			status = http.StatusUnprocessableEntity
			msg = "no soft subtitle track"
		case errors.Is(err, ErrImageSubtitle):
			status = http.StatusUnprocessableEntity
			msg = "image subtitle not supported"
		default:
			log.Printf("plot extract: %v", err)
		}
		writeJSON(w, status, ErrorResponse{Error: msg})
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (h *HTTP) handleGet(w http.ResponseWriter, _ *http.Request) {
	doc, ok := h.state.current()
	if !ok {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "no subtitles yet"})
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
