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
