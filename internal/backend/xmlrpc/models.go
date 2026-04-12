package xmlrpc

import (
	"time"

	"github.com/whg517/ovpn-sa-export/pkg/types"
)

// VPNStatusResponse represents the XML-RPC response for GetVPNStatus.
type VPNStatusResponse struct {
	Daemons []DaemonStatus `xml:"item"`
}

// DaemonStatus represents a single OpenVPN daemon status.
type DaemonStatus struct {
	Name         string       `xml:"name"`
	ClientList   []ClientList `xml:"client_list>item"`
	ClientHeader ClientHeader `xml:"client_list_headers"`
}

// ClientHeader maps header names to field indices.
type ClientHeader struct {
	CommonName     int `xml:"Common_Name"`
	RealAddress    int `xml:"Real_Address"`
	VirtualAddress int `xml:"Virtual_Address"`
	VirtualIPv6    int `xml:"Virtual_IPv6_Address"`
	BytesReceived  int `xml:"Bytes_Received"`
	BytesSent      int `xml:"Bytes_Sent"`
	ConnectedSince int `xml:"Connected_Since_time_t"`
	Username       int `xml:"Username"`
	ClientID       int `xml:"Client_ID"`
	PeerID         int `xml:"Peer_ID"`
}

// ClientList represents a single client entry.
type ClientList struct {
	Fields []string `xml:"item"`
}

func parseVPNStatusResponse(resp VPNStatusResponse) []types.VPNClientStatus {
	var clients []types.VPNClientStatus
	for _, daemon := range resp.Daemons {
		for _, cl := range daemon.ClientList {
			client := types.VPNClientStatus{}
			h := daemon.ClientHeader
			if h.CommonName < len(cl.Fields) {
				client.CommonName = cl.Fields[h.CommonName]
			}
			if h.Username < len(cl.Fields) {
				client.Username = cl.Fields[h.Username]
			}
			if h.RealAddress < len(cl.Fields) {
				client.RealAddress = cl.Fields[h.RealAddress]
			}
			if h.VirtualAddress < len(cl.Fields) {
				client.VirtualAddress = cl.Fields[h.VirtualAddress]
			}
			if h.VirtualIPv6 < len(cl.Fields) {
				client.VirtualIPv6Addr = cl.Fields[h.VirtualIPv6]
			}
			if h.BytesReceived < len(cl.Fields) {
				client.BytesReceived = parseBytes(cl.Fields[h.BytesReceived])
			}
			if h.BytesSent < len(cl.Fields) {
				client.BytesSent = parseBytes(cl.Fields[h.BytesSent])
			}
			if h.ClientID < len(cl.Fields) {
				client.ClientID = parseInt(cl.Fields[h.ClientID])
			}
			if h.PeerID < len(cl.Fields) {
				client.PeerID = parseInt(cl.Fields[h.PeerID])
			}
			if h.ConnectedSince < len(cl.Fields) {
				if sec := parseInt(cl.Fields[h.ConnectedSince]); sec > 0 {
					client.ConnectedSince = time.Unix(int64(sec), 0)
				}
			}
			clients = append(clients, client)
		}
	}
	return clients
}

// VPNSummaryResponse represents the XML-RPC response for GetVPNSummary.
type VPNSummaryResponse struct {
	NClients        int    `xml:"n_clients"`
	OvpnDcoVersion  string `xml:"ovpn_dco_ver"`
	OvpnDcoAvailable bool   `xml:"ovpn_dco_available"`
}

func (r VPNSummaryResponse) toTypes() *types.VPNSummary {
	return &types.VPNSummary{
		NClients:         r.NClients,
		OvpnDcoVersion:   r.OvpnDcoVersion,
		OvpnDcoAvailable: r.OvpnDcoAvailable,
	}
}

// SubscriptionStatusResponse represents the XML-RPC response for GetSubscriptionStatus.
type SubscriptionStatusResponse struct {
	CurrentConnections  int    `xml:"current_cc"`
	MaxConnections      int    `xml:"max_cc"`
	FallbackConnections int    `xml:"fallback_cc"`
	LastSuccessfulUpdate int   `xml:"last_successful_update"`
	State               string `xml:"state"`
}

func (r SubscriptionStatusResponse) toTypes() *types.SubscriptionStatus {
	s := &types.SubscriptionStatus{
		CurrentConnections:  r.CurrentConnections,
		MaxConnections:      r.MaxConnections,
		FallbackConnections: r.FallbackConnections,
		State:               r.State,
	}
	if r.LastSuccessfulUpdate > 0 {
		s.LastSuccessfulUpdate = time.Unix(int64(r.LastSuccessfulUpdate), 0)
	}
	return s
}

// RunStatusResponse represents the XML-RPC response for RunStatus.
type RunStatusResponse struct {
	Services []ServiceEntry `xml:"item"`
}

// ServiceEntry represents a single service status.
type ServiceEntry struct {
	Name    string `xml:"name"`
	Running bool   `xml:"running"`
}

func (r RunStatusResponse) toTypes() *types.ServiceStatus {
	services := make(map[string]bool)
	for _, s := range r.Services {
		services[s.Name] = s.Running
	}
	return &types.ServiceStatus{Services: services}
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

func parseBytes(s string) int64 {
	return int64(parseInt(s))
}
