package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/whg517/ovpn-sa-export/internal/config"
	"github.com/whg517/ovpn-sa-export/internal/metrics"
)

// Server serves Prometheus metrics and health endpoints.
type Server struct {
	cfg      config.ServerConfig
	registry *metrics.Registry
	server   *http.Server
	ready    bool
}

// New creates a new HTTP server.
func New(cfg config.ServerConfig, registry *metrics.Registry) *Server {
	s := &Server{
		cfg:      cfg,
		registry: registry,
		ready:    true,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(cfg.MetricsPath, s.handleMetrics)
	mux.HandleFunc(cfg.HealthPath, s.handleHealth)
	mux.HandleFunc(cfg.ReadyPath, s.handleReady)

	addr := cfg.ListenAddress
	if addr == "" {
		addr = ":9176"
	}

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	slog.Info("starting HTTP server", "addr", s.cfg.ListenAddress, "metrics", s.cfg.MetricsPath)
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() {
	s.ready = false
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()
	s.server.Shutdown(ctx)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if s.ready {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "Not Ready")
	}
}
