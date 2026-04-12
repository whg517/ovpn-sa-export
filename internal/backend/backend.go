package backend

import (
	"context"

	"github.com/whg517/ovpn-sa-export/pkg/types"
)

// CollectorBackend defines the interface for data collection backends.
type CollectorBackend interface {
	// Name returns the backend name (e.g., "sacli", "xmlrpc").
	Name() string

	// CollectVPNStatus retrieves detailed per-client VPN connection status.
	CollectVPNStatus(ctx context.Context) ([]types.VPNClientStatus, error)

	// CollectVPNSummary retrieves the VPN summary (connected clients, DCO info).
	CollectVPNSummary(ctx context.Context) (*types.VPNSummary, error)

	// CollectSubscriptionStatus retrieves the license/subscription state.
	CollectSubscriptionStatus(ctx context.Context) (*types.SubscriptionStatus, error)

	// CollectServiceStatus retrieves internal AS service statuses.
	CollectServiceStatus(ctx context.Context) (*types.ServiceStatus, error)
}
