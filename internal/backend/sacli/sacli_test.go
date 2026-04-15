package sacli

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestParseVPNStatus(t *testing.T) {
	output := `OpenVPN daemon: openvpn0 (TCP)
CLIENT_LIST	Common Name	Real Address	Virtual Address	Virtual IPv6 Address	Bytes Received	Bytes Sent	Connected Since (time_t)	Username	Client ID	Peer ID
CLIENT_LIST	user1	1.2.3.4:55555	172.27.228.2	::1	123456	789012	1712000000	user1	1	0
CLIENT_LIST	user2	5.6.7.8:12345	172.27.228.3	::2	100	200	1712001000	user2	2	1
`
	clients, err := parseVPNStatus(output)
	require.NoError(t, err)
	require.Len(t, clients, 2)

	assert.Equal(t, "user1", clients[0].CommonName)
	assert.Equal(t, "1.2.3.4:55555", clients[0].RealAddress)
	assert.Equal(t, "172.27.228.2", clients[0].VirtualAddress)
	assert.Equal(t, "::1", clients[0].VirtualIPv6Addr)
	assert.Equal(t, int64(123456), clients[0].BytesReceived)
	assert.Equal(t, int64(789012), clients[0].BytesSent)
	assert.Equal(t, time.Unix(1712000000, 0), clients[0].ConnectedSince)
	assert.Equal(t, 1, clients[0].ClientID)
	assert.Equal(t, 0, clients[0].PeerID)

	assert.Equal(t, "user2", clients[1].CommonName)
	assert.Equal(t, int64(100), clients[1].BytesReceived)
}

func TestParseVPNStatusEmpty(t *testing.T) {
	clients, err := parseVPNStatus("")
	require.NoError(t, err)
	assert.Len(t, clients, 0)
}

func TestParseVPNSummary(t *testing.T) {
	output := `n_clients=5
ovpn_dco_available=true
ovpn_dco_ver=v2`
	s, err := parseVPNSummary(output)
	require.NoError(t, err)
	assert.Equal(t, 5, s.NClients)
	assert.True(t, s.OvpnDcoAvailable)
	assert.Equal(t, "v2", s.OvpnDcoVersion)
}

func TestParseVPNSummaryNoDCO(t *testing.T) {
	output := `n_clients=0
ovpn_dco_available=false
ovpn_dco_ver=`
	s, err := parseVPNSummary(output)
	require.NoError(t, err)
	assert.Equal(t, 0, s.NClients)
	assert.False(t, s.OvpnDcoAvailable)
}

func TestParseSubscriptionStatus(t *testing.T) {
	output := `current_cc=10
max_cc=50
fallback_cc=2
last_successful_update=1712000000
state=ACTIVE
`
	s, err := parseSubscriptionStatus(output)
	require.NoError(t, err)
	assert.Equal(t, 10, s.CurrentConnections)
	assert.Equal(t, 50, s.MaxConnections)
	assert.Equal(t, 2, s.FallbackConnections)
	assert.Equal(t, "ACTIVE", s.State)
	assert.Equal(t, time.Unix(1712000000, 0), s.LastSuccessfulUpdate)
}

func TestParseSubscriptionStatusEmpty(t *testing.T) {
	s, err := parseSubscriptionStatus("")
	require.NoError(t, err)
	assert.Equal(t, 0, s.CurrentConnections)
}

func TestParseServiceStatus(t *testing.T) {
	output := `OPENVPN: running
AUTH: running
WEB: stopped
AGENT: running
`
	s, err := parseServiceStatus(output)
	require.NoError(t, err)
	assert.True(t, s.Services["OPENVPN"])
	assert.True(t, s.Services["AUTH"])
	assert.False(t, s.Services["WEB"])
	assert.True(t, s.Services["AGENT"])
}

func TestParseServiceStatusEmpty(t *testing.T) {
	s, err := parseServiceStatus("")
	require.NoError(t, err)
	assert.Len(t, s.Services, 0)
}

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

func TestParseVPNStatusTooFewFields(t *testing.T) {
	output := "CLIENT_LIST\tA\tB"
	clients, err := parseVPNStatus(output)
	require.NoError(t, err)
	assert.Len(t, clients, 0)
}

func TestParseVPNStatusWhitespaceLines(t *testing.T) {
	output := "\n\nCLIENT_LIST\tCommon Name\tReal Address\tVirtual Address\tVirtual IPv6 Address\tBytes Received\tBytes Sent\tConnected Since (time_t)\tUsername\tClient ID\tPeer ID\nCLIENT_LIST\tu1\t1.1.1.1\t10.0.0.1\t::1\t100\t200\t1712000000\tu1\t1\t0\n"
	clients, err := parseVPNStatus(output)
	require.NoError(t, err)
	assert.Len(t, clients, 1)
}

func TestParseVPNStatusInvalidTimestamp(t *testing.T) {
	output := "CLIENT_LIST\tCommon Name\tReal Address\tVirtual Address\tVirtual IPv6 Address\tBytes Received\tBytes Sent\tConnected Since (time_t)\tUsername\tClient ID\tPeer ID\nCLIENT_LIST\tu1\t1.1.1.1\t10.0.0.1\t::1\t100\t200\tnotatimestamp\tu1\t1\t0\n"
	clients, err := parseVPNStatus(output)
	require.NoError(t, err)
	require.Len(t, clients, 1)
	assert.True(t, clients[0].ConnectedSince.IsZero())
}

func TestParseVPNSummaryEmpty(t *testing.T) {
	s, err := parseVPNSummary("")
	require.NoError(t, err)
	assert.Equal(t, 0, s.NClients)
	assert.False(t, s.OvpnDcoAvailable)
}

func TestParseSubscriptionStatusAlternateKeys(t *testing.T) {
	output := "CurrentCc=5\nMaxCc=25\nFallbackCc=1\nLastSuccessfulUpdate=1712000000\nState=ACTIVE\n"
	s, err := parseSubscriptionStatus(output)
	require.NoError(t, err)
	assert.Equal(t, 5, s.CurrentConnections)
	assert.Equal(t, 25, s.MaxConnections)
}

func TestParseSubscriptionStatusInvalidUpdate(t *testing.T) {
	output := "last_successful_update=notanumber\n"
	s, err := parseSubscriptionStatus(output)
	require.NoError(t, err)
	assert.True(t, s.LastSuccessfulUpdate.IsZero())
}

func TestCollectVPNStatusWithRunner(t *testing.T) {
	b := NewWithRunner(Config{}, func(ctx context.Context, path string, args ...string) (string, error) {
		assert.Equal(t, "/usr/local/openvpn_as/scripts/sacli", path)
		assert.Equal(t, []string{"VPNStatus"}, args)
		return "CLIENT_LIST\tCommon Name\tReal Address\tVirtual Address\tVirtual IPv6 Address\tBytes Received\tBytes Sent\tConnected Since (time_t)\tUsername\tClient ID\tPeer ID\nCLIENT_LIST\tu1\t1.1.1.1\t10.0.0.1\t::1\t100\t200\t1712000000\tu1\t1\t0\n", nil
	})
	clients, err := b.CollectVPNStatus(context.Background())
	require.NoError(t, err)
	require.Len(t, clients, 1)
	assert.Equal(t, "u1", clients[0].CommonName)
}

func TestCollectVPNStatusError(t *testing.T) {
	b := NewWithRunner(Config{}, func(ctx context.Context, path string, args ...string) (string, error) {
		return "", errors.New("command not found")
	})
	_, err := b.CollectVPNStatus(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sacli VPNStatus")
}

func TestCollectVPNSummaryWithRunner(t *testing.T) {
	b := NewWithRunner(Config{}, func(ctx context.Context, path string, args ...string) (string, error) {
		return "n_clients=3\n", nil
	})
	s, err := b.CollectVPNSummary(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 3, s.NClients)
}

func TestCollectVPNSummaryError(t *testing.T) {
	b := NewWithRunner(Config{}, func(ctx context.Context, path string, args ...string) (string, error) {
		return "", errors.New("fail")
	})
	_, err := b.CollectVPNSummary(context.Background())
	assert.Contains(t, err.Error(), "sacli VPNSummary")
}

func TestCollectSubscriptionStatusWithRunner(t *testing.T) {
	b := NewWithRunner(Config{}, func(ctx context.Context, path string, args ...string) (string, error) {
		return "current_cc=5\nmax_cc=50\n", nil
	})
	s, err := b.CollectSubscriptionStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 5, s.CurrentConnections)
}

func TestCollectSubscriptionStatusError(t *testing.T) {
	b := NewWithRunner(Config{}, func(ctx context.Context, path string, args ...string) (string, error) {
		return "", errors.New("fail")
	})
	_, err := b.CollectSubscriptionStatus(context.Background())
	assert.Contains(t, err.Error(), "sacli SubscriptionStatus")
}

func TestCollectServiceStatusWithRunner(t *testing.T) {
	b := NewWithRunner(Config{}, func(ctx context.Context, path string, args ...string) (string, error) {
		return "OPENVPN: running\n", nil
	})
	s, err := b.CollectServiceStatus(context.Background())
	require.NoError(t, err)
	assert.True(t, s.Services["OPENVPN"])
}

func TestCollectServiceStatusError(t *testing.T) {
	b := NewWithRunner(Config{}, func(ctx context.Context, path string, args ...string) (string, error) {
		return "", errors.New("fail")
	})
	_, err := b.CollectServiceStatus(context.Background())
	assert.Contains(t, err.Error(), "sacli status")
}

func TestCollectContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	b := NewWithRunner(Config{}, func(ctx context.Context, path string, args ...string) (string, error) {
		return "", ctx.Err()
	})
	_, err := b.CollectVPNStatus(ctx)
	assert.Error(t, err)
}
