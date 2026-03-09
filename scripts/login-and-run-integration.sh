#!/usr/bin/env bash
# GHCR login + integration tests. Non-interactive when GH_PAT provided.
# Exit 3 = auth failure (login or manifest inspect).
# Usage: ./scripts/login-and-run-integration.sh [--help|--check]
# Requires: export GH_PAT="ghp_..." OR ~/.dockercomms_gh_pat (chmod 600)
set -euo pipefail
umask 077

PROJECT="$(cd "$(dirname "$0")/.." && pwd)"
SCRIPT="${PROJECT}/scripts/run-integration.sh"
PAT_FILE="${HOME}/.dockercomms_gh_pat"

usage() {
  cat <<'EOF'
Usage: ./scripts/login-and-run-integration.sh [--help|--check]

  --help   Print this message
  --check  Validate prerequisites and env (no login, no PAT required)

Non-interactive login requires one of:
  - export GH_PAT="ghp_..."
  - printf '%s' 'ghp_...' > ~/.dockercomms_gh_pat && chmod 600 ~/.dockercomms_gh_pat

Required env (or defaults used):
  DOCKERCOMMS_IT_GHCR_REPO (default: ghcr.io/codethor0/dockercomms)
  DOCKERCOMMS_IT_RECIPIENT (default: team-b)

Optional: GH_USER (default: codethor0)
  DOCKERCOMMS_IT_AUTH_TAG (default: latest; use existing tag if :latest missing)
EOF
}

preflight() {
  local err=0
  echo "[preflight] Docker daemon..."
  if ! docker info >/dev/null 2>&1; then
    echo "  FAIL: Docker daemon not reachable. Start Docker Desktop."
    err=1
  else
    echo "  OK"
  fi

  echo "[preflight] GHCR connectivity..."
  code=$(curl -sS -o /dev/null -w "%{http_code}" --max-time 10 https://ghcr.io/v2/ 2>/dev/null) || true
  if [[ "$code" == "401" ]] || [[ "$code" == "405" ]]; then
    echo "  OK (got $code)"
  else
    echo "  FAIL: Cannot reach ghcr.io (got ${code:-timeout})"
    err=1
  fi

  echo "[preflight] Env vars..."
  : "${DOCKERCOMMS_IT_GHCR_REPO:=ghcr.io/codethor0/dockercomms}"
  : "${DOCKERCOMMS_IT_RECIPIENT:=team-b}"
  export DOCKERCOMMS_IT_GHCR_REPO DOCKERCOMMS_IT_RECIPIENT
  echo "  DOCKERCOMMS_IT_GHCR_REPO=$DOCKERCOMMS_IT_GHCR_REPO"
  echo "  DOCKERCOMMS_IT_RECIPIENT=$DOCKERCOMMS_IT_RECIPIENT"
  echo "  OK"

  if [[ $err -ne 0 ]]; then
    echo "Preflight failed. Fix above and re-run."
    exit 1
  fi
}

auth_proof() {
  local auth_err tag
  : "${DOCKERCOMMS_IT_AUTH_TAG:=latest}"
  tag="${DOCKERCOMMS_IT_AUTH_TAG}"
  auth_err=$(DOCKER_CLIENT_TIMEOUT=20 DOCKER_HTTP_TIMEOUT=20 docker manifest inspect "${DOCKERCOMMS_IT_GHCR_REPO}:${tag}" 2>&1) || true
  if echo "${auth_err}" | grep -qE "manifest unknown|not found|no such manifest"; then
    echo "Login succeeded but ${DOCKERCOMMS_IT_GHCR_REPO}:${tag} not found."
    echo "  Set DOCKERCOMMS_IT_AUTH_TAG to an existing tag or proceed (tests will validate auth via registry operations)."
    return 0
  fi
  if echo "${auth_err}" | grep -qE "unauthorized|denied|authentication required|insufficient_scope|insufficient scope|invalid token|invalid username/password|access to the resource is denied"; then
    echo "Auth to GHCR failed (docker manifest inspect). Exiting 3."
    echo "  PAT must have read:packages + write:packages."
    echo "  Re-run ./scripts/purge-ghcr-creds.sh if old creds interfere."
    exit 3
  fi
}

main() {
  case "${1:-}" in
    --help|-h) usage; exit 0 ;;
    --check)
      preflight
      echo "[check] run-integration.sh..."
      "${SCRIPT}" --check
      echo "All checks passed."
      exit 0
      ;;
  esac

  GH_USER="${GH_USER:-codethor0}"
  GH_PAT="${GH_PAT:-}"
  if [[ -z "${GH_PAT}" ]] && [[ -f "${PAT_FILE}" ]]; then
    if [[ -r "${PAT_FILE}" ]]; then
      GH_PAT=$(cat "${PAT_FILE}")
    fi
  fi

  if [[ -z "${GH_PAT}" ]]; then
    if [[ ! -t 0 ]] || [[ ! -t 1 ]]; then
      echo "ERROR: Non-TTY and GH_PAT not set. Cannot prompt for credentials."
      echo "  Set: export GH_PAT='ghp_...'"
      echo "  Or:  printf '%s' 'ghp_...' > ~/.dockercomms_gh_pat && chmod 600 ~/.dockercomms_gh_pat"
      echo "  Never paste PAT into issues or logs."
      exit 1
    fi
    echo "ERROR: GH_PAT is not set."
    echo "  Option 1: export GH_PAT='YOUR_GITHUB_PAT'"
    echo "  Option 2: printf '%s' 'ghp_...' > ~/.dockercomms_gh_pat && chmod 600 ~/.dockercomms_gh_pat"
    echo "  Never paste PAT into issues or logs."
    exit 1
  fi

  preflight

  echo "[1/5] Logging out of ghcr.io (ignore errors if not logged in)..."
  docker logout ghcr.io >/dev/null 2>&1 || true

  echo "[2/5] Logging in to ghcr.io non-interactively as ${GH_USER}..."
  if ! printf '%s' "${GH_PAT}" | DOCKER_CLIENT_TIMEOUT=20 DOCKER_HTTP_TIMEOUT=20 docker login ghcr.io -u "${GH_USER}" --password-stdin; then
    echo "Auth to GHCR failed (docker login). Exiting 3."
    echo "  PAT must have read:packages + write:packages."
    echo "  Re-run ./scripts/purge-ghcr-creds.sh if old creds interfere."
    exit 3
  fi

  echo "[3/5] Auth proof (must require auth, return quickly)..."
  auth_proof

  echo "[4/5] Running integration script..."
  cd "${PROJECT}"
  chmod +x "${SCRIPT}"
  "${SCRIPT}" || {
    e=$?
    if [[ $e -ne 0 ]]; then
      echo ""
      echo "Integration test failed. If you saw 'context deadline exceeded',"
      echo "this is almost certainly missing/invalid GHCR auth; re-run this script"
      echo "after verifying PAT and purge if needed."
    fi
    exit $e
  }

  echo "[5/5] Paste-back template (no secrets):"
  echo "  Repo: $DOCKERCOMMS_IT_GHCR_REPO"
  echo "  Recipient: $DOCKERCOMMS_IT_RECIPIENT"
  echo "  Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "  Cosign: $(cosign version 2>/dev/null || echo 'NOT INSTALLED')"
}

main "$@"
