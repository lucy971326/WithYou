package plot

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"withyou/internal/library"
)

func TestExtractPathFromMuxedSRT(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not on PATH")
	}
	dir := t.TempDir()
	mkv := filepath.Join(dir, "soft.mkv")
	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "lavfi", "-i", "color=c=black:s=64x64:d=2",
		"-f", "lavfi", "-i", "anullsrc=r=8000:cl=mono:d=2",
		"-i", "testdata/sample.srt",
		"-c:v", "libx264", "-preset", "ultrafast",
		"-c:a", "aac",
		"-c:s", "srt",
		"-map", "0:v", "-map", "1:a", "-map", "2:s",
		mkv,
	)
	hideWindow(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("mux mkv: %v\n%s", err, out)
	}

	mod := newTestModule(t)
	doc, err := mod.Extractor.ExtractPath(context.Background(), mkv)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Format != "srt" || doc.Count != 2 {
		t.Fatalf("doc=%+v", doc)
	}
	if doc.Cues[0].Text != "夜神月：这个世界正在腐烂。" {
		t.Fatalf("cue %q", doc.Cues[0].Text)
	}
}

func TestExtractNoSubtitle(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not on PATH")
	}
	dir := t.TempDir()
	mkv := filepath.Join(dir, "bare.mkv")
	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "lavfi", "-i", "color=c=black:s=64x64:d=1",
		"-c:v", "libx264", "-preset", "ultrafast",
		mkv,
	)
	hideWindow(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("mux mkv: %v\n%s", err, out)
	}
	mod := newTestModule(t)
	_, err := mod.Extractor.ExtractPath(context.Background(), mkv)
	if err != ErrNoSubtitle {
		t.Fatalf("err=%v", err)
	}
}

func newTestModule(t *testing.T) *Module {
	t.Helper()
	lib := library.New()
	return New(Dependencies{Media: lib.Media})
}
