.PHONY: build test test-race lint clean install coverage coverage-gate

BINARY := dockercomms
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X github.com/codethor0/dockercomms/internal/version.Version=$(VERSION) -X github.com/codethor0/dockercomms/internal/version.Commit=$(COMMIT) -X github.com/codethor0/dockercomms/internal/version.Date=$(DATE)"

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/dockercomms

install:
	go install ./cmd/dockercomms

test:
	go test ./...

test-race:
	go test -race ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

coverage-gate: coverage
	@go run ./internal/tools/covergate coverage.out

lint:
	go fmt ./...
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run ./...; fi

clean:
	rm -f $(BINARY)
	go clean -cache -testcache

all: lint test build
