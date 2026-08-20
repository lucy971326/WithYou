package realtime

// Dependencies 是 realtime 需要的配置。
type Dependencies struct {
	APIKey      string
	Model       string
	URL         string
	Voice       string
	VoiceAPIURL string
}

// Module 组装 Realtime 中继。
type Module struct {
	Relay *Relay
	HTTP  *HTTP
}

// New 组装中继。缺 key 时仍可 Register，升级时返回 503。
func New(deps Dependencies) *Module {
	relay := &Relay{
		apiKey:      deps.APIKey,
		model:       deps.Model,
		base:        deps.URL,
		quietCounts: map[string]int{},
	}
	return &Module{
		Relay: relay,
		HTTP: &HTTP{
			relay:       relay,
			voice:       deps.Voice,
			voiceAPIURL: deps.VoiceAPIURL,
		},
	}
}
