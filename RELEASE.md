# DockerComms v1.0.0-rc3

DockerComms is an OCI-native secure file transport CLI: it pushes and pulls files as OCI artifacts with strong verification, strict tagging rules, and a "verify-before-materialize" model so untrusted payloads never touch disk until they pass verification.

This rc3 release finishes the security hardening and CI/gating work for v1.0.0 and adds a safe, repeatable way to run GHCR integration and Docker-based E2E tests.

---

## Highlights

### 1. Exit code taxonomy hardened for all CLI flows

The CLI now uses a consistent exit code mapping across `send`, `recv`, and `verify`:

- `0` – success  
- `2` – verification failed  
- `3` – registry auth/permission error (401/403, "denied", "unauthorized")  
- `5` – not found / missing bundle  
- `1` – other/unclassified errors  

`send`, `recv`, and `verify` all share the same classifier so auth errors and not-found cases behave predictably in automation and scripts.

### 2. Stronger filename sanitization (including Windows-style traversal)

`SanitizeFilename` already prevented basic `../`-style traversal and handled edge cases like `/` and `..`.

rc3 adds extra defenses:

- Windows-style traversal inputs like `..\evil.txt` and `..\\..\\evil.txt` are normalized and have backslashes stripped.
- Sanitized filenames are guaranteed to:
  - Be non-empty
  - Not equal `.` or `..`
  - Contain no `/` or `\`

This behavior is covered by unit tests and fuzz tests:
- `TestSanitizeFilename` includes Unix and Windows-style edge cases.
- `FuzzSanitizeFilename_NoTraversal` asserts the output never contains path separators.

### 3. GHCR integration runner and Docker E2E harness

To keep the codebase clean but still allow realistic registry tests, rc3 adds a set of scripts and docs that operators can use locally:

- `scripts/login-and-run-integration.sh`
- `scripts/run-integration.sh`
- `scripts/purge-ghcr-creds.sh`
- `scripts/docker-e2e.sh`
- Updated `docs/repro.md`

Key properties:

- `set -euo pipefail` everywhere
- `umask 077` when handling secret files
- PAT is read from `GH_PAT` or `~/.dockercomms_gh_pat` and is **never** echoed
- `--check` modes validate the environment and show what would run, without hitting GHCR
- Docker harness supports `gates`, `integration`, `cli`, and `full` modes

These scripts are opt-in for operators and not required by CI.

---

## Quality & Gates

All of the following gates pass at v1.0.0-rc3:

- `go test ./...`
- `go test -race ./...`
- `golangci-lint run ./...`
- `make coverage-gate`
- `./scripts/run-integration.sh --check`
- `./scripts/login-and-run-integration.sh --check`
- `./scripts/docker-e2e.sh gates`

Negative-case and safety tests:

- `TestVerifyBeforeMaterialize_NoOutputOnFailure`
- `TestVerifyBeforeMaterialize_NoTempLeftover`
- `TestSanitizeFilename` (including `/`, `..`, `../`, Windows-style `..\`, long names)
- `FuzzSanitizeFilename_NoTraversal`
- Exit-code classifier tests for `recv`, `verify`, and `send`

The result: under current test, fuzz, and script coverage, there are no known correctness or safety bugs at this tag.

---

## How to run GHCR integration tests (optional, local)

These are **not** required for normal use and are not run in CI, but if you want live evidence against GHCR:

```bash
# One-time config in your shell
export GH_USER="codethor0"
export GH_PAT="ghp_...your_token..."
export DOCKERCOMMS_IT_GHCR_REPO="ghcr.io/${GH_USER}/dockercomms-it"
export DOCKERCOMMS_IT_RECIPIENT="team-b"

cd /path/to/dockercomms

# Preflight / check-only (no network)
./scripts/run-integration.sh --check
./scripts/login-and-run-integration.sh --check

# Run host integration tests against GHCR
./scripts/login-and-run-integration.sh

# Dockerized gates (build + test + race + lint + coverage inside a container)
./scripts/docker-e2e.sh gates

# Optional: Dockerized integration + CLI flows
./scripts/docker-e2e.sh integration
./scripts/docker-e2e.sh cli
./scripts/docker-e2e.sh full
```

The scripts will:

- Use `docker login ghcr.io` non-interactively with `--password-stdin`
- Never print the PAT
- Skip Docker Hub or large-payload tests unless corresponding env vars are set

---

## Known limitations / future work

- **Live GHCR/Docker Hub tests are opt-in:** They require an operator PAT and are not run by default in CI.
- **Large payload E2E:** A ~256 MiB payload path exists behind `DOCKERCOMMS_IT_LARGE_PAYLOAD=1`. Running that test requires sufficient bandwidth and GHCR quota and is intended for local/operator use.
- **Resume/HEAD behavior:** Resume logic (HEAD-before-upload) is exercised by integration tests; a mock-based unit test could be added in a future version for an even tighter contract.

If you're testing rc3 and see anything off (exit codes, tag behavior, path handling, or registry quirks), please open an issue with the exact command + environment and we can tighten the spec further.
