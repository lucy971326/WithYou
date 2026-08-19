package config

// Config 是进程级配置，字段只表达身份与取值，没有行为。
type Config struct {
	Addr           string
	QwenAPIKey     string
	DeepSeekAPIKey string
	DeepSeekModel  string
}
