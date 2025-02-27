package router

type Gin struct {
	ReleaseMode bool `mapstructure:"release_mode"`
}

type Config struct {
	HealthCheckPath string `mapstructure:"healthcheck_path"`
	Port            string `mapstructure:"port"`
	Gin             Gin    `mapstructure:"gin"`
}

func DefaultConfig() *Config {
	return &Config{
		HealthCheckPath: "/ping",
		Port:            "8888",
		Gin: Gin{
			ReleaseMode: true,
		},
	}
}
