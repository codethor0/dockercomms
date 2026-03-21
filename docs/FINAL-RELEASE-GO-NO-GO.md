# Final release go / no-go checklist

## Current status line (quote as-is)

> DockerComms is locally end-to-end verified on Mac and green on all current non-credentialed gates. GitHub setup is complete subject to UI confirmation of private vulnerability reporting and optional release immutability. Final production-style end-to-end verification is still pending one successful authenticated GHCR run.

## Already satisfied (no action unless regressing)

- Local correctness: loopback plain HTTP fix; local `registry:2` E2E (text, binary, empty, medium, wrong recipient, path sanitization, repeat send).
- Repo gates: `make build`, `go test ./...`, `go test -race ./...`, `golangci-lint`, `make coverage-gate`, integration script `--check` modes.
- GitHub in repo: `SECURITY.md`, Dependabot, CodeQL workflow, secret scanning (API), branch protection, issue templates, PR template, docs (including `docs/repro.md` local registry section, `docs/RELEASE-RUNBOOK.md`, `docs/GITHUB-SECURITY-SETUP.md`, `docs/GITHUB-UI-FINAL-CHECKLIST.md`).

## Remaining before calling it fully proven

### A. GitHub UI (manual)

- [ ] **Private vulnerability reporting** enabled — see [GITHUB-UI-FINAL-CHECKLIST.md](GITHUB-UI-FINAL-CHECKLIST.md)
- [ ] **Release immutability** enabled if the option exists for this repo/account

### B. Authenticated GHCR (PAT with `read:packages` and `write:packages`; SSO authorized if required)

Run from repo root after `docker login ghcr.io` with that PAT:

- [ ] `./scripts/login-and-run-integration.sh` — completes without skip; smoke test passes
- [ ] `./scripts/docker-e2e.sh integration` — passes
- [ ] `./scripts/docker-e2e.sh cli` — passes
- [ ] (Optional) `./scripts/docker-e2e.sh full` — full harness

Exact env and commands: [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md)

## Go for final tag (e.g. v1.0.0)

**GO** when all are true:

- Status line above still accurate or updated after GHCR run
- Section A checked (or explicitly waived with reason)
- Section B required rows checked; optional `full` per your bar
- `main` CI and CodeQL green on the release commit
- No open stop-ship issues you care about

**NO-GO** if:

- GHCR integration or Docker CLI E2E fails with valid package-scoped auth, until root cause is fixed or scope is reduced (e.g. stay on RC).

## After GO

- Tag `v1.0.0` (or next final), publish GitHub Release, note in release body: local E2E + gates + live GHCR verification when applicable.
