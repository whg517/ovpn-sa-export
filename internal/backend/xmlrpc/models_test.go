package xmlrpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVPNStatusResponse(t *testing.T) {
	resp := VPNStatusResponse{
		Daemons: []DaemonStatus{
			{
				Name: "openvpn0",
				ClientHeader: ClientHeader{
					CommonName: 0, RealAddress: 1, VirtualAddress: 2,
					VirtualIPv6: 3, BytesReceived: 4, BytesSent: 5,
					ConnectedSince: 6, Username: 7, ClientID: 8, PeerID: 9,
				},
				ClientList: []ClientList{
					{Fields: []string{"user1", "1.2.3.4:5555", "10.0.0.1", "::1", "100", "200", "1712000000", "user1", "1", "0"}},
					{Fields: []string{"user2", "5.6.7.8:1234", "10.0.0.2", "::2", "300", "400", "1712001000", "user2", "2", "1"}},
				},
			},
		},
	}

	clients := parseVPNStatusResponse(resp)
	require.Len(t, clients, 2)

	assert.Equal(t, "user1", clients[0].CommonName)
	assert.Equal(t, "1.2.3.4:5555", clients[0].RealAddress)
	assert.Equal(t, "10.0.0.1", clients[0].VirtualAddress)
	assert.Equal(t, int64(100), clients[0].BytesReceived)
	assert.Equal(t, int64(200), clients[0].BytesSent)
	assert.Equal(t, time.Unix(1712000000, 0), clients[0].ConnectedSince)
	assert.Equal(t, 1, clients[0].ClientID)

	assert.Equal(t, "user2", clients[1].CommonName)
	assert.Equal(t, int64(300), clients[1].BytesReceived)
}

func TestParseVPNStatusResponseEmpty(t *testing.T) {
	clients := parseVPNStatusResponse(VPNStatusResponse{})
	assert.Len(t, clients, 0)
}

func TestVPNSummaryToTypes(t *testing.T) {
	r := VPNSummaryResponse{NClients: 10, OvpnDcoAvailable: true, OvpnDcoVersion: "v3"}
	s := r.toTypes()
	assert.Equal(t, 10, s.NClients)
	assert.True(t, s.OvpnDcoAvailable)
	assert.Equal(t, "v3", s.OvpnDcoVersion)
}

func TestSubscriptionStatusToTypes(t *testing.T) {
	r := SubscriptionStatusResponse{
		CurrentConnections:  15,
		MaxConnections:      100,
		FallbackConnections: 3,
		LastSuccessfulUpdate: 1712000000,
		State:               "ACTIVE",
	}
	s := r.toTypes()
	assert.Equal(t, 15, s.CurrentConnections)
	assert.Equal(t, 100, s.MaxConnections)
	assert.Equal(t, 3, s.FallbackConnections)
	assert.Equal(t, time.Unix(1712000000, 0), s.LastSuccessfulUpdate)
	assert.Equal(t, "ACTIVE", s.State)
}

func TestSubscriptionStatusToTypesNoUpdate(t *testing.T) {
	r := SubscriptionStatusResponse{CurrentConnections: 5, MaxConnections: 10}
	s := r.toTypes()
	assert.Equal(t, 5, s.CurrentConnections)
	assert.True(t, s.LastSuccessfulUpdate.IsZero())
}

func TestRunStatusToTypes(t *testing.T) {
	r := RunStatusResponse{
		Services: []ServiceEntry{
			{Name: "OPENVPN", Running: true},
			{Name: "WEB", Running: false},
		},
	}
	s := r.toTypes()
	assert.True(t, s.Services["OPENVPN"])
	assert.False(t, s.Services["WEB"])
}
