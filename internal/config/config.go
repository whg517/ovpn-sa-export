package config

import (
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
		"vpn_status", "vpn_summary", "subscription", "service",
	})
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	// Config file
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("ovpn-sa-export")
		v.SetConfigType("yaml")
		v.AddConfigPath("/etc/ovpn-sa-export")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.ovpn-sa-export")
	}

	// Environment variables
	v.SetEnvPrefix("OVPN_SA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
