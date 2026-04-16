package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/whg517/openvpn-as-exporter/internal/config"
	"github.com/whg517/openvpn-as-exporter/internal/metrics"
)

func newTestServer(t *testing.T) (*Server, *metrics.Registry) {
	t.Helper()
	registry := metrics.NewRegistry()
	cfg := config.ServerConfig{
		ListenAddress: ":0",
		MetricsPath:   "/metrics",
		HealthPath:    "/health",
		ReadyPath:     "/ready",
	}
	return New(cfg, registry), registry
}

func TestHealthEndpoint(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.handleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestReadyEndpoint(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	srv.handleReady(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestReadyEndpointNotReady(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.ready = false

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	srv.handleReady(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestNewDefaultAddress(t *testing.T) {
	registry := metrics.NewRegistry()
	cfg := config.ServerConfig{
		MetricsPath: "/metrics",
		HealthPath:  "/health",
		ReadyPath:   "/ready",
	}
	srv := New(cfg, registry)
	assert.NotNil(t, srv)
	assert.Equal(t, ":9176", srv.server.Addr)
}

func TestShutdown(t *testing.T) {
	srv, _ := newTestServer(t)
	// Shutdown should not panic
	srv.Shutdown()
	assert.False(t, srv.ready)
}

func TestMetricsEndpoint(t *testing.T) {
	_, reg := newTestServer(t)

	// Use promhttp directly with the custom registry
	handler := promhttp.HandlerFor(reg.PromRegistry(), promhttp.HandlerOpts{})
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, strings.Contains(w.Body.String(), "openvpn_as_exporter_"))
}
