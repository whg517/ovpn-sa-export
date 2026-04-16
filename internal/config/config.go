package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Backend   BackendConfig  `mapstructure:"backend"`
	Server    ServerConfig   `mapstructure:"server"`
	Collector CollectorConfig `mapstructure:"collector"`
	Log       LogConfig      `mapstructure:"log"`
}

type BackendConfig struct {
	Sacli SacliConfig `mapstructure:"sacli"`
}

type SacliConfig struct {
	Path    string        `mapstructure:"path"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type ServerConfig struct {
	ListenAddress string `mapstructure:"listen_address"`
	MetricsPath   string `mapstructure:"metrics_path"`
	HealthPath    string `mapstructure:"health_path"`
	ReadyPath     string `mapstructure:"ready_path"`
}

type CollectorConfig struct {
	ScrapeInterval    time.Duration `mapstructure:"scrape_interval"`
	Timeout           time.Duration `mapstructure:"timeout"`
	CacheTTL          time.Duration `mapstructure:"cache_ttl"`
	EnabledCollectors []string      `mapstructure:"enabled_collectors"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load(configFile string) (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("backend.sacli.path", "/usr/local/openvpn_as/scripts/sacli")
	v.SetDefault("backend.sacli.timeout", "10s")
	v.SetDefault("server.listen_address", ":9176")
	v.SetDefault("server.metrics_path", "/metrics")
	v.SetDefault("server.health_path", "/health")
	v.SetDefault("server.ready_path", "/ready")
	v.SetDefault("collector.scrape_interval", "15s")
	v.SetDefault("collector.timeout", "30s")
	v.SetDefault("collector.cache_ttl", "60s")
	v.SetDefault("collector.enabled_collectors", []string{
		"vpn_status", "vpn_summary", "service",
	})
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	// Environment variables (always loaded)
	v.SetEnvPrefix("OPENVPN_AS_EXPORTER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Config file (only when explicitly provided)
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("parse config file %s: %w", configFile, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
