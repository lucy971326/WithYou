package plot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"withyou/internal/library"
)

var (
	// ErrNoFile 还没打开视频。
	ErrNoFile = errors.New("plot: no file open")
	// ErrNoSubtitle 没有软字幕轨。
	ErrNoSubtitle = errors.New("plot: no soft subtitle track")
	// ErrImageSubtitle 图字幕（PGS 等），V0 不做。
	ErrImageSubtitle = errors.New("plot: image-based subtitle track")
)

// Extractor 从当前打开的视频抽出软字幕并解析。
type Extractor struct {
	media  *library.Media
	parser *Parser
	state  *state
}

// Extract 读 library 当前原片路径，抽轨、解析、写入状态。
func (e *Extractor) Extract(ctx context.Context) (SubtitleDoc, error) {
	file, ok := e.media.Current()
	if !ok {
		return SubtitleDoc{}, ErrNoFile
	}
	return e.ExtractPath(ctx, file.Path)
}

// ExtractPath 按路径抽轨解析，供测试和 Extract 共用。
func (e *Extractor) ExtractPath(ctx context.Context, path string) (SubtitleDoc, error) {
	codec, err := probeSubtitle(ctx, path)
	if err != nil {
		return SubtitleDoc{}, err
	}
	format, err := subtitleFormat(codec)
	if err != nil {
		return SubtitleDoc{}, err
	}
	raw, err := extractTrack(ctx, path, format)
	if err != nil {
		return SubtitleDoc{}, err
	}
	cues, err := e.parser.Parse(raw, format)
	if err != nil {
		return SubtitleDoc{}, err
	}
	doc := SubtitleDoc{Format: format, Count: len(cues), Cues: cues}
	e.state.replace(doc)
	return doc, nil
}

type probeOut struct {
	Streams []struct {
		CodecName string `json:"codec_name"`
	} `json:"streams"`
}

func probeSubtitle(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "s:0",
		"-show_entries", "stream=codec_name",
		"-of", "json",
		path,
	)
	hideWindow(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("plot: ffprobe: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var out probeOut
	err = json.Unmarshal(stdout.Bytes(), &out)
	if err != nil {
		return "", fmt.Errorf("plot: parse ffprobe: %w", err)
	}
	if len(out.Streams) == 0 || out.Streams[0].CodecName == "" {
		return "", ErrNoSubtitle
	}
	return out.Streams[0].CodecName, nil
}

func subtitleFormat(codec string) (string, error) {
	switch strings.ToLower(codec) {
	case "ass", "ssa":
		return "ass", nil
	case "subrip", "srt", "mov_text", "webvtt":
		return "srt", nil
	case "hdmv_pgs_subtitle", "dvd_subtitle", "dvb_subtitle":
		return "", ErrImageSubtitle
	default:
		return "", fmt.Errorf("plot: unsupported subtitle codec %q", codec)
	}
}

func extractTrack(ctx context.Context, path, format string) (string, error) {
	dir, err := os.MkdirTemp("", "withyou-sub-")
	if err != nil {
		return "", fmt.Errorf("plot: temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	out := filepath.Join(dir, "track."+format)
	args := []string{"-y", "-i", path, "-map", "0:s:0"}
	if format == "ass" {
		args = append(args, "-c:s", "copy", out)
	} else {
		args = append(args, "-c:s", "srt", out)
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	hideWindow(cmd)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("plot: ffmpeg extract: %w: %s", err, lastLine(stderr.String()))
	}
	data, err := os.ReadFile(out)
	if err != nil {
		return "", fmt.Errorf("plot: read extracted sub: %w", err)
	}
	return string(data), nil
}

func lastLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}
