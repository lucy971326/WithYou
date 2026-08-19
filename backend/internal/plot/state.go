package plot

import "sync"

type state struct {
	mu  sync.RWMutex
	doc SubtitleDoc
	has bool
}

func (s *state) current() (SubtitleDoc, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.doc, s.has
}

func (s *state) replace(doc SubtitleDoc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.doc = doc
	s.has = true
}
