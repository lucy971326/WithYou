package plot

import "testing"

func sampleDoc() PlotDoc {
	return PlotDoc{
		Title: "Death Note",
		Overview: PlotOverview{
			GrandSummary:  "Light finds the notebook.",
			KeyCharacters: []string{"Light"},
			KeyPlotPoints: []string{"Notebook"},
		},
		MajorSegments: []MajorSegment{{
			StartSec: 0,
			EndSec:   20,
			Title:    "Open",
			Summary:  "He tests it.",
			SubSegments: []SubSegment{{
				StartSec:            0,
				EndSec:              20,
				Beat:                "Test",
				Summary:             "He writes a name.",
				KeyDialogue:         "这个世界正在腐烂。",
				VisualScene:         "Desk, night.",
				CharacterMotivation: "Play god.",
				Emotion:             "Cold",
				StorySoFar:          "None.",
				SpoilersAvoided:     "L appears later.",
			}},
		}},
	}
}

func TestValidatePlotOK(t *testing.T) {
	cues := []Cue{{StartSec: 1, EndSec: 3.5, Text: "a"}, {StartSec: 4, EndSec: 18, Text: "b"}}
	if err := validatePlot(sampleDoc(), cues); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePlotEmptyField(t *testing.T) {
	doc := sampleDoc()
	doc.MajorSegments[0].SubSegments[0].Emotion = ""
	if err := validatePlot(doc, nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestCacheRoundTrip(t *testing.T) {
	c := &Cache{dir: t.TempDir()}
	doc := sampleDoc()
	if err := c.Put("D:\\a.mkv", 12, doc); err != nil {
		t.Fatal(err)
	}
	got, ok := c.Get("D:\\a.mkv", 12)
	if !ok || got.Title != "Death Note" {
		t.Fatalf("got=%+v ok=%v", got, ok)
	}
	if _, ok := c.Get("D:\\a.mkv", 99); ok {
		t.Fatal("different size should miss")
	}
}
