package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/whg517/ovpn-sa-export/internal/config"
	"github.com/whg517/ovpn-sa-export/internal/collector"
	"github.com/whg517/ovpn-sa-export/internal/metrics"
	"github.com/whg517/ovpn-sa-export/internal/server"
)

// Version is set at build time via -ldflags.
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	configFile := flag.String("config", "", "Path to config file")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ovpn-sa-export %s\n", version)
		os.Exit(0)
	}

	if err := run(*configFile); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(configFile string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	_ = configFile // TODO: support explicit config file path

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize metrics registry
	registry := metrics.NewRegistry()

	// Initialize collector
	coll, err := collector.New(ctx, cfg, registry)
	if err != nil {
		return fmt.Errorf("create collector: %w", err)
	}
	if err := coll.Start(); err != nil {
		return fmt.Errorf("start collector: %w", err)
	}
	defer coll.Stop()

	// Initialize and start HTTP server
	srv := server.New(cfg.Server, registry)
	go func() {
		<-ctx.Done()
		srv.Shutdown()
	}()

	fmt.Printf("ovpn-sa-export %s starting\n", version)
	fmt.Printf("backend: sacli\n")
	fmt.Printf("listening on: %s\n", cfg.Server.ListenAddress)
	fmt.Printf("metrics path: %s\n", cfg.Server.MetricsPath)

	return srv.ListenAndServe()
}
