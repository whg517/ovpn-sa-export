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
go build -o openvpn-as-exporter ./cmd/openvpn-as-exporter
```

### Run

```bash
# Default: load config from ./openvpn-as-exporter.yaml or /etc/openvpn-as-exporter/
sudo ./openvpn-as-exporter

# Specify config file
sudo ./openvpn-as-exporter -config /path/to/openvpn-as-exporter.yaml

# All defaults, no config file needed
sudo ./openvpn-as-exporter
```

### Docker

```bash
docker run -d \
  --name openvpn-as-exporter \
  -v /usr/local/openvpn_as/scripts:/sacli:ro \
  -p 9176:9176 \
  ghcr.io/whg517/openvpn-as-exporter:latest \
  -config /path/to/config.yaml
```

## Startup Arguments

| Flag | Default | Description |
|------|---------|-------------|
| `-version` | — | Print version, commit, build time, Go version and exit |
| `-config` | — | Path to YAML config file. If not specified, all defaults are used (no file search). |
| `-h` | — | Show help |

## Configuration

Configuration sources (highest priority first):

1. **Environment variables** (prefix `OPENVPN_AS_EXPORTER_`, nested keys use `_`):
   - `OPENVPN_AS_EXPORTER_SERVER_LISTEN_ADDRESS=:8080`
   - `OPENVPN_AS_EXPORTER_BACKEND_SACLI_TIMEOUT=15s`
   - `OPENVPN_AS_EXPORTER_COLLECTOR_SCRAPE_INTERVAL=30s`
2. **Config file** (YAML): `-config /path/to/file.yaml`
3. **Built-in defaults** (see below)

### Config File Example

See [`configs/openvpn-as-exporter.yaml`](configs/openvpn-as-exporter.yaml) for a full example:

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
| `backend.sacli.path` | `/usr/local/openvpn_as/scripts/sacli` | `OPENVPN_AS_EXPORTER_BACKEND_SACLI_PATH` |
| `backend.sacli.timeout` | `10s` | `OPENVPN_AS_EXPORTER_BACKEND_SACLI_TIMEOUT` |
| `server.listen_address` | `:9176` | `OPENVPN_AS_EXPORTER_SERVER_LISTEN_ADDRESS` |
| `server.metrics_path` | `/metrics` | `OPENVPN_AS_EXPORTER_SERVER_METRICS_PATH` |
| `server.health_path` | `/health` | `OPENVPN_AS_EXPORTER_SERVER_HEALTH_PATH` |
| `server.ready_path` | `/ready` | `OPENVPN_AS_EXPORTER_SERVER_READY_PATH` |
| `collector.scrape_interval` | `15s` | `OPENVPN_AS_EXPORTER_COLLECTOR_SCRAPE_INTERVAL` |
| `collector.timeout` | `30s` | `OPENVPN_AS_EXPORTER_COLLECTOR_TIMEOUT` |
| `collector.cache_ttl` | `60s` | `OPENVPN_AS_EXPORTER_COLLECTOR_CACHE_TTL` |
| `collector.enabled_collectors` | `vpn_status, vpn_summary, subscription, service` | — |
| `log.level` | `info` | `OPENVPN_AS_EXPORTER_LOG_LEVEL` |
| `log.format` | `json` | `OPENVPN_AS_EXPORTER_LOG_FORMAT` |

## Version

```bash
./openvpn-as-exporter -version
# openvpn-as-exporter 0.1.0-alpha.1
#   commit:     abc1234
#   build time: 2026-04-16T02:50:00Z
#   go:         go1.24.4
```

## Exposed Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `openvpn_as_exporter_up` | gauge | — | Last scrape success (1=success, 0=failure) |
| `openvpn_as_exporter_connected_clients` | gauge | — | Online client count |
| `openvpn_as_exporter_client_bytes_received` | gauge | username, common_name, real_addr, virtual_addr | Bytes received from client |
| `openvpn_as_exporter_client_bytes_sent` | gauge | username, common_name, real_addr, virtual_addr | Bytes sent to client |
| `openvpn_as_exporter_client_connected_since` | gauge | username, common_name | Connection start timestamp (Unix) |
| `openvpn_as_exporter_subscription_current_connections` | gauge | — | Current license connections used |
| `openvpn_as_exporter_subscription_max_connections` | gauge | — | Max allowed connections |
| `openvpn_as_exporter_subscription_fallback_connections` | gauge | — | Fallback connections used |
| `openvpn_as_exporter_service_up` | gauge | service | Service running status (1=running, 0=stopped) |
| `openvpn_as_exporter_scrape_duration_seconds` | histogram | backend | Scrape duration |
| `openvpn_as_exporter_scrape_errors_total` | counter | backend, collector | Total scrape errors |

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'openvpn_as_exporter'
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
