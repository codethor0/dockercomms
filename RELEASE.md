# Release Provenance

Reproducible build and verification steps for DockerComms releases.

## Toolchain

- Go: version pinned in `go.mod` (currently 1.23)
- CI uses `go-version-file: go.mod` for consistency

Verify local toolchain:

```bash
go version
# Must match go.mod directive
```

## Build

```bash
make build
./dockercomms version
```

## Gates (all must pass)

```bash
go test ./...
go test -race ./...
golangci-lint run ./...
make coverage-gate
```

## Integration Tests

Behind `//go:build integration`. Skip unless env vars are set. No secrets in code.

```bash
# GHCR round-trip
DOCKERCOMMS_IT_GHCR_REPO=ghcr.io/OWNER/REPO DOCKERCOMMS_IT_RECIPIENT=alice@example.com go test -tags=integration ./test/integration/...

# Docker Hub tag listing
DOCKERCOMMS_IT_DH_REPO=docker.io/USERNAME/REPO go test -tags=integration -run TestDockerHubTagListing ./test/integration/...
```

## Release Provenance Checklist

Before tagging or publishing:

- [ ] `git status --porcelain` is empty
- [ ] `git rev-parse --short HEAD` returns a hash (not "not a git repo")
- [ ] Tag exists: `git tag -l v1.0.0-rc1`
- [ ] All gates pass: `go test ./...`, `go test -race ./...`, `golangci-lint run ./...`, `make coverage-gate`
