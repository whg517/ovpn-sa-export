package sacli

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixture reads a testdata file and returns its content.
func fixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
	return string(data)
}

// fixtureRunner returns a RunFunc that returns the content of the named fixture.
func fixtureRunner(t *testing.T, name string) RunFunc {
	t.Helper()
	output := fixture(t, name)
	return func(_ context.Context, _ string, _ ...string) (string, error) {
		return output, nil
	}
}

// --- Constructor Tests ---

func TestNew(t *testing.T) {
	b := New(Config{Timeout: 5 * time.Second})
	assert.Equal(t, "sacli", b.Name())
	assert.Contains(t, b.path, "sacli")
	assert.Equal(t, 5*time.Second, b.timeout)
}

func TestNewDefaultPath(t *testing.T) {
	b := New(Config{})
	assert.Equal(t, "/usr/local/openvpn_as/scripts/sacli", b.path)
	assert.Equal(t, 10*time.Second, b.timeout)
}

// --- VPNStatus Parser Tests ---

func TestParseVPNStatus_MultipleClients(t *testing.T) {
	clients, err := parseVPNStatus(fixture(t, "vpn_status_multiple.txt"))
	require.NoError(t, err)
	require.Len(t, clients, 2)

	// kimi.fang
	assert.Equal(t, "kimi.fang", clients[0].CommonName)
	assert.Equal(t, "203.0.113.1:11991", clients[0].RealAddress)
	assert.Equal(t, "172.31.31.59", clients[0].VirtualAddress)
	assert.Equal(t, "", clients[0].VirtualIPv6Addr)
	assert.Equal(t, int64(4720051), clients[0].BytesReceived)
	assert.Equal(t, int64(15260167), clients[0].BytesSent)
	assert.Equal(t, time.Unix(1776316170, 0), clients[0].ConnectedSince)
	assert.Equal(t, "kimi.fang", clients[0].Username)
	assert.Equal(t, 8716, clients[0].ClientID)
	assert.Equal(t, 7, clients[0].PeerID)
	assert.Equal(t, "AES-256-GCM", clients[0].Cipher)

	// john.doe
	assert.Equal(t, "john.doe", clients[1].CommonName)
	assert.Equal(t, "203.0.113.2:15747", clients[1].RealAddress)
	assert.Equal(t, "172.31.31.34", clients[1].VirtualAddress)
	assert.Equal(t, int64(72997), clients[1].BytesReceived)
	assert.Equal(t, int64(350987), clients[1].BytesSent)
	assert.Equal(t, time.Unix(1776332601, 0), clients[1].ConnectedSince)
	assert.Equal(t, "AES-256-GCM", clients[1].Cipher)
}

func TestParseVPNStatus_Empty(t *testing.T) {
	clients, err := parseVPNStatus(fixture(t, "vpn_status_empty.txt"))
	require.NoError(t, err)
	assert.Len(t, clients, 0)
}

func TestParseVPNStatus_InvalidTimestamp(t *testing.T) {
	clients, err := parseVPNStatus(fixture(t, "vpn_status_invalid_timestamp.txt"))
	require.NoError(t, err)
	require.Len(t, clients, 1)
	assert.True(t, clients[0].ConnectedSince.IsZero())
	assert.Equal(t, "203.0.113.1:50000", clients[0].RealAddress)
}

func TestParseVPNStatus_InvalidJSON(t *testing.T) {
	_, err := parseVPNStatus("not json at all")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse VPNStatus JSON")
}

// --- VPNSummary Parser Tests ---

func TestParseVPNSummary_WithDCO(t *testing.T) {
	s, err := parseVPNSummary(fixture(t, "vpn_summary.txt"))
	require.NoError(t, err)
	assert.Equal(t, 5, s.NClients)
	assert.True(t, s.OvpnDcoAvailable)
	assert.Equal(t, "v2", s.OvpnDcoVersion)
}

func TestParseVPNSummary_NoDCO(t *testing.T) {
	s, err := parseVPNSummary(fixture(t, "vpn_summary_no_dco.txt"))
	require.NoError(t, err)
	assert.Equal(t, 0, s.NClients)
	assert.False(t, s.OvpnDcoAvailable)
}

func TestParseVPNSummary_Empty(t *testing.T) {
	s, err := parseVPNSummary(fixture(t, "empty.txt"))
	require.NoError(t, err)
	assert.Equal(t, 0, s.NClients)
	assert.False(t, s.OvpnDcoAvailable)
}

func TestParseVPNSummary_InvalidJSON(t *testing.T) {
	_, err := parseVPNSummary("not json")
	assert.Error(t, err)
}

// --- SubscriptionStatus Parser Tests ---

func TestParseSubscriptionStatus(t *testing.T) {
	s, err := parseSubscriptionStatus(fixture(t, "subscription_status.txt"))
	require.NoError(t, err)
	assert.Equal(t, 10, s.CurrentConnections)
	assert.Equal(t, 50, s.MaxConnections)
	assert.Equal(t, 2, s.FallbackConnections)
	assert.Equal(t, "ACTIVE", s.State)
	assert.Equal(t, time.Unix(1776316170, 0), s.LastSuccessfulUpdate)
}

func TestParseSubscriptionStatus_AlternateKeys(t *testing.T) {
	s, err := parseSubscriptionStatus(fixture(t, "subscription_status_alternate_keys.txt"))
	require.NoError(t, err)
	assert.Equal(t, 5, s.CurrentConnections)
	assert.Equal(t, 25, s.MaxConnections)
}

func TestParseSubscriptionStatus_InvalidUpdate(t *testing.T) {
	s, err := parseSubscriptionStatus(fixture(t, "subscription_status_invalid_update.txt"))
	require.NoError(t, err)
	assert.True(t, s.LastSuccessfulUpdate.IsZero())
}

func TestParseSubscriptionStatus_Empty(t *testing.T) {
	s, err := parseSubscriptionStatus(fixture(t, "empty.txt"))
	require.NoError(t, err)
	assert.Equal(t, 0, s.CurrentConnections)
}

func TestParseSubscriptionStatus_InvalidJSON(t *testing.T) {
	_, err := parseSubscriptionStatus("not json")
	assert.Error(t, err)
}

func TestParseSubscriptionStatus_NotConfigured(t *testing.T) {
	s, err := parseSubscriptionStatus(fixture(t, "subscription_status_not_configured.txt"))
	require.NoError(t, err)
	assert.Equal(t, "NOT_CONFIGURED", s.State)
	assert.Equal(t, 0, s.CurrentConnections)
}

func TestToJSON_PythonDict(t *testing.T) {
	input := "{'key': 'value', 'flag': True, 'none': None}"
	result := toJSON(input)
	assert.Equal(t, `{"key": "value", "flag": true, "none": null}`, result)
}

func TestToJSON_StandardJSON(t *testing.T) {
	input := `{"key": "value"}`
	result := toJSON(input)
	// Already double-quoted, single quote replacement is no-op
	assert.Equal(t, `{"key": "value"}`, result)
}

// --- ServiceStatus Parser Tests ---

func TestParseServiceStatus(t *testing.T) {
	s, err := parseServiceStatus(fixture(t, "service_status.txt"))
	require.NoError(t, err)
	assert.True(t, s.Services["OPENVPN"])
	assert.True(t, s.Services["AUTH"])
	assert.False(t, s.Services["WEB"])
	assert.True(t, s.Services["AGENT"])
}

func TestParseServiceStatus_Empty(t *testing.T) {
	s, err := parseServiceStatus(fixture(t, "empty.txt"))
	require.NoError(t, err)
	assert.Len(t, s.Services, 0)
}

func TestParseServiceStatus_InvalidJSON(t *testing.T) {
	_, err := parseServiceStatus("not json")
	assert.Error(t, err)
}

// --- Integration Tests: Collect* with FixtureRunner ---

func TestCollectVPNStatus_Integration(t *testing.T) {
	b := NewWithRunner(Config{}, fixtureRunner(t, "vpn_status_multiple.txt"))
	clients, err := b.CollectVPNStatus(context.Background())
	require.NoError(t, err)
	require.Len(t, clients, 2)
	assert.Equal(t, "kimi.fang", clients[0].CommonName)
	assert.Equal(t, "john.doe", clients[1].CommonName)
}

func TestCollectVPNSummary_Integration(t *testing.T) {
	b := NewWithRunner(Config{}, fixtureRunner(t, "vpn_summary.txt"))
	s, err := b.CollectVPNSummary(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 5, s.NClients)
	assert.True(t, s.OvpnDcoAvailable)
}

func TestCollectSubscriptionStatus_Integration(t *testing.T) {
	b := NewWithRunner(Config{}, fixtureRunner(t, "subscription_status.txt"))
	s, err := b.CollectSubscriptionStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 10, s.CurrentConnections)
	assert.Equal(t, 50, s.MaxConnections)
	assert.Equal(t, "ACTIVE", s.State)
}

func TestCollectServiceStatus_Integration(t *testing.T) {
	b := NewWithRunner(Config{}, fixtureRunner(t, "service_status.txt"))
	s, err := b.CollectServiceStatus(context.Background())
	require.NoError(t, err)
	assert.True(t, s.Services["OPENVPN"])
	assert.False(t, s.Services["WEB"])
}

// --- Error Path Tests ---

func TestCollectVPNStatus_Error(t *testing.T) {
	b := NewWithRunner(Config{}, func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("command not found")
	})
	_, err := b.CollectVPNStatus(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sacli VPNStatus")
}

func TestCollectVPNSummary_Error(t *testing.T) {
	b := NewWithRunner(Config{}, func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("fail")
	})
	_, err := b.CollectVPNSummary(context.Background())
	assert.Contains(t, err.Error(), "sacli VPNSummary")
}

func TestCollectSubscriptionStatus_Error(t *testing.T) {
	b := NewWithRunner(Config{}, func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("fail")
	})
	_, err := b.CollectSubscriptionStatus(context.Background())
	assert.Contains(t, err.Error(), "sacli SubscriptionStatus")
}

func TestCollectServiceStatus_Error(t *testing.T) {
	b := NewWithRunner(Config{}, func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("fail")
	})
	_, err := b.CollectServiceStatus(context.Background())
	assert.Contains(t, err.Error(), "sacli status")
}

func TestCollectContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	b := NewWithRunner(Config{}, func(ctx context.Context, _ string, _ ...string) (string, error) {
		return "", ctx.Err()
	})
	_, err := b.CollectVPNStatus(ctx)
	assert.Error(t, err)
}

// --- JSON Helper Tests ---

func TestJsonHelpers(t *testing.T) {
	row := []interface{}{"hello", "123", float64(456), "789"}
	header := map[string]float64{"Name": 0, "Count": 1, "Value": 2, "StrNum": 3}

	assert.Equal(t, "hello", jsonStr(row, header["Name"]))
	assert.Equal(t, int64(456), jsonInt64(row, header["Value"]))
	assert.Equal(t, float64(456), jsonFloat(row, header["Value"]))

	// String-as-number
	assert.Equal(t, int64(123), jsonInt64(row, header["Count"]))
	assert.Equal(t, int64(789), jsonInt64(row, header["StrNum"]))
}

func TestJsonHelpersOutOfRange(t *testing.T) {
	row := []interface{}{"a"}
	assert.Equal(t, "", jsonStr(row, 5))
	assert.Equal(t, int64(0), jsonInt64(row, 5))
}
