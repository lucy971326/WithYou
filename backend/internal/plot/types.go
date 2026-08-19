package plot

// Cue 是一条带时间戳的对白。
type Cue struct {
	StartSec float64 `json:"start_sec"`
	EndSec   float64 `json:"end_sec"`
	Speaker  string  `json:"speaker,omitempty"`
	Text     string  `json:"text"`
}

// SubtitleDoc 是抽轨并解析后的结果。
type SubtitleDoc struct {
	Format string `json:"format"`
	Count  int    `json:"count"`
	Cues   []Cue  `json:"cues"`
}

// ErrorResponse 是 HTTP 错误体。
type ErrorResponse struct {
	Error string `json:"error"`
}
