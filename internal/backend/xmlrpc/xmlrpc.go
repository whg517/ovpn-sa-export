package xmlrpc

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"time"

	alexejkXmlrpc "alexejk.io/go-xmlrpc"
	"github.com/whg517/ovpn-sa-export/internal/backend"
	"github.com/whg517/ovpn-sa-export/pkg/types"
)

// Config holds XML-RPC backend configuration.
type Config struct {
	Endpoint           string
	Username           string
	Password           string
	SocketPath         string
	Timeout            time.Duration
	InsecureSkipVerify bool
}

// Backend implements backend.CollectorBackend using XML-RPC API.
type Backend struct {
	client   *alexejkXmlrpc.Client
	endpoint string
}

// New creates a new XML-RPC backend.
func New(cfg Config) (*Backend, error) {
	if cfg.SocketPath != "" {
		return newSocketBackend(cfg)
	}
	return newHTTPBackend(cfg)
}

func newHTTPBackend(cfg Config) (*Backend, error) {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		return nil, fmt.Errorf("xmlrpc backend: endpoint is required")
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	// Add Basic Auth header if credentials provided
	var opts []alexejkXmlrpc.Option
	opts = append(opts, alexejkXmlrpc.HttpClient(httpClient))

	if cfg.Username != "" && cfg.Password != "" {
		authHeader := map[string]string{
			"Authorization": basicAuth(cfg.Username, cfg.Password),
		}
		opts = append(opts, alexejkXmlrpc.Headers(authHeader))
	}

	client, err := alexejkXmlrpc.NewClient(endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("create xmlrpc client: %w", err)
	}

	return &Backend{
		client:   client,
		endpoint: endpoint,
	}, nil
}

func newSocketBackend(cfg Config) (*Backend, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	dialer := func(_, _ string) (net.Conn, error) {
		return net.Dial("unix", cfg.SocketPath)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: dialer,
		},
		Timeout: timeout,
	}

	client, err := alexejkXmlrpc.NewClient("http://localhost/",
		alexejkXmlrpc.HttpClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("create xmlrpc socket client: %w", err)
	}

	return &Backend{
		client:   client,
		endpoint: cfg.SocketPath,
	}, nil
}

// Name returns the backend name.
func (b *Backend) Name() string { return "xmlrpc" }

// CollectVPNStatus retrieves VPN status via GetVPNStatus RPC call.
func (b *Backend) CollectVPNStatus(ctx context.Context) ([]types.VPNClientStatus, error) {
	var result VPNStatusResponse
	if err := b.client.Call("GetVPNStatus", nil, &result); err != nil {
		return nil, fmt.Errorf("xmlrpc GetVPNStatus: %w", err)
	}
	return parseVPNStatusResponse(result), nil
}

// CollectVPNSummary retrieves VPN summary via GetVPNSummary RPC call.
func (b *Backend) CollectVPNSummary(ctx context.Context) (*types.VPNSummary, error) {
	var result struct {
		VPNSummary VPNSummaryResponse `xml:"VPN_SUMMARY"`
	}
	if err := b.client.Call("GetVPNSummary", nil, &result); err != nil {
		return nil, fmt.Errorf("xmlrpc GetVPNSummary: %w", err)
	}
	return result.VPNSummary.toTypes(), nil
}

// CollectSubscriptionStatus retrieves subscription status via GetSubscriptionStatus RPC call.
func (b *Backend) CollectSubscriptionStatus(ctx context.Context) (*types.SubscriptionStatus, error) {
	var result SubscriptionStatusResponse
	if err := b.client.Call("GetSubscriptionStatus", nil, &result); err != nil {
		return nil, fmt.Errorf("xmlrpc GetSubscriptionStatus: %w", err)
	}
	return result.toTypes(), nil
}

// CollectServiceStatus retrieves service status via RunStatus RPC call.
func (b *Backend) CollectServiceStatus(ctx context.Context) (*types.ServiceStatus, error) {
	var result RunStatusResponse
	if err := b.client.Call("RunStatus", nil, &result); err != nil {
		return nil, fmt.Errorf("xmlrpc RunStatus: %w", err)
	}
	return result.toTypes(), nil
}

// Ensure Backend implements backend.CollectorBackend.
var _ backend.CollectorBackend = (*Backend)(nil)

func basicAuth(username, password string) string {
	creds := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return "Basic " + creds
}
