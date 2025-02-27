package config

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func Read(config any) error {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()

	// AutomaticEnv does not work
	// https://github.com/spf13/viper/issues/188
	var defaultConfigMap map[string]interface{}
	if err := mapstructure.Decode(config, &defaultConfigMap); err != nil {
		return fmt.Errorf("failed to decode default config. Error: %w", err)
	}

	if err := viper.MergeConfigMap(defaultConfigMap); err != nil {
		return fmt.Errorf("failed to merge default config. Error: %w", err)
	}

	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to load config. Error: %w", err)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to unmarshal config. Error: %w", err)
	}

	return nil
}
