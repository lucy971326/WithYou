package library

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type videoStream struct {
	CodecName string `json:"codec_name"`
	PixFmt    string `json:"pix_fmt"`
	Profile   string `json:"profile"`
}

type probeOutput struct {
	Streams []videoStream `json:"streams"`
}

func probeVideo(ctx context.Context, path string) (videoStream, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name,pix_fmt,profile",
		"-of", "json",
		path,
	)
	hideWindow(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return videoStream{}, fmt.Errorf("library: ffprobe: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var out probeOutput
	err = json.Unmarshal(stdout.Bytes(), &out)
	if err != nil {
		return videoStream{}, fmt.Errorf("library: parse ffprobe: %w", err)
	}
	if len(out.Streams) == 0 {
		return videoStream{}, fmt.Errorf("library: no video stream")
	}
	return out.Streams[0], nil
}

// browserSafe 是 Chrome 能直接硬解的常见组合。HEVC / 10bit / MPEG-4 ASP 都会黑屏只剩声音。
func browserSafe(stream videoStream) bool {
	codec := strings.ToLower(stream.CodecName)
	pix := strings.ToLower(stream.PixFmt)
	if codec != "h264" {
		return false
	}
	return pix == "yuv420p" || pix == "yuvj420p"
}
