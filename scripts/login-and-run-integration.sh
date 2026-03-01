#!/usr/bin/env bash
# GHCR login + integration tests. Non-interactive when GH_PAT provided.
# Usage: ./scripts/login-and-run-integration.sh [--help|--check]
# Requires: export GH_PAT="ghp_..." OR ~/.dockercomms_gh_pat (chmod 600)
set -euo pipefail
# Restrict new file perms; PAT file must already be chmod 600
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
  if ! curl -sS -I --max-time 15 https://ghcr.io/v2/ >/dev/null 2>&1; then
    echo "  FAIL: Cannot reach ghcr.io (network/proxy/DNS?)"
    err=1
  else
    echo "  OK"
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
    GH_PAT=$(cat "${PAT_FILE}")
  fi

  if [[ -z "${GH_PAT}" ]]; then
    if [[ ! -t 0 ]] || [[ ! -t 1 ]]; then
      echo "ERROR: Non-TTY and GH_PAT not set. Cannot prompt for credentials."
      echo "  Set: export GH_PAT='ghp_...'"
      echo "  Or:  echo 'ghp_...' > ~/.dockercomms_gh_pat && chmod 600 ~/.dockercomms_gh_pat"
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

  echo "[1/4] Logging out of ghcr.io (ignore errors if not logged in)..."
  docker logout ghcr.io >/dev/null 2>&1 || true

  echo "[2/4] Logging in to ghcr.io non-interactively as ${GH_USER}..."
  printf '%s' "${GH_PAT}" | docker login ghcr.io -u "${GH_USER}" --password-stdin

  echo "[3/4] Verifying Docker can hit GHCR with auth..."
  docker pull "ghcr.io/${GH_USER}/nonexistent-verify" 2>/dev/null || true
  echo "OK (auth path exercised; nonexistent image pull failure is expected)."

  echo "[4/4] Running integration script..."
  cd "${PROJECT}"
  chmod +x "${SCRIPT}"
  "${SCRIPT}"
}

main "$@"
