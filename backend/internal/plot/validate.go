package plot

import (
	"fmt"
	"strings"
)

func validatePlot(doc PlotDoc, cues []Cue) error {
	if strings.TrimSpace(doc.Title) == "" {
		return fmt.Errorf("plot: missing title")
	}
	if strings.TrimSpace(doc.Overview.GrandSummary) == "" {
		return fmt.Errorf("plot: missing overview.grand_summary")
	}
	if len(doc.Overview.KeyCharacters) == 0 {
		return fmt.Errorf("plot: empty overview.key_characters")
	}
	if len(doc.Overview.KeyPlotPoints) == 0 {
		return fmt.Errorf("plot: empty overview.key_plot_points")
	}
	if len(doc.MajorSegments) == 0 {
		return fmt.Errorf("plot: empty major_segments")
	}

	lo, hi := cueSpan(cues)
	for i, maj := range doc.MajorSegments {
		if maj.StartSec >= maj.EndSec {
			return fmt.Errorf("plot: major[%d] bad time %d-%d", i, maj.StartSec, maj.EndSec)
		}
		if strings.TrimSpace(maj.Title) == "" || strings.TrimSpace(maj.Summary) == "" {
			return fmt.Errorf("plot: major[%d] empty title/summary", i)
		}
		if len(maj.SubSegments) == 0 {
			return fmt.Errorf("plot: major[%d] empty sub_segments", i)
		}
		for j, sub := range maj.SubSegments {
			if sub.StartSec >= sub.EndSec {
				return fmt.Errorf("plot: major[%d].sub[%d] bad time %d-%d", i, j, sub.StartSec, sub.EndSec)
			}
			err := requireNonEmpty(sub)
			if err != nil {
				return fmt.Errorf("plot: major[%d].sub[%d]: %w", i, j, err)
			}
		}
	}

	if len(cues) == 0 {
		return nil
	}
	first := doc.MajorSegments[0].StartSec
	last := doc.MajorSegments[len(doc.MajorSegments)-1].EndSec
	const slack = 30.0
	if float64(first) > lo+slack {
		return fmt.Errorf("plot: start %d is after subtitles (%.0f)", first, lo)
	}
	if float64(last)+slack < hi {
		return fmt.Errorf("plot: end %d is before subtitles (%.0f)", last, hi)
	}
	return nil
}

func requireNonEmpty(sub SubSegment) error {
	switch {
	case strings.TrimSpace(sub.Beat) == "":
		return fmt.Errorf("empty beat")
	case strings.TrimSpace(sub.Summary) == "":
		return fmt.Errorf("empty summary")
	case strings.TrimSpace(sub.KeyDialogue) == "":
		return fmt.Errorf("empty key_dialogue")
	case strings.TrimSpace(sub.VisualScene) == "":
		return fmt.Errorf("empty visual_scene")
	case strings.TrimSpace(sub.CharacterMotivation) == "":
		return fmt.Errorf("empty character_motivation")
	case strings.TrimSpace(sub.Emotion) == "":
		return fmt.Errorf("empty emotion")
	case strings.TrimSpace(sub.StorySoFar) == "":
		return fmt.Errorf("empty story_so_far")
	case strings.TrimSpace(sub.SpoilersAvoided) == "":
		return fmt.Errorf("empty spoilers_avoided")
	}
	return nil
}

func cueSpan(cues []Cue) (float64, float64) {
	if len(cues) == 0 {
		return 0, 0
	}
	lo, hi := cues[0].StartSec, cues[0].EndSec
	for _, c := range cues[1:] {
		if c.StartSec < lo {
			lo = c.StartSec
		}
		if c.EndSec > hi {
			hi = c.EndSec
		}
	}
	return lo, hi
}
