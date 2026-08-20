package transcript

import (
	"testing"
)

func TestAppendLoad(t *testing.T) {
	s := &Store{dir: t.TempDir()}
	path := "C:\\media\\ep01.mkv"
	const size = 1234

	if err := s.Append(path, size, Entry{ID: "1", Kind: "user", Text: "hi", Status: "done", At: 1}); err != nil {
		t.Fatalf("append 1: %v", err)
	}
	if err := s.Append(path, size, Entry{ID: "2", Kind: "ai", Text: "hello", Status: "done", At: 2}); err != nil {
		t.Fatalf("append 2: %v", err)
	}

	got, err := s.Load(path, size)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Text != "hi" || got[1].Text != "hello" {
		t.Fatalf("unexpected entries: %+v", got)
	}
}

func TestClear(t *testing.T) {
	s := &Store{dir: t.TempDir()}
	path := "C:\\media\\ep01.mkv"
	const size = 1234

	if err := s.Append(path, size, Entry{ID: "1", Kind: "event", Text: "x", At: 1}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := s.Clear(path, size); err != nil {
		t.Fatalf("clear: %v", err)
	}
	got, err := s.Load(path, size)
	if err != nil {
		t.Fatalf("load after clear: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestLoadTrimsToMax(t *testing.T) {
	s := &Store{dir: t.TempDir()}
	path := "C:\\media\\ep01.mkv"
	const size = 1234

	for i := 0; i < maxEntries+1; i++ {
		if err := s.Append(path, size, Entry{ID: "id", Kind: "event", Text: "x", At: int64(i)}); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	got, err := s.Load(path, size)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got) != maxEntries {
		t.Fatalf("len = %d, want %d", len(got), maxEntries)
	}
	if got[0].At != 1 {
		t.Fatalf("first At = %d, want 1 (oldest kept)", got[0].At)
	}
}
