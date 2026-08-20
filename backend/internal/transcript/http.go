package transcript

import (
	"encoding/json"
	"log"
	"net/http"

	"withyou/internal/library"
)

// HTTP 把对话记录的读写挂到 ServeMux 上。
type HTTP struct {
	media *library.Media
	store *Store
}

type getResponse struct {
	Name    string  `json:"name"`
	Entries []Entry `json:"entries"`
}

type appendRequest struct {
	Entry Entry `json:"entry"`
}

type appendResponse struct {
	Appended bool `json:"appended"`
}

type clearResponse struct {
	Cleared bool `json:"cleared"`
}

// Register 注册三个接口：读、追加、清空。
func (h *HTTP) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/transcript", h.handleGet)
	mux.HandleFunc("POST /api/transcript", h.handleAppend)
	mux.HandleFunc("DELETE /api/transcript", h.handleClear)
}

func (h *HTTP) handleGet(w http.ResponseWriter, _ *http.Request) {
	file, ok := h.media.Current()
	if !ok {
		writeJSON(w, http.StatusConflict, library.ErrorResponse{Error: "no file open"})
		return
	}
	entries, err := h.store.Load(file.Path, file.Size)
	if err != nil {
		log.Printf("transcript load: %v", err)
		writeJSON(w, http.StatusInternalServerError, library.ErrorResponse{Error: "load transcript"})
		return
	}
	if entries == nil {
		entries = []Entry{}
	}
	writeJSON(w, http.StatusOK, getResponse{Name: file.Name, Entries: entries})
}

func (h *HTTP) handleAppend(w http.ResponseWriter, r *http.Request) {
	file, ok := h.media.Current()
	if !ok {
		writeJSON(w, http.StatusConflict, library.ErrorResponse{Error: "no file open"})
		return
	}
	var req appendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Entry.ID == "" || req.Entry.Kind == "" {
		writeJSON(w, http.StatusBadRequest, library.ErrorResponse{Error: "bad entry"})
		return
	}
	if err := h.store.Append(file.Path, file.Size, req.Entry); err != nil {
		log.Printf("transcript append: %v", err)
		writeJSON(w, http.StatusInternalServerError, library.ErrorResponse{Error: "append transcript"})
		return
	}
	writeJSON(w, http.StatusOK, appendResponse{Appended: true})
}

func (h *HTTP) handleClear(w http.ResponseWriter, _ *http.Request) {
	file, ok := h.media.Current()
	if !ok {
		writeJSON(w, http.StatusConflict, library.ErrorResponse{Error: "no file open"})
		return
	}
	if err := h.store.Clear(file.Path, file.Size); err != nil {
		log.Printf("transcript clear: %v", err)
		writeJSON(w, http.StatusInternalServerError, library.ErrorResponse{Error: "clear transcript"})
		return
	}
	writeJSON(w, http.StatusOK, clearResponse{Cleared: true})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
