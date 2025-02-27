package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

func Init(config *Config) error {
	zapConfig := zap.NewProductionConfig()
	zapConfig.DisableStacktrace = !config.StackTrace
	zapConfig.Encoding = config.Encoding
	if err := zapConfig.Level.UnmarshalText([]byte(config.Level)); err != nil {
		return nil
	}

	if Log != nil {
		_ = Log.Sync()
		return nil
	}

	var err error
	Log, err = zapConfig.Build()
	return err
}

func Set(logger *zap.Logger) {
	Log = logger
}
