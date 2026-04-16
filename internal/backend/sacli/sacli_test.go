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

	// user1
	assert.Equal(t, "user1", clients[0].CommonName)
	assert.Equal(t, "1.2.3.4:55555", clients[0].RealAddress)
	assert.Equal(t, "172.27.228.2", clients[0].VirtualAddress)
	assert.Equal(t, "::1", clients[0].VirtualIPv6Addr)
	assert.Equal(t, int64(123456), clients[0].BytesReceived)
	assert.Equal(t, int64(789012), clients[0].BytesSent)
	assert.Equal(t, time.Unix(1712000000, 0), clients[0].ConnectedSince)
	assert.Equal(t, 1, clients[0].ClientID)
	assert.Equal(t, 0, clients[0].PeerID)

	// user2
	assert.Equal(t, "user2", clients[1].CommonName)
	assert.Equal(t, int64(100), clients[1].BytesReceived)
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
}

func TestParseVPNStatus_TooFewFields(t *testing.T) {
	clients, err := parseVPNStatus(fixture(t, "vpn_status_too_few_fields.txt"))
	require.NoError(t, err)
	assert.Len(t, clients, 0)
}

func TestParseVPNStatus_WhitespaceLines(t *testing.T) {
	clients, err := parseVPNStatus(fixture(t, "vpn_status_whitespace.txt"))
	require.NoError(t, err)
	assert.Len(t, clients, 1)
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

// --- SubscriptionStatus Parser Tests ---

func TestParseSubscriptionStatus(t *testing.T) {
	s, err := parseSubscriptionStatus(fixture(t, "subscription_status.txt"))
	require.NoError(t, err)
	assert.Equal(t, 10, s.CurrentConnections)
	assert.Equal(t, 50, s.MaxConnections)
	assert.Equal(t, 2, s.FallbackConnections)
	assert.Equal(t, "ACTIVE", s.State)
	assert.Equal(t, time.Unix(1712000000, 0), s.LastSuccessfulUpdate)
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

// --- Helper Function Tests ---

func TestGetFieldValue(t *testing.T) {
	fields := []string{"A", "B", "C"}
	headerMap := map[string]int{"A": 0, "B": 1, "C": 2}

	assert.Equal(t, "A", getFieldValue(fields, headerMap, "A"))
	assert.Equal(t, "B", getFieldValue(fields, headerMap, "X", "B"))
	assert.Equal(t, "", getFieldValue(fields, headerMap, "NONEXISTENT"))
}

func TestParseInt(t *testing.T) {
	assert.Equal(t, 42, parseInt("42"))
	assert.Equal(t, 0, parseInt("abc"))
	assert.Equal(t, 0, parseInt(""))
}

func TestParseInt64(t *testing.T) {
	assert.Equal(t, int64(42), parseInt64("42"))
	assert.Equal(t, int64(0), parseInt64("abc"))
}

// --- Integration Tests: Collect* with FixtureRunner ---

func TestCollectVPNStatus_Integration(t *testing.T) {
	b := NewWithRunner(Config{}, fixtureRunner(t, "vpn_status_multiple.txt"))
	clients, err := b.CollectVPNStatus(context.Background())
	require.NoError(t, err)
	require.Len(t, clients, 2)
	assert.Equal(t, "user1", clients[0].CommonName)
	assert.Equal(t, "user2", clients[1].CommonName)
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
