package types

import "time"

// VPNClientStatus represents a single connected VPN client.
type VPNClientStatus struct {
	CommonName      string
	Username        string
	RealAddress     string
	VirtualAddress  string
	VirtualIPv6Addr string
	BytesReceived   int64
	BytesSent       int64
	ConnectedSince  time.Time
	ClientID        int
	PeerID          int
	Cipher          string
}

// VPNSummary represents the overall VPN status summary.
type VPNSummary struct {
	NClients        int
	OvpnDcoVersion  string
	OvpnDcoAvailable bool
}

// ServiceStatus represents the status of internal AS services.
type ServiceStatus struct {
	Services map[string]bool // service_name -> running
}
