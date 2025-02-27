package lucirpc

const (
	defaultRpcID              = 1
	defaultTimeout            = 15
	defaultInsecureSkipVerify = false
	defaultRpcServerPort      = 443
	defaultSSL                = true
)

type Auth struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Config struct {
	Hostname           string `mapstructure:"hostname"`
	Port               int    `mapstructure:"port"`
	SSL                bool   `mapstructure:"ssl"`
	RpcID              int    `mapstructure:"rpc_id"`
	Timeout            int    `mapstructure:"timeout"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
	Auth               Auth   `mapstructure:"auth"`
}

func DefaultConfig() *Config {
	return &Config{
		Port:               defaultRpcServerPort,
		SSL:                defaultSSL,
		RpcID:              defaultRpcID,
		Timeout:            defaultTimeout,
		InsecureSkipVerify: defaultInsecureSkipVerify,
	}
}
