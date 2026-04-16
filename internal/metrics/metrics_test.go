package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/whg517/ovpn-sa-export/pkg/types"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	require.NotNil(t, r)
	assert.NotZero(t, testutil.CollectAndCount(r.up))
}

func TestRecordScrape(t *testing.T) {
	r := NewRegistry()
	r.RecordScrape("vpn_status", 50*time.Millisecond, nil)

	count := testutil.ToFloat64(r.scrapeTotal.WithLabelValues("vpn_status"))
	assert.Equal(t, 1.0, count)

	errCount := testutil.ToFloat64(r.scrapeErrors.WithLabelValues("vpn_status"))
	assert.Equal(t, 0.0, errCount)
}

func TestRecordScrapeError(t *testing.T) {
	r := NewRegistry()
	r.RecordScrape("vpn_status", 100*time.Millisecond, assert.AnError)

	errCount := testutil.ToFloat64(r.scrapeErrors.WithLabelValues("vpn_status"))
	assert.Equal(t, 1.0, errCount)
}

func TestUpdateVPNStatus(t *testing.T) {
	r := NewRegistry()
	clients := []types.VPNClientStatus{
		{
			Username:       "user1",
			CommonName:     "user1",
			RealAddress:    "1.2.3.4:5555",
			VirtualAddress: "10.0.0.1",
			BytesReceived:  1000,
			BytesSent:      2000,
			ConnectedSince: time.Unix(1712000000, 0),
		},
	}
	r.UpdateVPNStatus(clients)

	assert.Equal(t, 1.0, testutil.ToFloat64(r.connectedClients))
	assert.Equal(t, 1.0, testutil.ToFloat64(r.up))

	sent := testutil.ToFloat64(r.clientBytesSent.WithLabelValues("user1", "user1", "1.2.3.4:5555", "10.0.0.1"))
	assert.Equal(t, 2000.0, sent)
}

func TestUpdateVPNSummary(t *testing.T) {
	r := NewRegistry()
	r.UpdateVPNSummary(&types.VPNSummary{NClients: 5, OvpnDcoAvailable: true, OvpnDcoVersion: "v2"})

	assert.Equal(t, 5.0, testutil.ToFloat64(r.connectedClients))
	assert.Equal(t, 1.0, testutil.ToFloat64(r.dcoAvailable))
}

func TestUpdateServiceStatus(t *testing.T) {
	r := NewRegistry()
	r.UpdateServiceStatus(&types.ServiceStatus{
		ServiceStatus: map[string]bool{"openvpn_0": true, "web": false},
	})

	assert.Equal(t, 1.0, testutil.ToFloat64(r.serviceUp.WithLabelValues("openvpn_0")))
	assert.Equal(t, 0.0, testutil.ToFloat64(r.serviceUp.WithLabelValues("web")))
}

func TestUpdateNil(t *testing.T) {
	r := NewRegistry()
	// Should not panic
	r.UpdateVPNSummary(nil)
	r.UpdateServiceStatus(nil)
}
