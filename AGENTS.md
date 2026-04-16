# AGENTS.md - OvpnSaExport

## Project Overview

OpenVPN Access Server Prometheus Metrics Exporter. Collects VPN metrics via local `sacli` commands and exposes them as Prometheus metrics.

## Tech Stack

- **Language:** Go 1.24+
- **Dependencies:** Go standard library + Prometheus client_golang + Viper + stretchr/testify
- **Build:** Makefile or `go build`
- **CI:** GitHub Actions (lint + test + release via GoReleaser)

## Project Structure

```
cmd/ovpn-sa-export/     # Entry point
internal/
  backend/sacli/         # sacli command execution and JSON parsing
  collector/             # Orchestrates collection and metric updates
  config/                # Configuration loading (Viper)
  metrics/               # Prometheus metric definitions and updates
  server/                # HTTP server (metrics/health/ready)
pkg/types/               # Shared type definitions
deployments/             # Dockerfiles
configs/                 # Example config files
testdata/                # Fixture files for tests (within respective packages)
```

## Development Workflow

### Before Every Commit

**MANDATORY:** All code must pass lint and tests before committing and pushing.

```bash
# Run lint (golangci-lint)
make lint

# Run all tests with race detection
make test

# Quick check: build + test
go build ./... && go test -race -count=1 ./...
```

### Commit Rules

1. **Never commit code that fails `make lint` or `make test`**
2. Run `make lint` first — golangci-lint catches unused returns, missing errors, etc.
3. Run `make test` — all tests must pass with `-race` flag
4. If CI fails on lint/test, fix it before pushing new code
5. Use conventional commit messages: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`, `ci:`

### Testing

- **Fixture-based tests:** All sacli parser tests use files in `internal/backend/sacli/testdata/`
- **Fixture data must be real sacli output** — do not fabricate test data
- **Integration tests:** Use `fixtureRunner()` to test Collect* methods end-to-end
- **Error paths:** Always test error scenarios (invalid JSON, missing fields, context cancellation)
- **No external dependencies in tests:** Use `RunFunc` injection, never call real `sacli` in tests

### sacli Output Format

**CRITICAL:** sacli outputs Python dict syntax, NOT standard JSON:
- Single quotes instead of double quotes: `{'key': 'value'}`
- Python booleans: `True`/`False` instead of `true`/`false`
- Python null: `None` instead of `null`

The `toJSON()` converter in `sacli.go` handles this. All new parsers must use it.

### Adding a New Collector

1. Add the type to `pkg/types/types.go`
2. Add the `Collect*` method to `Backend` interface in `collector.go`
3. Implement it in `internal/backend/sacli/sacli.go`
4. Add parser function with `toJSON()` conversion
5. Add real sacli output fixture to `testdata/`
6. Add parser tests + integration tests
7. Add Prometheus metrics in `internal/metrics/metrics.go`
8. Add the collector call in `collectAll()` in `collector.go`
9. Add the collector name to `config.go` defaults
10. Run `make lint && make test` — must pass before commit

## Key Design Decisions

- **No implicit config file search:** `-config` must be explicitly provided; defaults + env vars otherwise
- **No SubscriptionStatus:** Removed because it requires server-side subscription configuration that may not exist
- **Dockerfile split:** `Dockerfile` for GoReleaser (COPY pre-built binary), `Dockerfile.legacy` for CI (multi-stage build)
- **Release requires lint+test:** The release workflow runs lint and test as prerequisites

## Release Process

1. Ensure all tests pass locally: `make lint && make test`
2. Commit and push all changes
3. Tag: `git tag v0.1.0-beta.3 && git push origin v0.1.0-beta.3`
4. GitHub Actions runs: lint → test → GoReleaser (builds binaries + Docker image + creates GitHub Release)
5. Verify the release on GitHub before announcing

## Common Pitfalls

- `fmt.Sscanf` return value must be checked (errcheck lint rule)
- `git add -A` can pick up build artifacts — use `.gitignore` and review staged files
- sacli output varies between OpenVPN AS versions — always verify against real output
