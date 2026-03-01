# Reproducible Proof Runs

Exact commands for validating DockerComms against real registries. No marketing language.

## Gate Outputs (paste from local run)

```
go version go1.26.0 darwin/arm64

go test ./...
?   	github.com/dockercomms/dockercomms/cmd/dockercomms	[no test files]
?   	github.com/dockercomms/dockercomms/internal/tools/covergate	[no test files]
ok  	github.com/dockercomms/dockercomms/internal/version
?   	github.com/dockercomms/dockercomms/pkg/cli	[no test files]
?   	github.com/dockercomms/dockercomms/pkg/config	[no test files]
ok  	github.com/dockercomms/dockercomms/pkg/crypto
ok  	github.com/dockercomms/dockercomms/pkg/oci
ok  	github.com/dockercomms/dockercomms/pkg/transfer

go test -race ./...
(all packages ok)

golangci-lint run ./...
0 issues.

make coverage-gate
github.com/dockercomms/dockercomms/pkg/crypto: 66.7% OK
github.com/dockercomms/dockercomms/pkg/transfer: 36.1% OK
github.com/dockercomms/dockercomms/pkg/oci: 54.1% OK

go test -run Test -tags=integration ./test/integration/...
ok  	github.com/dockercomms/dockercomms/test/integration
```

## Prerequisites

- `dockercomms` binary (from `make build`)
- Registry credentials: `docker login ghcr.io` or `docker login` for Docker Hub
- Env vars set per section below

## GHCR Round-Trip (send -> recv -> verify)

```bash
# Required env
export DOCKERCOMMS_IT_GHCR_REPO=ghcr.io/OWNER/REPO
export DOCKERCOMMS_IT_RECIPIENT=alice@example.com

# Optional: output directory (default: temp dir)
export DOCKERCOMMS_IT_OUTDIR=/tmp/dockercomms-out

# Create test file
echo "proof run" > /tmp/proof.txt

# Send
./dockercomms send --repo $DOCKERCOMMS_IT_GHCR_REPO --recipient $DOCKERCOMMS_IT_RECIPIENT /tmp/proof.txt

# Expected: exit 0, digest printed

# Recv
./dockercomms recv --repo $DOCKERCOMMS_IT_GHCR_REPO --me $DOCKERCOMMS_IT_RECIPIENT --out $DOCKERCOMMS_IT_OUTDIR --max 1

# Expected: exit 0, "Received N message(s)"

# Verify (use digest from send output)
./dockercomms verify --repo $DOCKERCOMMS_IT_GHCR_REPO --digest sha256:XXXX

# Expected: exit 0, "verified: sha256:XXXX"
```

## Docker Hub Tag Listing

```bash
# Required env (use full path)
export DOCKERCOMMS_IT_DH_REPO=docker.io/USERNAME/REPO

# Run integration test
go test -tags=integration -run TestDockerHubTagListing -v ./test/integration/...

# Expected: exit 0, tag count logged
```

## Integration Tests (opt-in)

Integration tests are behind `//go:build integration` and skip when env is missing.

```bash
# GHCR round-trip (skips if DOCKERCOMMS_IT_GHCR_REPO or DOCKERCOMMS_IT_RECIPIENT unset)
DOCKERCOMMS_IT_GHCR_REPO=ghcr.io/user/repo DOCKERCOMMS_IT_RECIPIENT=alice@example.com \
  go test -tags=integration -run TestGHCRRoundTrip -v ./test/integration/...

# Docker Hub tag listing (skips if DOCKERCOMMS_IT_DH_REPO unset)
DOCKERCOMMS_IT_DH_REPO=docker.io/user/repo \
  go test -tags=integration -run TestDockerHubTagListing -v ./test/integration/...
```

## Success Criteria

- send: exit 0, digest and tag printed
- recv: exit 0, file materialized in out dir
- verify: exit 0, "verified" message
- Integration tests: exit 0 when env set; skip when env unset

## Module Path

Current: `github.com/dockercomms/dockercomms`. If publishing under a different path (e.g. `github.com/codethor0/dockercomms`):

```bash
go mod edit -module=github.com/codethor0/dockercomms
# Then update all import paths in the codebase
```
