# OvpnSaExport

OpenVPN Access Server Prometheus Metrics Exporter — expose sacli data as standard Prometheus metrics.

## Features

- **sacli Backend**: Collects metrics via local `sacli` commands
- **Per-Client Metrics**: Username, IP addresses, bytes sent/received, connection duration
- **License Monitoring**: Current vs max connections, fallback usage
- **Service Health**: Internal AS service status (auth, agent, web, etc.)
- **Self-Observability**: Scrape duration, error counts
- **Zero Dependencies**: Single static binary, no runtime dependencies

## Quick Start

### Build

```bash
make build
# or
go build -o ovpn-sa-export ./cmd/ovpn-sa-export
```

### Run

```bash
# Default: load config from ./ovpn-sa-export.yaml or /etc/ovpn-sa-export/
sudo ./ovpn-sa-export

# Specify config file
sudo ./ovpn-sa-export -config /path/to/ovpn-sa-export.yaml

# All defaults, no config file needed
sudo ./ovpn-sa-export
```

### Docker

```bash
docker run -d \
  --name ovpn-sa-export \
  -v /usr/local/openvpn_as/scripts:/sacli:ro \
  -p 9176:9176 \
  ghcr.io/whg517/ovpn-sa-export:latest \
  -config /path/to/config.yaml
```

## Startup Arguments

| Flag | Default | Description |
|------|---------|-------------|
| `-version` | — | Print version, commit, build time, Go version and exit |
| `-config` | auto-detect | Path to YAML config file. If not specified, searches `./ovpn-sa-export.yaml` → `/etc/ovpn-sa-export/` → `~/.ovpn-sa-export/`. If no config found, all defaults are used. |
| `-h` | — | Show help |

## Configuration

Configuration sources (highest priority first):

1. **Environment variables** (prefix `OVPN_SA_`, nested keys use `_`):
   - `OVPN_SA_SERVER_LISTEN_ADDRESS=:8080`
   - `OVPN_SA_BACKEND_SACLI_TIMEOUT=15s`
   - `OVPN_SA_COLLECTOR_SCRAPE_INTERVAL=30s`
2. **Config file** (YAML): `-config /path/to/file.yaml`
3. **Built-in defaults** (see below)

### Config File Example

See [`configs/ovpn-sa-export.yaml`](configs/ovpn-sa-export.yaml) for a full example:

```yaml
# sacli backend configuration
backend:
  sacli:
    path: /usr/local/openvpn_as/scripts/sacli
    timeout: 10s

# HTTP server
server:
  listen_address: ":9176"
  metrics_path: /metrics
  health_path: /health
  ready_path: /ready

# Collector
collector:
  scrape_interval: 15s
  timeout: 30s
  cache_ttl: 60s
  enabled_collectors:
    - vpn_status
    - vpn_summary
    - subscription
    - service

# Logging
log:
  level: info
  format: json
```

### Default Values

No config file is required. All settings have sensible defaults:

| Setting | Default | Env Var |
|---------|---------|---------|
| `backend.sacli.path` | `/usr/local/openvpn_as/scripts/sacli` | `OVPN_SA_BACKEND_SACLI_PATH` |
| `backend.sacli.timeout` | `10s` | `OVPN_SA_BACKEND_SACLI_TIMEOUT` |
| `server.listen_address` | `:9176` | `OVPN_SA_SERVER_LISTEN_ADDRESS` |
| `server.metrics_path` | `/metrics` | `OVPN_SA_SERVER_METRICS_PATH` |
| `server.health_path` | `/health` | `OVPN_SA_SERVER_HEALTH_PATH` |
| `server.ready_path` | `/ready` | `OVPN_SA_SERVER_READY_PATH` |
| `collector.scrape_interval` | `15s` | `OVPN_SA_COLLECTOR_SCRAPE_INTERVAL` |
| `collector.timeout` | `30s` | `OVPN_SA_COLLECTOR_TIMEOUT` |
| `collector.cache_ttl` | `60s` | `OVPN_SA_COLLECTOR_CACHE_TTL` |
| `collector.enabled_collectors` | `vpn_status, vpn_summary, subscription, service` | — |
| `log.level` | `info` | `OVPN_SA_LOG_LEVEL` |
| `log.format` | `json` | `OVPN_SA_LOG_FORMAT` |

## Version

```bash
./ovpn-sa-export -version
# ovpn-sa-export 0.1.0-alpha.1
#   commit:     abc1234
#   build time: 2026-04-16T02:50:00Z
#   go:         go1.24.4
```

## Exposed Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ovpn_sa_export_up` | gauge | — | Last scrape success (1=success, 0=failure) |
| `ovpn_sa_export_connected_clients` | gauge | — | Online client count |
| `ovpn_sa_export_client_bytes_received` | gauge | username, common_name, real_addr, virtual_addr | Bytes received from client |
| `ovpn_sa_export_client_bytes_sent` | gauge | username, common_name, real_addr, virtual_addr | Bytes sent to client |
| `ovpn_sa_export_client_connected_since` | gauge | username, common_name | Connection start timestamp (Unix) |
| `ovpn_sa_export_subscription_current_connections` | gauge | — | Current license connections used |
| `ovpn_sa_export_subscription_max_connections` | gauge | — | Max allowed connections |
| `ovpn_sa_export_subscription_fallback_connections` | gauge | — | Fallback connections used |
| `ovpn_sa_export_service_up` | gauge | service | Service running status (1=running, 0=stopped) |
| `ovpn_sa_export_scrape_duration_seconds` | histogram | backend | Scrape duration |
| `ovpn_sa_export_scrape_errors_total` | counter | backend, collector | Total scrape errors |

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
