package plot

import (
	"os"
	"testing"
)

func TestParseSRT(t *testing.T) {
	raw, err := os.ReadFile("testdata/sample.srt")
	if err != nil {
		t.Fatal(err)
	}
	cues, err := (&Parser{}).Parse(string(raw), "srt")
	if err != nil {
		t.Fatal(err)
	}
	if len(cues) != 2 {
		t.Fatalf("len=%d", len(cues))
	}
	if cues[0].StartSec != 1 || cues[0].EndSec != 3.5 {
		t.Fatalf("srt times %+v", cues[0])
	}
	if cues[0].Text != "夜神月：这个世界正在腐烂。" {
		t.Fatalf("srt text %q", cues[0].Text)
	}
}

func TestParseASS(t *testing.T) {
	raw, err := os.ReadFile("testdata/sample.ass")
	if err != nil {
		t.Fatal(err)
	}
	cues, err := (&Parser{}).Parse(string(raw), "ass")
	if err != nil {
		t.Fatal(err)
	}
	if len(cues) != 2 {
		t.Fatalf("len=%d", len(cues))
	}
	if cues[0].StartSec != 1 || cues[0].EndSec != 3.5 {
		t.Fatalf("ass times %+v", cues[0])
	}
	if cues[0].Speaker != "月" {
		t.Fatalf("speaker %q", cues[0].Speaker)
	}
	if cues[1].Text != "嘿嘿，有意思。" {
		t.Fatalf("ass tags not stripped: %q", cues[1].Text)
	}
}
