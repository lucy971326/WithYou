package config

const (
	DefaultQwenPlotModel     = "qwen3.8-max"
	DefaultQwenRealtimeModel = "qwen3.5-omni-plus-realtime"
)

// Config 是进程级配置，字段只表达身份与取值，没有行为。
type Config struct {
	Addr              string
	QwenAPIKey        string
	QwenPlotModel     string
	QwenRealtimeModel string
	QwenRealtimeURL   string
}
