package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("OPENVPN_AS_EXPORTER_SERVER_LISTEN_ADDRESS")

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
	os.Setenv("OPENVPN_AS_EXPORTER_SERVER_LISTEN_ADDRESS", ":8080")
	defer os.Unsetenv("OPENVPN_AS_EXPORTER_SERVER_LISTEN_ADDRESS")

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, ":8080", cfg.Server.ListenAddress)
}

func TestLoadFromValidFile(t *testing.T) {
	f, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString("server:\n  listen_address: \":9090\"\n")
	require.NoError(t, err)
	f.Close()

	cfg, err := Load(f.Name())
	require.NoError(t, err)
	assert.Equal(t, ":9090", cfg.Server.ListenAddress)
	// Defaults still apply for unspecified fields
	assert.Equal(t, "/metrics", cfg.Server.MetricsPath)
}

func TestLoadFromCorruptedFile(t *testing.T) {
	f, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString("\x00invalid: yaml\n")
	require.NoError(t, err)
	f.Close()

	_, err = Load(f.Name())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse config file")
	assert.Contains(t, err.Error(), f.Name())
}

func TestLoadFromNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse config file")
}

func TestLoadEmptyFileUsesDefaults(t *testing.T) {
	f, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.Close()

	cfg, err := Load(f.Name())
	require.NoError(t, err)
	assert.Equal(t, ":9176", cfg.Server.ListenAddress)
}
