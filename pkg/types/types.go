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
	NClients         int
	OvpnDcoAvailable bool
	OvpnDcoVersion   string
}

// ServiceStatus represents the sacli status output.
type ServiceStatus struct {
	ActiveProfile     string
	LastRestarted     string
	ServiceStatus     map[string]bool   // service_name -> on/off
	AuthModulesStatus map[string]bool   // auth_module -> enabled/disabled
	DCOAvailable      bool
	DCOVersion        string
	Errors            map[string]string
}
