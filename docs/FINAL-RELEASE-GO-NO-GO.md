# Final release go / no-go checklist

## Official split

### GitHub setup

**Complete**, pending:

- UI confirmation of **private vulnerability reporting**
- Optional **release immutability** (if the setting is available)

### Release proof still remaining

- Authenticated **GHCR** integration
- **Signed** send / `recv --verify=true`
- **Cryptographic verification failure** (bundle present, signature/policy check fails): exit **`2`** (distinct from **exit `5`** when no bundle exists, e.g. `recv --verify=true` after `--sign=false`)

## Status lines (quote as-is)

**GitHub setup:**

> GitHub setup is complete for DockerComms, subject only to UI confirmation of private vulnerability reporting and optional release immutability.

**Full product bar (after local + matrix work):**

> DockerComms is green on host/Docker gates and locally verified (macOS + Linux distro matrices). Final production-style proof is still pending authenticated GHCR and signed verification flows (including verify-with-bundle crypto failure → exit 2).

## Already satisfied (no action unless regressing)

- Local correctness: loopback + Docker DNS single-label plain HTTP; local `registry:2` E2E; six-distro Linux matrix (arm64 and amd64) per `docs/repro.md` / `scripts/linux-distro-matrix.sh`.
- Repo gates: `make build`, `go test ./...`, `go test -race ./...`, `golangci-lint`, `make coverage-gate`, integration script `--check` modes.
- GitHub in repo: `SECURITY.md`, Dependabot, CodeQL workflow, secret scanning (API), branch protection, issue templates, PR template, docs (including `docs/repro.md`, `docs/RELEASE-RUNBOOK.md`, `docs/GITHUB-SECURITY-SETUP.md`, `docs/GITHUB-UI-FINAL-CHECKLIST.md`).

## Remaining before calling it fully proven

### A. GitHub UI (manual)

- [ ] **Private vulnerability reporting** enabled — see [GITHUB-UI-FINAL-CHECKLIST.md](GITHUB-UI-FINAL-CHECKLIST.md)
- [ ] **Release immutability** enabled if the option exists for this repo/account

### B. Authenticated GHCR + signed / verify (PAT with `read:packages` and `write:packages`; SSO authorized if required)

**Next real task** — from repo root. Logs from this sequence are the **final yes/no gate** for moving from `v1.0.0-rc3` to final.

```bash
export GH_USER="codethor0"
# PAT needs read:packages and write:packages (SSO authorized if required)
export GH_PAT="ghp_..."
export DOCKERCOMMS_IT_GHCR_REPO="ghcr.io/codethor0/dockercomms"
export DOCKERCOMMS_IT_RECIPIENT="team-b"
export DOCKERCOMMS_IT_AUTH_TAG="v1.0.0-rc3"

./scripts/purge-ghcr-creds.sh
printf '%s' "$GH_PAT" | docker login ghcr.io -u "$GH_USER" --password-stdin
./scripts/login-and-run-integration.sh
./scripts/docker-e2e.sh integration
./scripts/docker-e2e.sh cli
./scripts/docker-e2e.sh full
```

Checklist (must match your release bar):

- [ ] `./scripts/login-and-run-integration.sh` — completes without skip; smoke test passes
- [ ] `./scripts/docker-e2e.sh integration` — passes
- [ ] `./scripts/docker-e2e.sh cli` — passes
- [ ] `./scripts/docker-e2e.sh full` — passes (if you require the all-in-one harness)
- [ ] **Signed** send and **`recv --verify=true`** exercised by those runs (or explicitly documented if split across scripts)
- [ ] **Verification failure** with **exit code `2`** (bundle present, crypto/policy check fails), e.g. after **signed** send; not to be confused with **exit `5`** when no bundle exists

More context and variants: [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md)

## Closure and environment record (read before GA)

This section records automation and audit findings. It does **not** replace Section A or B when those are required for your GA bar.

| Item | Status |
|------|--------|
| Section A (GitHub UI) | **Manual** only. Automation does not confirm private vulnerability reporting or release immutability. |
| Section B (GHCR + signed flows) | **Not executed** when no `GH_PAT` / `~/.dockercomms_gh_pat` is available. **GA (`v1.0.0`) stays NO-GO** until Section B succeeds with operator credentials **or** maintainers publish a **written, intentional GA scope reduction** (for example, staying on rc). |
| Local unsigned registry E2E | **Demonstrated:** `send --sign=false`, `recv --verify=false`, payload `cmp`; `recv --verify=true` on an unsigned message ends **exit 5** (bundle not found), no payload file. |
| Signed + default `recv --verify` | **Not closed here without keyless/OIDC + registry** (see README: cosign v3, keyless expected). Static `cosign sign --key` bundles are **not** interchangeable with the default Sigstore PKI / Rekor expectations inside `recv` without additional product design. |
| Dependabot | **Batch triage (2026-04-21):** PRs **#6, #9, #10** **squash-merged** after fresh CI/CodeQL. **#2–#5** (conflicting `go.sum`) addressed by a **maintainer commit** on `main` (`go get` / `go mod tidy` for rekor, sigstore stack, grpc, go-tuf; **sigstore resolves to v1.10.5** with **timestamp-authority v2.0.6**). **#7** (`actions/setup-go` v6) and **#8** (`github/codeql-action` v4) **remain open**: pushing workflow edits from automation was **blocked** by GitHub (`workflow` OAuth scope / ruleset). Merge **#7** and **#8** from the GitHub UI or a token with **workflow** scope. Re-check **Dependabot alerts** after these land. |
| Actions hygiene | Node 20 deprecation notices for some actions are **forward-looking**; plan dependency upgrades; not an automatic GA veto by themselves. |

**Declared levels:** **rc / public-share** readiness can be green when CI/CodeQL and local gates match `main`. **Final GA** requires Section B (or an explicit, documented waiver of that scope).

## Go for final tag (e.g. v1.0.0)

**GO** when all are true:

- Section A checked (or explicitly waived with reason)
- Section B matches your bar (including signed/verify and exit `2` if required)
- `main` CI and CodeQL green on the release commit
- No open stop-ship issues you care about

**NO-GO** if:

- GHCR integration or Docker CLI E2E fails with valid package-scoped auth, until root cause is fixed or scope is reduced (e.g. stay on RC).

## After GO

- Tag `v1.0.0` (or next final), publish GitHub Release, note in release body: local E2E + gates + distro matrix + live GHCR verification when applicable.
