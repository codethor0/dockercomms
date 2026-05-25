# Contributing

Thank you for improving DockerComms. This project targets a small, security-sensitive CLI; keep changes focused and covered by tests.

## Prerequisites

- Go version in `go.mod`
- `golangci-lint` (CI installs v2.10.1; match locally when possible)
- Optional: Docker for [docs/repro.md](docs/repro.md) and `scripts/docker-e2e.sh`

## Local checks (required before a PR)

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
golangci-lint run ./...
make coverage-gate
./scripts/run-integration.sh --check
./scripts/login-and-run-integration.sh --check
```

CI runs the same steps on Ubuntu via [.github/workflows/ci.yml](.github/workflows/ci.yml).

## Optional integration

GHCR integration tests are opt-in and need your own PAT and repository:

```bash
export GH_USER=your-username
export GH_PAT=ghp_...
export DOCKERCOMMS_IT_GHCR_REPO=ghcr.io/${GH_USER}/dockercomms-it
export DOCKERCOMMS_IT_RECIPIENT=team-b
./scripts/login-and-run-integration.sh
```

Never commit tokens or `.env` files.

## Docker E2E harness

```bash
./scripts/docker-e2e.sh gates
```

Requires a running Docker daemon.

## Security-sensitive changes

Do not weaken verify-before-materialize, filename sanitization, chunk limits, or exit-code semantics without discussion. Report vulnerabilities via [SECURITY.md](SECURITY.md), not public issues.

## Pull requests

Use the PR template checklist. Update user-facing docs when behavior changes.
