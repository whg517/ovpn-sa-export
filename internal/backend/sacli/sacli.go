package sacli

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/whg517/ovpn-sa-export/pkg/types"
)

// Config holds sacli backend configuration.
type Config struct {
	Path    string
	Timeout time.Duration
}

// RunFunc is a function that executes a command and returns its output.
type RunFunc func(ctx context.Context, path string, args ...string) (string, error)

// Backend implements backend.CollectorBackend using local sacli commands.
type Backend struct {
	path    string
	timeout time.Duration
	runFn   RunFunc
}

// New creates a new sacli backend.
func New(cfg Config) *Backend {
	return NewWithRunner(cfg, nil)
}

// NewWithRunner creates a new sacli backend with a custom run function (for testing).
func NewWithRunner(cfg Config, runFn RunFunc) *Backend {
	path := cfg.Path
	if path == "" {
		path = "/usr/local/openvpn_as/scripts/sacli"
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	if runFn == nil {
		runFn = defaultRun
	}
	return &Backend{path: path, timeout: timeout, runFn: runFn}
}

func defaultRun(ctx context.Context, path string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
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
	return b.runFn(ctx, b.path, args...)
}

// toJSON converts sacli's Python dict output (single quotes, True/False/None) to valid JSON.
func toJSON(input string) string {
	s := strings.ReplaceAll(input, "'", "\"")
	s = strings.ReplaceAll(s, "True", "true")
	s = strings.ReplaceAll(s, "False", "false")
	s = strings.ReplaceAll(s, "None", "null")
	return s
}

// vpnStatusResponse is the JSON output of "sacli VPNStatus".
type vpnStatusResponse map[string]vpnDaemonStatus

type vpnDaemonStatus struct {
	ClientList      [][]interface{}          `json:"client_list"`
	ClientListHeader map[string]float64      `json:"client_list_header"`
	GlobalStats     map[string]string       `json:"global_stats"`
	RoutingTable    [][]interface{}          `json:"routing_table"`
	RoutingTableHeader map[string]float64   `json:"routing_table_header"`
	Time            []string                `json:"time"`
	Title           string                  `json:"title"`
}

// --- Parsers ---

func parseVPNStatus(output string) ([]types.VPNClientStatus, error) {
	var resp vpnStatusResponse
	if err := json.Unmarshal([]byte(toJSON(output)), &resp); err != nil {
		return nil, fmt.Errorf("parse VPNStatus JSON: %w", err)
	}

	var clients []types.VPNClientStatus

	for _, daemon := range resp {
		header := daemon.ClientListHeader
		for _, row := range daemon.ClientList {
			if len(row) < 11 {
				continue
			}
			client := types.VPNClientStatus{
				CommonName:     jsonStr(row, header["Common Name"]),
				RealAddress:    jsonStr(row, header["Real Address"]),
				VirtualAddress: jsonStr(row, header["Virtual Address"]),
				BytesReceived:  jsonInt64(row, header["Bytes Received"]),
				BytesSent:      jsonInt64(row, header["Bytes Sent"]),
				ClientID:       int(jsonFloat(row, header["Client ID"])),
				PeerID:         int(jsonFloat(row, header["Peer ID"])),
			}

			if idx, ok := header["Connected Since (time_t)"]; ok {
				if sec, err := jsonInt64At(row, int(idx)); err == nil {
					client.ConnectedSince = time.Unix(sec, 0)
				}
			}

			if idx, ok := header["Username"]; ok {
				client.Username = jsonStrAt(row, int(idx))
			}

			if idx, ok := header["Virtual IPv6 Address"]; ok {
				client.VirtualIPv6Addr = jsonStrAt(row, int(idx))
			}

			if idx, ok := header["Data Channel Cipher"]; ok {
				client.Cipher = jsonStrAt(row, int(idx))
			}

			clients = append(clients, client)
		}
	}

	return clients, nil
}

func parseVPNSummary(output string) (*types.VPNSummary, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(toJSON(output)), &raw); err != nil {
		return nil, fmt.Errorf("parse VPNSummary JSON: %w", err)
	}

	summary := &types.VPNSummary{}
	if v, ok := raw["n_clients"]; ok {
		summary.NClients = int(jsonToFloat64(v))
	}
	if v, ok := raw["ovpn_dco_available"]; ok {
		summary.OvpnDcoAvailable = v == true || v == "true"
	}
	if v, ok := raw["ovpn_dco_ver"]; ok {
		summary.OvpnDcoVersion = fmt.Sprintf("%v", v)
	}
	return summary, nil
}

func parseSubscriptionStatus(output string) (*types.SubscriptionStatus, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(toJSON(output)), &raw); err != nil {
		return nil, fmt.Errorf("parse SubscriptionStatus JSON: %w", err)
	}

	status := &types.SubscriptionStatus{}
	if v, ok := raw["current_cc"]; ok {
		status.CurrentConnections = int(jsonToFloat64(v))
	}
	if v, ok := raw["max_cc"]; ok {
		status.MaxConnections = int(jsonToFloat64(v))
	}
	if v, ok := raw["fallback_cc"]; ok {
		status.FallbackConnections = int(jsonToFloat64(v))
	}
	if v, ok := raw["state"]; ok {
		status.State = fmt.Sprintf("%v", v)
	}
	if v, ok := raw["last_successful_update"]; ok {
		if sec := int64(jsonToFloat64(v)); sec > 0 {
			status.LastSuccessfulUpdate = time.Unix(sec, 0)
		}
	}
	return status, nil
}

func parseServiceStatus(output string) (*types.ServiceStatus, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(toJSON(output)), &raw); err != nil {
		return nil, fmt.Errorf("parse status JSON: %w", err)
	}

	services := make(map[string]bool)
	for name, val := range raw {
		// sacli status returns values like "running", "stopped"
		state := fmt.Sprintf("%v", val)
		services[name] = state == "running"
	}
	return &types.ServiceStatus{Services: services}, nil
}

// --- JSON helpers ---

func jsonStr(row []interface{}, idx float64) string {
	return jsonStrAt(row, int(idx))
}

func jsonStrAt(row []interface{}, idx int) string {
	if idx >= 0 && idx < len(row) {
		if s, ok := row[idx].(string); ok {
			return s
		}
	}
	return ""
}

func jsonInt64(row []interface{}, idx float64) int64 {
	v, _ := jsonInt64At(row, int(idx))
	return v
}

func jsonInt64At(row []interface{}, idx int) (int64, error) {
	if idx >= 0 && idx < len(row) {
		switch v := row[idx].(type) {
		case string:
			var n int64
			_, err := fmt.Sscanf(v, "%d", &n)
			return n, err
		case float64:
			return int64(v), nil
		}
	}
	return 0, fmt.Errorf("index %d out of range or invalid type", idx)
}

func jsonFloat(row []interface{}, idx float64) float64 {
	i := int(idx)
	if i >= 0 && i < len(row) {
		switch v := row[i].(type) {
		case float64:
			return v
		case string:
			var f float64
			fmt.Sscanf(v, "%f", &f)
			return f
		}
	}
	return 0
}

func jsonToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	case int:
		return float64(val)
	case int64:
		return float64(val)
	}
	return 0
}
