package library

import (
	"net/http"
	"os"
	"path/filepath"
)

// Media 按 HTTP Range 读当前可播文件。解码在浏览器。
type Media struct {
	state *state
}

// Serve 把可播文件交给 http.ServeContent（自带 Range / HEAD）。
func (m *Media) Serve(w http.ResponseWriter, r *http.Request) {
	file, ok := m.state.current()
	if !ok {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "no file open"})
		return
	}
	src := file.PlayablePath
	if src == "" {
		src = file.Path
	}

	f, err := os.Open(src)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "open media"})
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "stat media"})
		return
	}

	http.ServeContent(w, r, filepath.Base(src), info.ModTime(), f)
}

// Current 返回正在播的那部，没有则 ok=false。
func (m *Media) Current() (OpenedFile, bool) {
	return m.state.current()
}
