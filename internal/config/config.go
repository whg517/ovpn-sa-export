package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Backend   BackendConfig   `mapstructure:"backend"`
	Instances []InstanceConfig `mapstructure:"instances"`
	Server    ServerConfig    `mapstructure:"server"`
	Collector CollectorConfig `mapstructure:"collector"`
	Log       LogConfig       `mapstructure:"log"`
}

type BackendConfig struct {
	Mode   string        `mapstructure:"mode"`
	Sacli  SacliConfig   `mapstructure:"sacli"`
	XMLRPC XMLRPCConfig  `mapstructure:"xmlrpc"`
}

type SacliConfig struct {
	Path    string        `mapstructure:"path"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type XMLRPCConfig struct {
	Endpoint              string        `mapstructure:"endpoint"`
	Username              string        `mapstructure:"username"`
	Password              string        `mapstructure:"password"`
	SocketPath            string        `mapstructure:"socket_path"`
	Timeout               time.Duration `mapstructure:"timeout"`
	InsecureSkipVerify    bool          `mapstructure:"insecure_skip_verify"`
}

type InstanceConfig struct {
	Name    string        `mapstructure:"name"`
	Backend BackendConfig `mapstructure:"backend"`
}

type ServerConfig struct {
	ListenAddress string `mapstructure:"listen_address"`
	MetricsPath   string `mapstructure:"metrics_path"`
	HealthPath    string `mapstructure:"health_path"`
	ReadyPath     string `mapstructure:"ready_path"`
}

type CollectorConfig struct {
	ScrapeInterval   time.Duration `mapstructure:"scrape_interval"`
	Timeout          time.Duration `mapstructure:"timeout"`
	CacheTTL         time.Duration `mapstructure:"cache_ttl"`
	EnabledCollectors []string     `mapstructure:"enabled_collectors"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("backend.mode", "sacli")
	v.SetDefault("backend.sacli.path", "/usr/local/openvpn_as/scripts/sacli")
	v.SetDefault("backend.sacli.timeout", "10s")
	v.SetDefault("backend.xmlrpc.timeout", "10s")
	v.SetDefault("backend.xmlrpc.insecure_skip_verify", false)
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
	v.SetConfigName("ovpn-sa-export")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/ovpn-sa-export")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.ovpn-sa-export")

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
