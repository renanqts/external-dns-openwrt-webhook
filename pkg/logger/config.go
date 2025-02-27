package logger

type Config struct {
	Level      string `mapstructure:"level"`
	StackTrace bool   `mapstructure:"stack_trace"`
	Encoding   string `mapstructure:"encoding"`
}

func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		StackTrace: false,
		Encoding:   "json",
	}
}
