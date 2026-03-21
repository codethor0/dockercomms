# DockerComms

OCI-native secure file transport CLI. Push and pull files as OCI artifacts with signing and verification.

## What is DockerComms?

DockerComms is an OCI-native secure file transport CLI.

Instead of building a new protocol, it uses standard OCI registries (GHCR, Docker Hub, GCR, etc.) to push and pull encrypted, signed payloads as OCI artifacts. It is designed for:

- Environments where HTTP(S) access to registries already exists
- Teams that need verify-before-materialize semantics
- Operators who want strong guarantees against path traversal and archive bombs

Key properties:

- **Verify-before-materialize:** payloads are never written to the final destination until verification succeeds.
- **Strict path and filename sanitization** (no `../`, no absolute paths, no backslashes).
- **Clear exit-code taxonomy** so scripts and CI can distinguish auth, verify, not-found, and generic failures.

## Quickstart for Contributors

### How to run tests (local gates)

All standard gates are runnable locally without any registry credentials.

```bash
# From repo root

# Unit tests
go test ./...

# Race detector
go test -race ./...

# Lint (golangci-lint)
golangci-lint run ./...

# Coverage gate (enforces minimum coverage thresholds)
make coverage-gate

# Script preflight checks (no network calls)
./scripts/run-integration.sh --check
./scripts/login-and-run-integration.sh --check
```

All of these should pass before you open a PR.

### How to run GHCR integration safely (optional)

Live GHCR tests are optional and must be run with your own credentials. They are not required for normal development or CI.

**Prerequisites:**

- Docker daemon running (Docker Desktop or equivalent)
- A GitHub PAT with `read:packages` and `write:packages`
- A GHCR repo you control, for example:

```bash
export GH_USER="your-gh-username"
export GH_PAT="ghp_...your_token..."
export DOCKERCOMMS_IT_GHCR_REPO="ghcr.io/${GH_USER}/dockercomms-it"
export DOCKERCOMMS_IT_RECIPIENT="team-b"
```

**To run integration tests safely:**

```bash
cd /path/to/dockercomms

# Optional: check scripts without hitting the network
./scripts/run-integration.sh --check
./scripts/login-and-run-integration.sh --check

# Login and run Go integration tests against GHCR
./scripts/login-and-run-integration.sh
```

**Security notes:**

- PAT is read from `GH_PAT` or `~/.dockercomms_gh_pat`.
- Scripts use `set -euo pipefail` and `umask 077` when touching secret files.
- PAT is never echoed or logged; there is no `set -x` around secret handling.
- `scripts/purge-ghcr-creds.sh` removes only GHCR-related Docker credentials if you need to recover from a bad login.

### Dockerized E2E harness (optional)

There is a Docker harness that runs the same gates inside a container and can exercise integration and CLI E2E flows.

```bash
# From repo root
./scripts/docker-e2e.sh gates        # build + tests + race + lint + coverage inside Docker
./scripts/docker-e2e.sh integration  # GHCR integration (requires GH_* env and login)
./scripts/docker-e2e.sh cli          # CLI send/recv/verify E2E flows
./scripts/docker-e2e.sh full         # gates + integration + CLI
```

Use these when you want extra assurance that DockerComms behaves the same way in a clean container as it does on your host.

---

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
- [SECURITY.md](SECURITY.md): vulnerability reporting (do not use public issues for security bugs)
- docs/GITHUB-SECURITY-SETUP.md: GitHub Security settings checklist for maintainers
- docs/GITHUB-UI-FINAL-CHECKLIST.md: short UI closeout (private reporting, optional immutability)
- .cursor/rules/dockercomms.mdc: Cursor rules for implementation constraints
- golangci-lint: `golangci-lint run ./...` (errcheck enabled)
