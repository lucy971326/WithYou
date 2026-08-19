package plot

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// HTTP 把 plot 抽轨和富化挂到 ServeMux 上。
type HTTP struct {
	extractor *Extractor
	enricher  *Enricher
	state     *state
}

// Register 注册抽字幕和富化接口。
func (h *HTTP) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/plot/subtitles", h.handleExtract)
	mux.HandleFunc("GET /api/plot/subtitles", h.handleGet)
	mux.HandleFunc("POST /api/plot/enrich", h.handleEnrich)
	mux.HandleFunc("GET /api/plot/enrich", h.handleGetPlot)
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

func (h *HTTP) handleEnrich(w http.ResponseWriter, r *http.Request) {
	doc, cached, err := h.enricher.Enrich(r.Context())
	if err != nil {
		status := http.StatusInternalServerError
		msg := "enrich failed"
		switch {
		case errors.Is(err, ErrNoFile):
			status = http.StatusConflict
			msg = "no file open"
		case errors.Is(err, ErrNoCues):
			status = http.StatusConflict
			msg = "no subtitles yet"
		case errors.Is(err, ErrNoAPIKey):
			status = http.StatusServiceUnavailable
			msg = "missing DASHSCOPE_API_KEY"
		case errors.Is(err, ErrEmptyContent), errors.Is(err, ErrInvalidSchema):
			status = http.StatusBadGateway
			msg = "invalid json from model"
		default:
			log.Printf("plot enrich: %v", err)
		}
		writeJSON(w, status, ErrorResponse{Error: msg})
		return
	}
	writeJSON(w, http.StatusOK, summarize(doc, cached))
}

func (h *HTTP) handleGetPlot(w http.ResponseWriter, _ *http.Request) {
	doc, ok := h.state.currentPlot()
	if !ok {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "no plot yet"})
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
