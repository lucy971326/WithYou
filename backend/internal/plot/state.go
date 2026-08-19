package plot

import "sync"

type state struct {
	mu      sync.RWMutex
	doc     SubtitleDoc
	has     bool
	plot    PlotDoc
	hasPlot bool
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

func (s *state) currentPlot() (PlotDoc, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.plot, s.hasPlot
}

func (s *state) replacePlot(doc PlotDoc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.plot = doc
	s.hasPlot = true
}
