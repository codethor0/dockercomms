# PR: scripts/ghcr-integration-runner → main

Use this when opening the PR at:
https://github.com/codethor0/dockercomms/compare/main...scripts/ghcr-integration-runner

---

## PR title

```
scripts: add GHCR login+integration runner with safe PAT handling
```

## PR description

This PR adds a surgical, operator-safe path to run DockerComms integration tests against GHCR without touching Go code or CI gates. It introduces `scripts/login-and-run-integration.sh` and `scripts/run-integration.sh` with preflight checks, `--check`/`--help`, non-TTY handling, strict permissions (umask 077), and PAT sourcing via `GH_PAT` or `~/.dockercomms_gh_pat` (never logged). It also updates `scripts/purge-ghcr-creds.sh` for credential recovery and documents the scripted workflow in `docs/repro.md`. All existing gates remain green (unit tests, race, lint, coverage-gate).

**How to verify (copy/paste):**

```bash
make test && make coverage-gate
./scripts/run-integration.sh --check
./scripts/login-and-run-integration.sh --check
```

**Reviewer checklist:**

* [ ] No Go code changes; only `scripts/` and `docs/repro.md` changed
* [ ] PAT handling is safe (never echoed, file outside repo, umask 077)
* [ ] `./scripts/login-and-run-integration.sh --check` passes
* [ ] `./scripts/run-integration.sh --check` passes
* [ ] `docs/repro.md` matches the scripted workflow and behavior
* [ ] `scripts/purge-ghcr-creds.sh` provides correct manual recovery steps and does not delete unrelated Docker creds
* [ ] Scripts are deterministic and do not mutate repo state unexpectedly
