package plot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/responses"

	"withyou/internal/library"
)

var (
	// ErrNoCues 还没抽出对白。
	ErrNoCues = errors.New("plot: no subtitles extracted")
	// ErrNoAPIKey 缺少 Qwen key。
	ErrNoAPIKey = errors.New("plot: missing DASHSCOPE_API_KEY")
	// ErrEmptyContent JSON Output 返回了空 content。
	ErrEmptyContent = errors.New("plot: empty json content")
	// ErrInvalidSchema 重试后 schema 仍不过。
	ErrInvalidSchema = errors.New("plot: invalid plot json")
)

const enrichAttempts = 2

// Enricher 一次 Responses API + strict JSON Schema，失败再试一次。
type Enricher struct {
	media  *library.Media
	state  *state
	cache  *Cache
	client openai.Client
	model  string
	ready  bool
}

// Enrich 命中缓存直接回；否则走 Responses API 结构化输出。
func (e *Enricher) Enrich(ctx context.Context) (PlotDoc, bool, error) {
	if !e.ready {
		return PlotDoc{}, false, ErrNoAPIKey
	}
	file, ok := e.media.Current()
	if !ok {
		return PlotDoc{}, false, ErrNoFile
	}
	cues, ok := e.state.current()
	if !ok || len(cues.Cues) == 0 {
		return PlotDoc{}, false, ErrNoCues
	}

	if doc, hit := e.cache.Get(file.Path, file.Size); hit {
		err := validatePlot(doc, cues.Cues)
		if err == nil {
			e.state.replacePlot(doc)
			return doc, true, nil
		}
	}

	title := strings.TrimSuffix(file.Name, filepath.Ext(file.Name))
	var last error
	for attempt := 1; attempt <= enrichAttempts; attempt++ {
		raw, err := e.completeJSON(ctx, title, cues.Cues)
		if err != nil {
			last = err
			log.Printf("plot enrich attempt %d: %v", attempt, err)
			continue
		}
		var doc PlotDoc
		err = json.Unmarshal([]byte(raw), &doc)
		if err != nil {
			last = fmt.Errorf("plot: parse json: %w", err)
			log.Printf("plot enrich attempt %d: %v", attempt, last)
			continue
		}
		err = validatePlot(doc, cues.Cues)
		if err != nil {
			last = err
			log.Printf("plot enrich attempt %d: %v", attempt, err)
			continue
		}
		err = e.cache.Put(file.Path, file.Size, doc)
		if err != nil {
			log.Printf("plot cache put: %v", err)
		}
		e.state.replacePlot(doc)
		return doc, false, nil
	}
	if last == nil {
		last = ErrInvalidSchema
	}
	return PlotDoc{}, false, last
}

func (e *Enricher) completeJSON(ctx context.Context, title string, cues []Cue) (string, error) {
	resp, err := e.client.Responses.New(ctx, responses.ResponseNewParams{
		Model:        e.model,
		Instructions: openai.String(enrichSystemPrompt),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(buildUserPrompt(title, cues)),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "plot_archive",
					Strict: openai.Bool(true),
					Schema: plotJSONSchema(),
				},
			},
		},
	}, option.WithJSONSet("tools", []map[string]any{
		{"type": "web_search"},
		{"type": "web_extractor"},
	}))
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(resp.OutputText()) == "" {
		return "", ErrEmptyContent
	}
	return resp.OutputText(), nil
}

func buildUserPrompt(title string, cues []Cue) string {
	var b strings.Builder
	fmt.Fprintf(&b, "作品：%s\n", title)
	b.WriteString("世界观与角色背景：未提供世界背景。请仅依据字幕与你的通用知识进行分析，不要臆造超出字幕范围的具体设定。\n\n")
	b.WriteString("本集对白字幕（时间戳秒）。请按 system 里的 EXAMPLE JSON OUTPUT 输出一个 json 对象：\n")
	for _, c := range cues {
		fmt.Fprintf(&b, "[%.1f-%.1f] %s\n", c.StartSec, c.EndSec, c.Text)
	}
	return b.String()
}

func newClient(apiKey, model, baseURL string) (openai.Client, string, bool) {
	if apiKey == "" {
		return openai.Client{}, model, false
	}
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
		option.WithRequestTimeout(10*time.Minute),
	)
	return client, model, true
}

func plotJSONSchema() map[string]any {
	text := map[string]any{"type": "string"}
	segment := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"start_sec":            map[string]any{"type": "integer"},
			"end_sec":              map[string]any{"type": "integer"},
			"beat":                 text,
			"summary":              text,
			"key_dialogue":         text,
			"visual_scene":         text,
			"character_motivation": text,
			"emotion":              text,
			"story_so_far":         text,
			"spoilers_avoided":     text,
		},
		"required": []string{
			"start_sec", "end_sec", "beat", "summary", "key_dialogue",
			"visual_scene", "character_motivation", "emotion", "story_so_far",
			"spoilers_avoided",
		},
		"additionalProperties": false,
	}
	major := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"start_sec":    map[string]any{"type": "integer"},
			"end_sec":      map[string]any{"type": "integer"},
			"title":        text,
			"summary":      text,
			"sub_segments": map[string]any{"type": "array", "items": segment},
		},
		"required":             []string{"start_sec", "end_sec", "title", "summary", "sub_segments"},
		"additionalProperties": false,
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": text,
			"overview": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"grand_summary":   text,
					"key_characters":  map[string]any{"type": "array", "items": text},
					"key_plot_points": map[string]any{"type": "array", "items": text},
				},
				"required":             []string{"grand_summary", "key_characters", "key_plot_points"},
				"additionalProperties": false,
			},
			"major_segments": map[string]any{"type": "array", "items": major},
		},
		"required":             []string{"title", "overview", "major_segments"},
		"additionalProperties": false,
	}
}

func summarize(doc PlotDoc, cached bool) EnrichResponse {
	subs := 0
	for _, m := range doc.MajorSegments {
		subs += len(m.SubSegments)
	}
	return EnrichResponse{
		Title:        doc.Title,
		Cached:       cached,
		MajorCount:   len(doc.MajorSegments),
		SubCount:     subs,
		GrandSummary: doc.Overview.GrandSummary,
	}
}
