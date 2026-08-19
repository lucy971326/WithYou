package library

import "mime"

// Module 组装「打开并播本机视频」。业务在 Picker / Playable / Media / HTTP 上。
type Module struct {
	Picker   *Picker
	Playable *Playable
	Media    *Media
	HTTP     *HTTP
}

// New 校验依赖并组装。本模块暂无外部依赖。
func New() *Module {
	_ = mime.AddExtensionType(".mkv", "video/x-matroska")
	_ = mime.AddExtensionType(".mp4", "video/mp4")

	s := &state{}
	picker := &Picker{}
	playable := &Playable{state: s}
	media := &Media{state: s}
	return &Module{
		Picker:   picker,
		Playable: playable,
		Media:    media,
		HTTP:     &HTTP{picker: picker, playable: playable, media: media},
	}
}
