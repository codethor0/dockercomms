# DockerComms

OCI-native secure file transport CLI. Push and pull files as OCI artifacts with signing and verification.

## Prerequisites

- Go 1.23+
- OCI registry (e.g. ghcr.io, Docker Hub, GCR)
- Registry credentials (docker config or env)
- Cosign v3 (for signing; keyless OIDC expected)

## Build

```bash
go build ./cmd/dockercomms
# or
make build
```

## Run

```bash
./dockercomms --help
./dockercomms send --help
./dockercomms recv --help
./dockercomms verify --help
./dockercomms ack --help
```

## Test

```bash
go test ./...
go test -race ./...
make test
make test-race
make coverage-gate
```

CI: `.github/workflows/ci.yml` enforces build, test, race, lint, coverage-gate.

Integration tests (opt-in, skip when creds missing):

```bash
DOCKERCOMMS_IT_GHCR_REPO=ghcr.io/user/repo DOCKERCOMMS_IT_RECIPIENT=alice@example.com go test -tags=integration ./test/integration/...
```

## Exit Codes

- 0: success
- 1: generic failure
- 2: verification failed
- 3: registry auth/permission error
- 4: protocol/format error
- 5: not found

## Security Model

- Verify-before-materialize: payload is never written until verification succeeds
- Defenses: path traversal, zip/tar bombs, resource exhaustion
- Constant-time comparisons where applicable

## Implementation Notes

### Registry Compatibility

- Tag listing is the primary discovery mechanism (universal support)
- Referrers API is optional; used only for related artifacts (bundle/receipt), not message discovery
- If referrers returns 404/unsupported, fallback to tag-based bundle lookup
- oras-go/v2 (stable) is used; v3 is dev line. Docker config for auth: DOCKER_CONFIG or ~/.docker/config.json

### Docker Hub Fallback

- Docker Hub may have different pagination or rate limits
- Tag listing is unordered and eventually consistent; deduplicate by message id and digest
- For Docker Hub (docker.io), use full repo path: docker.io/username/repo

## Development

- SPEC.md: protocol specification
- ARCH.md: implementation architecture
- RELEASE_CHECKLIST.md: stop-ship gates
- .cursor/rules/dockercomms.mdc: Cursor rules for implementation constraints
- golangci-lint: `golangci-lint run ./...` (errcheck enabled)
