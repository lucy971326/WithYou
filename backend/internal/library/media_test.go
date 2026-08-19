package library

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestMediaRangeAndMissing(t *testing.T) {
	mod := New()
	mux := http.NewServeMux()
	mod.HTTP.Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/media")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("empty library GET /media status=%d", resp.StatusCode)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "clip.bin")
	payload := []byte("0123456789abcdef")
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	mod.Media.state.replace(OpenedFile{
		Path:         path,
		PlayablePath: path,
		Name:         "clip.bin",
		Size:         int64(len(payload)),
	})

	full, err := http.Get(srv.URL + "/media")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(full.Body)
	full.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if full.StatusCode != http.StatusOK {
		t.Fatalf("GET /media status=%d", full.StatusCode)
	}
	if string(body) != string(payload) {
		t.Fatalf("GET /media body=%q", body)
	}

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/media", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Range", "bytes=0-3")
	partial, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	piece, err := io.ReadAll(partial.Body)
	partial.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if partial.StatusCode != http.StatusPartialContent {
		t.Fatalf("Range GET status=%d", partial.StatusCode)
	}
	if string(piece) != "0123" {
		t.Fatalf("Range GET body=%q", piece)
	}
}

func TestHealth(t *testing.T) {
	mod := New()
	mux := http.NewServeMux()
	mod.HTTP.Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health status=%d", resp.StatusCode)
	}
}
