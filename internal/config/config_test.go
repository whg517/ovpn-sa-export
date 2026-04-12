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
	os.Unsetenv("OVPN_SA_BACKEND_MODE")
	os.Unsetenv("OVPN_SA_BACKEND_XMLRPC_ENDPOINT")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, ":9176", cfg.Server.ListenAddress)
	assert.Equal(t, "/metrics", cfg.Server.MetricsPath)
	assert.Equal(t, "/health", cfg.Server.HealthPath)
	assert.Equal(t, "/ready", cfg.Server.ReadyPath)
	assert.Equal(t, "sacli", cfg.Backend.Mode)
	assert.Equal(t, 15*time.Second, cfg.Collector.ScrapeInterval)
	assert.Equal(t, 10*time.Second, cfg.Backend.Sacli.Timeout)
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("OVPN_SA_SERVER_LISTEN_ADDRESS", ":8080")
	os.Setenv("OVPN_SA_BACKEND_MODE", "xmlrpc")
	defer func() {
		os.Unsetenv("OVPN_SA_SERVER_LISTEN_ADDRESS")
		os.Unsetenv("OVPN_SA_BACKEND_MODE")
	}()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, ":8080", cfg.Server.ListenAddress)
	assert.Equal(t, "xmlrpc", cfg.Backend.Mode)
}
