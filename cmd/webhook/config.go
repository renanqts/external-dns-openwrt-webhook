package main

import (
	"github.com/renanqts/external-dns-openwrt-webhook/internal/provider"
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/logger"
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/router"
)

type Config struct {
	ShutodwnTimeout int              `mapstructure:"shutdown_timeout_seconds"`
	Log             *logger.Config   `mapstructure:"log"`
	Router          *router.Config   `mapstructure:"router"`
	Provider        *provider.Config `mapstructure:"provider"`
}

func defaultConfig() *Config {
	return &Config{
		ShutodwnTimeout: 5,
		Log:             logger.DefaultConfig(),
		Router:          router.DefaultConfig(),
		Provider:        provider.DefaultConfig(),
	}
}
