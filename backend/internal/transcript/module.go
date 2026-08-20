package transcript

import "withyou/internal/library"

// Dependencies 是 transcript 需要的外部能力。
type Dependencies struct {
	Media *library.Media
}

// Module 组装对话记录存储。
type Module struct {
	Store *Store
	HTTP  *HTTP
}

// New 建 Store 并挂好 HTTP。
func New(deps Dependencies) *Module {
	store := newStore()
	return &Module{
		Store: store,
		HTTP:  &HTTP{media: deps.Media, store: store},
	}
}
