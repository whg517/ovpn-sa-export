package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/whg517/openvpn-as-exporter/internal/config"
	"github.com/whg517/openvpn-as-exporter/internal/metrics"
	"github.com/whg517/openvpn-as-exporter/pkg/types"
)

// mockBackend implements Backend for testing.
type mockBackend struct {
	name            string
	vpnStatusFn     func(ctx context.Context) ([]types.VPNClientStatus, error)
	vpnSummaryFn    func(ctx context.Context) (*types.VPNSummary, error)
	serviceStatusFn func(ctx context.Context) (*types.ServiceStatus, error)
}

func (m *mockBackend) Name() string { return m.name }

func (m *mockBackend) CollectVPNStatus(ctx context.Context) ([]types.VPNClientStatus, error) {
	return m.vpnStatusFn(ctx)
}

func (m *mockBackend) CollectVPNSummary(ctx context.Context) (*types.VPNSummary, error) {
	return m.vpnSummaryFn(ctx)
}

func (m *mockBackend) CollectServiceStatus(ctx context.Context) (*types.ServiceStatus, error) {
	return m.serviceStatusFn(ctx)
}

func testConfig() *config.Config {
	return &config.Config{
		Collector: config.CollectorConfig{
			ScrapeInterval: 50 * time.Millisecond,
		},
		Backend: config.BackendConfig{
			Sacli: config.SacliConfig{
				Path:    "/bin/true",
				Timeout: 1 * time.Second,
			},
		},
	}
}

func newCollectorWithMock(backend Backend) *Collector {
	cfg := testConfig()
	reg := metrics.NewRegistry()
	return &Collector{
		cfg:      cfg,
		registry: reg,
		backend:  backend,
	}
}

func TestCollectorStartAndStop(t *testing.T) {
	mock := &mockBackend{
		name: "mock",
		vpnStatusFn: func(ctx context.Context) ([]types.VPNClientStatus, error) {
			return []types.VPNClientStatus{{CommonName: "user1"}}, nil
		},
		vpnSummaryFn: func(ctx context.Context) (*types.VPNSummary, error) {
			return &types.VPNSummary{NClients: 1}, nil
		},
		serviceStatusFn: func(ctx context.Context) (*types.ServiceStatus, error) {
			return &types.ServiceStatus{ServiceStatus: map[string]bool{"OPENVPN": true}}, nil
		},
	}

	c := newCollectorWithMock(mock)

	err := c.Start()
	require.NoError(t, err)

	// Wait for at least one collection cycle
	time.Sleep(100 * time.Millisecond)

	c.Stop()
}

func TestCollectorCollectAllSuccess(t *testing.T) {
	mock := &mockBackend{
		name: "mock",
		vpnStatusFn: func(ctx context.Context) ([]types.VPNClientStatus, error) {
			return []types.VPNClientStatus{
				{CommonName: "user1", Username: "user1", RealAddress: "1.2.3.4", VirtualAddress: "10.0.0.1"},
			}, nil
		},
		vpnSummaryFn: func(ctx context.Context) (*types.VPNSummary, error) {
			return &types.VPNSummary{NClients: 1, OvpnDcoAvailable: true, OvpnDcoVersion: "v2"}, nil
		},
		serviceStatusFn: func(ctx context.Context) (*types.ServiceStatus, error) {
			return &types.ServiceStatus{ServiceStatus: map[string]bool{"OPENVPN": true, "WEB": false}}, nil
		},
	}

	c := newCollectorWithMock(mock)
	c.collectAll(context.Background())
}

func TestCollectorCollectVPNSummaryError(t *testing.T) {
	mock := &mockBackend{
		name: "mock",
		vpnStatusFn: func(ctx context.Context) ([]types.VPNClientStatus, error) {
			return nil, nil
		},
		vpnSummaryFn: func(ctx context.Context) (*types.VPNSummary, error) {
			return nil, errors.New("sacli failed")
		},
		serviceStatusFn: func(ctx context.Context) (*types.ServiceStatus, error) {
			return nil, nil
		},
	}

	c := newCollectorWithMock(mock)
	// Should not panic, just log error
	c.collectAll(context.Background())
}

func TestCollectorCollectAllErrors(t *testing.T) {
	mockErr := errors.New("backend error")
	mock := &mockBackend{
		name:            "mock",
		vpnStatusFn:     func(ctx context.Context) ([]types.VPNClientStatus, error) { return nil, mockErr },
		vpnSummaryFn:    func(ctx context.Context) (*types.VPNSummary, error) { return nil, mockErr },
		serviceStatusFn: func(ctx context.Context) (*types.ServiceStatus, error) { return nil, mockErr },
	}

	c := newCollectorWithMock(mock)
	c.collectAll(context.Background())
}

func TestCollectorStopWithoutStart(t *testing.T) {
	mock := &mockBackend{name: "mock"}
	c := newCollectorWithMock(mock)
	// Stop without Start should not panic
	c.Stop()
}

func TestCollectorContextCancellation(t *testing.T) {
	mock := &mockBackend{
		name: "mock",
		vpnStatusFn: func(ctx context.Context) ([]types.VPNClientStatus, error) {
			return nil, nil
		},
		vpnSummaryFn: func(ctx context.Context) (*types.VPNSummary, error) {
			return nil, nil
		},
		serviceStatusFn: func(ctx context.Context) (*types.ServiceStatus, error) {
			return nil, nil
		},
	}

	c := newCollectorWithMock(mock)
	err := c.Start()
	require.NoError(t, err)

	// Stop quickly to test cancellation
	c.Stop()
}

func TestCollectorEmptyClients(t *testing.T) {
	mock := &mockBackend{
		name: "mock",
		vpnStatusFn: func(ctx context.Context) ([]types.VPNClientStatus, error) {
			return []types.VPNClientStatus{}, nil
		},
		vpnSummaryFn: func(ctx context.Context) (*types.VPNSummary, error) {
			return &types.VPNSummary{NClients: 0}, nil
		},
		serviceStatusFn: func(ctx context.Context) (*types.ServiceStatus, error) {
			return &types.ServiceStatus{ServiceStatus: map[string]bool{}}, nil
		},
	}

	c := newCollectorWithMock(mock)
	c.collectAll(context.Background())
}

func TestCollectorDefaultInterval(t *testing.T) {
	mock := &mockBackend{
		name: "mock",
		vpnStatusFn: func(ctx context.Context) ([]types.VPNClientStatus, error) {
			return nil, nil
		},
		vpnSummaryFn: func(ctx context.Context) (*types.VPNSummary, error) {
			return nil, nil
		},
		serviceStatusFn: func(ctx context.Context) (*types.ServiceStatus, error) {
			return nil, nil
		},
	}

	cfg := &config.Config{
		Collector: config.CollectorConfig{ScrapeInterval: 0}, // triggers default
		Backend: config.BackendConfig{
			Sacli: config.SacliConfig{Path: "/bin/true", Timeout: 1 * time.Second},
		},
	}
	reg := metrics.NewRegistry()
	c := &Collector{
		cfg:      cfg,
		registry: reg,
		backend:  mock,
	}

	err := c.Start()
	require.NoError(t, err)
	// Give it a moment for initial collection
	time.Sleep(50 * time.Millisecond)
	c.Stop()
}

func TestCollectorRecordScrapeMetrics(t *testing.T) {
	mock := &mockBackend{
		name: "mock",
		vpnStatusFn: func(ctx context.Context) ([]types.VPNClientStatus, error) {
			return []types.VPNClientStatus{{CommonName: "u1"}}, nil
		},
		vpnSummaryFn: func(ctx context.Context) (*types.VPNSummary, error) {
			return &types.VPNSummary{NClients: 1}, nil
		},
		serviceStatusFn: func(ctx context.Context) (*types.ServiceStatus, error) {
			return &types.ServiceStatus{ServiceStatus: map[string]bool{"X": true}}, nil
		},
	}

	c := newCollectorWithMock(mock)
	err := c.Start()
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)
	c.Stop()
}

func TestBackendFactory(t *testing.T) {
	cfg := &config.Config{
		Backend: config.BackendConfig{
			Sacli: config.SacliConfig{
				Path:    "/custom/sacli",
				Timeout: 5 * time.Second,
			},
		},
	}
	backend, err := createBackend(cfg)
	require.NoError(t, err)
	assert.Equal(t, "sacli", backend.Name())
}

func TestBackendFactoryEmptyPath(t *testing.T) {
	cfg := &config.Config{
		Backend: config.BackendConfig{
			Sacli: config.SacliConfig{},
		},
	}
	backend, err := createBackend(cfg)
	require.NoError(t, err)
	assert.Equal(t, "sacli", backend.Name())
}
