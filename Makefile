BINARY_NAME=ovpn-sa-export
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_TIME?=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOVERSION?=$(shell go version | awk '{print $$3}')
GO?=go
GOFLAGS?=-ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME) -X main.goVersion=$(GOVERSION)"

.PHONY: all build test lint fmt vet clean docker-build help

all: lint test build

## build: Build the binary
build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/ovpn-sa-export

## test: Run all tests
test:
	$(GO) test -race -count=1 -v ./...

## lint: Run go vet
lint: vet
	@echo "No external linter configured. Install golangci-lint for deeper analysis."

## vet: Run go vet
vet:
	$(GO) vet ./...

## fmt: Format code with gofmt
fmt:
	gofmt -s -w .

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)

## docker-build: Build Docker image
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) -f deployments/Dockerfile .

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
