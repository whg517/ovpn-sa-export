package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/whg517/ovpn-sa-export/internal/config"
	"github.com/whg517/ovpn-sa-export/internal/metrics"
	"github.com/whg517/ovpn-sa-export/pkg/types"
)

// Collector coordinates periodic data collection from a backend.
type Collector struct {
	cfg      *config.Config
	registry *metrics.Registry
	backend  Backend

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Backend abstracts the data collection interface used by Collector.
type Backend interface {
	Name() string
	CollectVPNStatus(ctx context.Context) ([]types.VPNClientStatus, error)
	CollectVPNSummary(ctx context.Context) (*types.VPNSummary, error)
	CollectServiceStatus(ctx context.Context) (*types.ServiceStatus, error)
}

// New creates a new Collector.
func New(ctx context.Context, cfg *config.Config, registry *metrics.Registry) (*Collector, error) {
	// TODO: Support multiple instances
	// For now, create backend from cfg.Backend
	backend, err := createBackend(cfg)
	if err != nil {
		return nil, err
	}

	return &Collector{
		cfg:      cfg,
		registry: registry,
		backend:  backend,
	}, nil
}

// Start begins periodic collection.
func (c *Collector) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	interval := c.cfg.Collector.ScrapeInterval
	if interval == 0 {
		interval = 15 * time.Second
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Initial collection
		c.collectAll(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.collectAll(ctx)
			}
		}
	}()

	return nil
}

// Stop halts collection.
func (c *Collector) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
}

func (c *Collector) collectAll(ctx context.Context) {
	c.collectVPNStatus(ctx)
	c.collectVPNSummary(ctx)
	c.collectServiceStatus(ctx)
}

func (c *Collector) collectVPNStatus(ctx context.Context) {
	start := time.Now()
	clients, err := c.backend.CollectVPNStatus(ctx)
	duration := time.Since(start)

	c.registry.RecordScrape("vpn_status", duration, err)

	if err != nil {
		slog.Error("failed to collect VPN status", "error", err, "backend", c.backend.Name())
		return
	}

	c.registry.UpdateVPNStatus(clients)
}

func (c *Collector) collectVPNSummary(ctx context.Context) {
	start := time.Now()
	summary, err := c.backend.CollectVPNSummary(ctx)
	duration := time.Since(start)

	c.registry.RecordScrape("vpn_summary", duration, err)

	if err != nil {
		slog.Error("failed to collect VPN summary", "error", err, "backend", c.backend.Name())
		return
	}

	c.registry.UpdateVPNSummary(summary)
}

func (c *Collector) collectServiceStatus(ctx context.Context) {
	start := time.Now()
	status, err := c.backend.CollectServiceStatus(ctx)
	duration := time.Since(start)

	c.registry.RecordScrape("service", duration, err)

	if err != nil {
		slog.Error("failed to collect service status", "error", err, "backend", c.backend.Name())
		return
	}

	c.registry.UpdateServiceStatus(status)
}
