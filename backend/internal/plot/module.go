package plot

import "withyou/internal/library"

// Dependencies 是 plot 需要的外部能力。
type Dependencies struct {
	Media *library.Media
}

// Module 组装「抽软字幕并解析」。
type Module struct {
	Extractor *Extractor
	Parser    *Parser
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
	return &Module{
		Extractor: extractor,
		Parser:    parser,
		HTTP:      &HTTP{extractor: extractor, state: s},
	}
}
