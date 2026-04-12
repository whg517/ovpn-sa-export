package sacli

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/whg517/ovpn-sa-export/internal/backend"
	"github.com/whg517/ovpn-sa-export/pkg/types"
)

// Config holds sacli backend configuration.
type Config struct {
	Path    string
	Timeout time.Duration
}

// Backend implements backend.CollectorBackend using local sacli commands.
type Backend struct {
	path    string
	timeout time.Duration
}

// New creates a new sacli backend.
func New(cfg Config) *Backend {
	path := cfg.Path
	if path == "" {
		path = "/usr/local/openvpn_as/scripts/sacli"
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Backend{path: path, timeout: timeout}
}

// Name returns the backend name.
func (b *Backend) Name() string { return "sacli" }

// CollectVPNStatus retrieves per-client VPN status via sacli VPNStatus.
func (b *Backend) CollectVPNStatus(ctx context.Context) ([]types.VPNClientStatus, error) {
	output, err := b.run(ctx, "VPNStatus")
	if err != nil {
		return nil, fmt.Errorf("sacli VPNStatus: %w", err)
	}
	return parseVPNStatus(output)
}

// CollectVPNSummary retrieves VPN summary via sacli VPNSummary.
func (b *Backend) CollectVPNSummary(ctx context.Context) (*types.VPNSummary, error) {
	output, err := b.run(ctx, "VPNSummary")
	if err != nil {
		return nil, fmt.Errorf("sacli VPNSummary: %w", err)
	}
	return parseVPNSummary(output)
}

// CollectSubscriptionStatus retrieves subscription status via sacli SubscriptionStatus.
func (b *Backend) CollectSubscriptionStatus(ctx context.Context) (*types.SubscriptionStatus, error) {
	output, err := b.run(ctx, "SubscriptionStatus")
	if err != nil {
		return nil, fmt.Errorf("sacli SubscriptionStatus: %w", err)
	}
	return parseSubscriptionStatus(output)
}

// CollectServiceStatus retrieves service status via sacli status.
func (b *Backend) CollectServiceStatus(ctx context.Context) (*types.ServiceStatus, error) {
	output, err := b.run(ctx, "status")
	if err != nil {
		return nil, fmt.Errorf("sacli status: %w", err)
	}
	return parseServiceStatus(output)
}

func (b *Backend) run(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	allArgs := append([]string{}, args...)
	cmd := exec.CommandContext(ctx, b.path, allArgs...)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Ensure Backend implements backend.CollectorBackend.
var _ backend.CollectorBackend = (*Backend)(nil)

// --- Parsers ---

func parseVPNStatus(output string) ([]types.VPNClientStatus, error) {
	// sacli VPNStatus output format:
	//   OpenVPN daemon: openvpn0 (TCP)
	//   CLIENT_LIST   Common Name    Real Address    Virtual Address    ...
	//   CLIENT_LIST   user1    1.2.3.4:55555    172.27.228.2    ...
	var clients []types.VPNClientStatus
	lines := strings.Split(output, "\n")

	var headers []string
	var headerIdx = -1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "CLIENT_LIST") {
			fields := strings.Split(line, "	")
			if len(fields) < 11 {
				continue
			}

			if headerIdx < 0 {
				// First CLIENT_LIST is the header row
				headers = fields
				headerIdx = 0
				continue
			}

			// Build header-to-index mapping
			headerMap := make(map[string]int)
			for i, h := range headers {
				headerMap[h] = i
			}

			client := types.VPNClientStatus{
				CommonName:      getFieldValue(fields, headerMap, "Common Name", "COMMON_NAME"),
				Username:        getFieldValue(fields, headerMap, "Username", "USERNAME"),
				RealAddress:     getFieldValue(fields, headerMap, "Real Address", "REAL_ADDRESS"),
				VirtualAddress:  getFieldValue(fields, headerMap, "Virtual Address", "VIRTUAL_ADDRESS"),
				VirtualIPv6Addr: getFieldValue(fields, headerMap, "Virtual IPv6 Address", "VIRTUAL_IPV6_ADDRESS"),
				BytesReceived:   parseInt64(getFieldValue(fields, headerMap, "Bytes Received", "BYTES_RECEIVED")),
				BytesSent:       parseInt64(getFieldValue(fields, headerMap, "Bytes Sent", "BYTES_SENT")),
				ClientID:        parseInt(getFieldValue(fields, headerMap, "Client ID", "CLIENT_ID")),
				PeerID:          parseInt(getFieldValue(fields, headerMap, "Peer ID", "PEER_ID")),
			}

			if ts := getFieldValue(fields, headerMap, "Connected Since (time_t)", "CONNECTED_SINCE"); ts != "" {
				if sec, err := strconv.ParseInt(ts, 10, 64); err == nil {
					client.ConnectedSince = time.Unix(sec, 0)
				}
			}

			clients = append(clients, client)
		}
	}

	return clients, nil
}

func getFieldValue(fields []string, headerMap map[string]int, keys ...string) string {
	for _, key := range keys {
		if idx, ok := headerMap[key]; ok && idx < len(fields) {
			return fields[idx]
		}
	}
	return ""
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func parseInt64(s string) int64 {
	return int64(parseInt(s))
}

func parseVPNSummary(output string) (*types.VPNSummary, error) {
	summary := &types.VPNSummary{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "n_clients") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				summary.NClients = parseInt(parts[1])
			}
		}
		if strings.HasPrefix(line, "ovpn_dco_available") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				summary.OvpnDcoAvailable = parts[1] == "true"
			}
		}
		if strings.HasPrefix(line, "ovpn_dco_ver") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				summary.OvpnDcoVersion = parts[1]
			}
		}
	}
	return summary, nil
}

func parseSubscriptionStatus(output string) (*types.SubscriptionStatus, error) {
	status := &types.SubscriptionStatus{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "current_cc", "CurrentCc":
			status.CurrentConnections = parseInt(parts[1])
		case "max_cc", "MaxCc":
			status.MaxConnections = parseInt(parts[1])
		case "fallback_cc", "FallbackCc":
			status.FallbackConnections = parseInt(parts[1])
		case "last_successful_update", "LastSuccessfulUpdate":
			if sec, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				status.LastSuccessfulUpdate = time.Unix(sec, 0)
			}
		case "state", "State":
			status.State = parts[1]
		}
	}
	return status, nil
}

func parseServiceStatus(output string) (*types.ServiceStatus, error) {
	services := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		// sacli status output: "service_name: running" or "service_name: stopped"
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			state := strings.TrimSpace(parts[1])
			services[name] = strings.Contains(strings.ToLower(state), "running")
		}
	}
	return &types.ServiceStatus{Services: services}, nil
}
