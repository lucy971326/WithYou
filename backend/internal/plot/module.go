package plot

import "withyou/internal/library"

// Dependencies 是 plot 需要的外部能力。
type Dependencies struct {
	Media       *library.Media
	QwenAPIKey  string
	QwenModel   string
	QwenBaseURL string
}

// Module 组装抽字幕 + 富化。
type Module struct {
	Extractor *Extractor
	Parser    *Parser
	Cache     *Cache
	Enricher  *Enricher
	HTTP      *HTTP
}

// New 校验依赖并组装。
func New(deps Dependencies) *Module {
	if deps.Media == nil {
		panic("plot: Media is required")
	}
	s := &state{}
	parser := &Parser{}
	extractor := &Extractor{media: deps.Media, parser: parser, state: s}
	cache := newCache()
	client, model, ready := newClient(deps.QwenAPIKey, deps.QwenModel, deps.QwenBaseURL)
	enricher := &Enricher{
		media:  deps.Media,
		state:  s,
		cache:  cache,
		client: client,
		model:  model,
		ready:  ready,
	}
	return &Module{
		Extractor: extractor,
		Parser:    parser,
		Cache:     cache,
		Enricher:  enricher,
		HTTP:      &HTTP{extractor: extractor, enricher: enricher, state: s},
	}
}
