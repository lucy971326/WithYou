package library

import "sync"

type state struct {
	mu   sync.RWMutex
	file OpenedFile
	has  bool
}

func (s *state) current() (OpenedFile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.file, s.has
}

func (s *state) replace(file OpenedFile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.file = file
	s.has = true
}
