package transcript

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"withyou/internal/reporoot"
)

// Entry 是对话记录里的一行，最终以 JSONL 落盘。
type Entry struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"` // user / ai / event
	Text   string `json:"text,omitempty"`
	Detail string `json:"detail,omitempty"`
	Status string `json:"status,omitempty"` // done / interrupted
	At     int64  `json:"at"`
}

// maxEntries 单媒体最多保留多少条，超出在 Load 时截断。
const maxEntries = 1000

// Store 把每个媒体的对话记录写成一个 JSONL 文件。
type Store struct {
	dir string
	mu  sync.Mutex
}

// newStore 建 Store，目录在仓库根 cache/transcript。
func newStore() *Store {
	dir := filepath.Join(reporoot.Root(), "cache", "transcript")
	_ = os.MkdirAll(dir, 0o755)
	log.Printf("transcript store: %s", dir)
	return &Store{dir: dir}
}

// Load 读当前媒体的全部记录；不存在返回空切片。
func (s *Store) Load(path string, size int64) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.file(path, size))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var out []Entry
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			log.Printf("transcript: skip bad line: %v", err)
			continue
		}
		out = append(out, e)
	}
	if len(out) > maxEntries {
		out = out[len(out)-maxEntries:]
		if err := s.rewriteLocked(s.file(path, size), out); err != nil {
			log.Printf("transcript: trim rewrite: %v", err)
		}
	}
	return out, nil
}

// Append 追加一行记录。
func (s *Store) Append(path string, size int64, e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.file(path, size), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(line, '\n'))
	return err
}

// Clear 删掉当前媒体的记录文件，不存在也视为成功。
func (s *Store) Clear(path string, size int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.Remove(s.file(path, size))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *Store) rewriteLocked(file string, entries []Entry) error {
	var b strings.Builder
	for _, e := range entries {
		line, err := json.Marshal(e)
		if err != nil {
			return err
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return os.WriteFile(file, []byte(b.String()), 0o644)
}

func (s *Store) file(path string, size int64) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "%s|%d", path, size))
	return filepath.Join(s.dir, hex.EncodeToString(sum[:])+".jsonl")
}
