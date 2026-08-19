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

	"withyou/internal/library"
)

var (
	// ErrNoCues 还没抽出对白。
	ErrNoCues = errors.New("plot: no subtitles extracted")
	// ErrNoAPIKey 缺少 DeepSeek key。
	ErrNoAPIKey = errors.New("plot: missing DEEPSEEK_API_KEY")
	// ErrEmptyContent JSON Output 返回了空 content。
	ErrEmptyContent = errors.New("plot: empty json content")
	// ErrInvalidSchema 重试后 schema 仍不过。
	ErrInvalidSchema = errors.New("plot: invalid plot json")
)

const enrichAttempts = 2

// Enricher 一次 Chat Completions + json_object，失败再试一次。
type Enricher struct {
	media  *library.Media
	state  *state
	cache  *Cache
	client openai.Client
	model  string
	ready  bool
}

// Enrich 命中缓存直接回；否则走官方 JSON Output。不设 max_tokens。
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
		if err := validatePlot(doc, cues.Cues); err == nil {
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
		if err := json.Unmarshal([]byte(raw), &doc); err != nil {
			last = fmt.Errorf("plot: parse json: %w", err)
			log.Printf("plot enrich attempt %d: %v", attempt, last)
			continue
		}
		if err := validatePlot(doc, cues.Cues); err != nil {
			last = err
			log.Printf("plot enrich attempt %d: %v", attempt, err)
			continue
		}
		if err := e.cache.Put(file.Path, file.Size, doc); err != nil {
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
	resp, err := e.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: e.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(enrichSystemPrompt),
			openai.UserMessage(buildUserPrompt(title, cues)),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{
				Type: "json_object",
			},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 || strings.TrimSpace(resp.Choices[0].Message.Content) == "" {
		return "", ErrEmptyContent
	}
	return resp.Choices[0].Message.Content, nil
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

func newClient(apiKey, model string) (openai.Client, string, bool) {
	if apiKey == "" {
		return openai.Client{}, model, false
	}
	if model == "" {
		model = "deepseek-v4-flash"
	}
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://api.deepseek.com"),
		option.WithRequestTimeout(10*time.Minute),
	)
	return client, model, true
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
