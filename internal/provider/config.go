package provider

import (
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/openwrt"
)

type Config struct {
	OpenWRT *openwrt.Config `mapstructure:"openwrt"`
}

func DefaultConfig() *Config {
	return &Config{
		OpenWRT: openwrt.DefaultConfig(),
	}
}
