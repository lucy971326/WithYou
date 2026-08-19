package library

import (
	"context"
	"log"
)

// Playable 探测编码，原片直给浏览器。不解的编码不转码，只标记。
type Playable struct {
	state *state
}

// Ensure 写入当前打开状态。探测失败不挡播放。
func (p *Playable) Ensure(ctx context.Context, file OpenedFile) (OpenedFile, error) {
	file.PlayablePath = file.Path
	stream, err := probeVideo(ctx, file.Path)
	if err != nil {
		log.Printf("library probe: %v", err)
	} else {
		file.Codec = stream.CodecName
		file.PixelFormat = stream.PixFmt
		file.BrowserSafe = browserSafe(stream)
	}
	p.state.replace(file)
	return file, nil
}
