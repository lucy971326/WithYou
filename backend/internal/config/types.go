package config

const (
	DefaultQwenSite = "intl"

	DefaultQwenPlotModel     = "qwen3.8-max"
	DefaultQwenRealtimeModel = "qwen3.5-omni-plus-realtime"
	DefaultQwenRealtimeVoice = "Tina"

	DefaultQwenPlotBaseURLIntl = "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"
	DefaultQwenPlotBaseURLCN   = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	DefaultQwenRealtimeURLIntl = "wss://dashscope-intl.aliyuncs.com/api-ws/v1/realtime"
	DefaultQwenRealtimeURLCN   = "wss://dashscope.aliyuncs.com/api-ws/v1/realtime"
	DefaultQwenVoiceAPIURLIntl = "https://dashscope-intl.aliyuncs.com/api/v1/services/audio/tts/customization"
	DefaultQwenVoiceAPIURLCN   = "https://dashscope.aliyuncs.com/api/v1/services/audio/tts/customization"
)

// Config 是进程级配置，字段只表达身份与取值，没有行为。
type Config struct {
	Addr              string
	QwenSite          string
	QwenAPIKey        string
	QwenPlotModel     string
	QwenPlotBaseURL   string
	QwenRealtimeModel string
	QwenRealtimeURL   string
	QwenRealtimeVoice string
	QwenVoiceAPIURL   string
}
