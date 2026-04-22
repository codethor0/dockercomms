# Release Runbook: v1.0.0-rc3 → Final

Use this when you have `GH_PAT` with `read:packages` and `write:packages` and want to complete the live GHCR verification before releasing.

## Current status (until live run passes)

> DockerComms is green on all current non-credentialed gates, but live GHCR end-to-end verification is still pending valid package-scoped auth.

## Prerequisites

- Docker daemon running
- GitHub PAT with `read:packages` and `write:packages` (SSO authorized if org repo)
- GHCR repo you control (e.g. `ghcr.io/codethor0/dockercomms`)

## Steps

### 1. Export credentials and purge prior creds

```bash
export GH_USER="codethor0"
export GH_PAT="ghp_...your_pat_with_read:packages_write:packages..."
export DOCKERCOMMS_IT_GHCR_REPO="ghcr.io/codethor0/dockercomms"
export DOCKERCOMMS_IT_RECIPIENT="team-b"
export DOCKERCOMMS_IT_AUTH_TAG="v1.0.0-rc3"

cd /path/to/dockercomms
./scripts/purge-ghcr-creds.sh
printf '%s' "$GH_PAT" | docker login ghcr.io -u "$GH_USER" --password-stdin
```

### 2. Run host integration

```bash
./scripts/login-and-run-integration.sh 2>&1 | tee /tmp/dockercomms_host_integration.log
```

### 3. Run Dockerized integration and CLI E2E

```bash
./scripts/docker-e2e.sh integration 2>&1 | tee /tmp/dockercomms_docker_integration.log
./scripts/docker-e2e.sh cli 2>&1 | tee /tmp/dockercomms_docker_cli.log
./scripts/docker-e2e.sh full 2>&1 | tee /tmp/dockercomms_docker_full.log
```

## Success criteria (all-clear)

- Host integration runs and does not skip
- Docker integration passes
- Docker CLI E2E passes
- send/recv round-trip succeeds
- Negative verify case proves **no materialization**
- resume/HEAD behavior works
- Host gates still pass afterward
- Docker gates still pass afterward
- Repo stays clean except for untracked local evidence files

## After a successful run

If all four commands pass, flip from rc to final:

1. Tag `v1.0.0` from the same commit as `v1.0.0-rc3` (or latest main)
2. Create GitHub Release with body from `RELEASE.md`
3. Update release notes to include: "Fully verified end-to-end against live GHCR"

Paste the output for audit; a straight yes/no on flipping to final can be given from the logs.

## Master prompt report hygiene (v8/v9-friendly)

**GA line (§4):** Use **YELLOW** when the GA/attestation track is **incomplete or blocked** (no token, no UI proof, not run) but **nothing in §A/§B was disproven**. Reserve **RED** for that track when your rubric says “not satisfied” *or* when evidence shows a **hard failure** (broken checks, policy violation, falsified claim), not merely “we could not prove it this session.”

**MODE=ALL vs E2E:** If any required E2E sub-step is not executed in that session, label the run **MODE=ALL (partial E2E)** and list each skipped sub-step under **NOT RUN** (or **BLOCKED**) with one reason each. That keeps “MODE=ALL” from reading as “full matrix complete” when only gates + host build ran.

**Shell:** One-liner evidence scripts that rely on `set +e` before a command that may return non-zero should be run under **bash**; **zsh** treats `errexit` differently, so `E1=$?` after a failing command may never run unless `setopt` is adjusted.

**Self-grade (v9 §12):** Lines 3–5 of the six-line block are the **v10 preview**; for strict v9, each of those three lines is **exactly six words** (count at write time; if you cannot meet that, label the block **v9-style six lines** in the report header so audits know).

**Self-grade line 1 (v10 §13):** After `Evidence High|Med|Low:`, the clause is **8 words max** (count after the colon).

**Pristine E2E bar (v10):** For a stop-ship “clean” pass, use an **empty** `git status` before E2E, or **commit/push** doc changes and re-anchor gates on clean `main`—or obtain a **one-line** same-message operator approval for **E2E on a doc-only dirty tree**; otherwise list **B-PROC-1** / procedure risk honestly.

**E2E matrix triage column:** Use **—** or **PRODUCT clean** when the step only succeeded. **B-PROC-1** = operator/procedure (e.g. dirty tree, shell discipline before `bash` fix). **NOT RUN (auth)** = missing `GH_PAT` / credentialed work—not the same as **B-PROC-1**.

**License (CLOSURE):** `gh api repos/.../license` may show `license.key: other` while the root **LICENSE** file is Apache-2.0; one line that **API license metadata ≠ LICENSE file text** prevents audit contradiction.
