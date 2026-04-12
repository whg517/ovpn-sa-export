# OvpnSaExport

OpenVPN Access Server Prometheus Metrics Exporter — expose sacli/XML-RPC data as standard Prometheus metrics.

## Overview

OvpnSaExport collects VPN connection status, traffic statistics, license info, and service health from OpenVPN Access Server, then exposes them as Prometheus metrics for Grafana dashboards and Alertmanager alerts.

## Features

- **Dual Backend**: Local `sacli` command or remote XML-RPC API
- **Per-Client Metrics**: Username, IP addresses, bytes sent/received, connection duration
- **License Monitoring**: Current vs max connections, fallback usage
- **Service Health**: Internal AS service status (auth, agent, web, etc.)
- **Self-Observability**: Scrape duration, error counts
- **Multi-Instance**: Monitor multiple AS nodes with instance labels
- **Zero Dependencies**: Single static binary, no runtime dependencies

## Quick Start

### Binary

```bash
# Build
go build -o ovpn-sa-export ./cmd/ovpn-sa-export

# Run with sacli backend (default)
sudo ./ovpn-sa-export

# Run with XML-RPC backend
./ovpn-sa-export --backend.mode=xmlrpc --backend.xmlrpc.endpoint=https://vpn.example.com/RPC2/
```

### Docker

```bash
docker compose -f deployments/docker-compose.yaml up -d
```

## Configuration

Configuration is loaded from (in order of priority):

1. Command-line arguments (prefix: `--`)
2. Environment variables (prefix: `OVPN_SA_`)
3. Config file `/etc/ovpn-sa-export/ovpn-sa-export.yaml`

See [configs/ovpn-sa-export.yaml](configs/ovpn-sa-export.yaml) for all options.

## Exposed Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ovpn_sa_export_up` | gauge | — | Last scrape success |
| `ovpn_sa_export_connected_clients` | gauge | — | Online client count |
| `ovpn_sa_export_client_bytes_received` | gauge | username, common_name, real_addr, virtual_addr | Bytes from client |
| `ovpn_sa_export_client_bytes_sent` | gauge | username, common_name, real_addr, virtual_addr | Bytes to client |
| `ovpn_sa_export_client_connected_since` | gauge | username, common_name | Connection timestamp |
| `ovpn_sa_export_subscription_current_connections` | gauge | — | Current license usage |
| `ovpn_sa_export_subscription_max_connections` | gauge | — | Max allowed connections |
| `ovpn_sa_export_service_up` | gauge | service | Service running status |

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'ovpn_sa_export'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:9176']
```

## Endpoints

- `:9176/metrics` — Prometheus metrics
- `:9176/health` — Health check (always 200)
- `:9176/ready` — Readiness check

## License

MIT
