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

// PlotDoc 是 DeepSeek JSON Output 校验后的剧情档案。
type PlotDoc struct {
	Title         string         `json:"title"`
	Overview      PlotOverview   `json:"overview"`
	MajorSegments []MajorSegment `json:"major_segments"`
}

// PlotOverview 是整集总览。
type PlotOverview struct {
	GrandSummary   string   `json:"grand_summary"`
	KeyCharacters  []string `json:"key_characters"`
	KeyPlotPoints  []string `json:"key_plot_points"`
}

// MajorSegment 是一个大阶段。
type MajorSegment struct {
	StartSec     int           `json:"start_sec"`
	EndSec       int           `json:"end_sec"`
	Title        string        `json:"title"`
	Summary      string        `json:"summary"`
	SubSegments  []SubSegment  `json:"sub_segments"`
}

// SubSegment 是一个小阶段/节拍。
type SubSegment struct {
	StartSec             int    `json:"start_sec"`
	EndSec               int    `json:"end_sec"`
	Beat                 string `json:"beat"`
	Summary              string `json:"summary"`
	KeyDialogue          string `json:"key_dialogue"`
	VisualScene          string `json:"visual_scene"`
	CharacterMotivation  string `json:"character_motivation"`
	Emotion              string `json:"emotion"`
	StorySoFar           string `json:"story_so_far"`
	SpoilersAvoided      string `json:"spoilers_avoided"`
}

// EnrichResponse 是 POST/GET /api/plot/enrich 给前端的摘要。
type EnrichResponse struct {
	Title       string `json:"title"`
	Cached      bool   `json:"cached"`
	MajorCount  int    `json:"major_count"`
	SubCount    int    `json:"sub_count"`
	GrandSummary string `json:"grand_summary"`
}

// ErrorResponse 是 HTTP 错误体。
type ErrorResponse struct {
	Error string `json:"error"`
}
