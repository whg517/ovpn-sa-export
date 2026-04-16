package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("OVPN_SA_SERVER_LISTEN_ADDRESS")

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, ":9176", cfg.Server.ListenAddress)
	assert.Equal(t, "/metrics", cfg.Server.MetricsPath)
	assert.Equal(t, "/health", cfg.Server.HealthPath)
	assert.Equal(t, "/ready", cfg.Server.ReadyPath)
	assert.Equal(t, 15*time.Second, cfg.Collector.ScrapeInterval)
	assert.Equal(t, 10*time.Second, cfg.Backend.Sacli.Timeout)
	assert.Equal(t, "/usr/local/openvpn_as/scripts/sacli", cfg.Backend.Sacli.Path)
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("OVPN_SA_SERVER_LISTEN_ADDRESS", ":8080")
	defer os.Unsetenv("OVPN_SA_SERVER_LISTEN_ADDRESS")

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, ":8080", cfg.Server.ListenAddress)
}
