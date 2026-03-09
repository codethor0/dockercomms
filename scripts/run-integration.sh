#!/usr/bin/env bash
# Run integration tests against GHCR. Requires: docker login ghcr.io first.
# Exit 3 = auth failure (see login-and-run-integration.sh).
# Usage: ./scripts/run-integration.sh [--check]
# Optional env: DOCKERCOMMS_IT_DH_REPO, DOCKERCOMMS_IT_LARGE_PAYLOAD
set -euo pipefail

PROJECT="$(cd "$(dirname "$0")/.." && pwd)"
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-120s}"

check_mode() {
  echo "[check] Project root: ${PROJECT}"
  echo "[check] Env vars:"
  : "${DOCKERCOMMS_IT_GHCR_REPO:=ghcr.io/codethor0/dockercomms}"
  : "${DOCKERCOMMS_IT_RECIPIENT:=team-b}"
  echo "  DOCKERCOMMS_IT_GHCR_REPO=${DOCKERCOMMS_IT_GHCR_REPO}"
  echo "  DOCKERCOMMS_IT_RECIPIENT=${DOCKERCOMMS_IT_RECIPIENT}"
  if [[ -n "${DOCKERCOMMS_IT_DH_REPO:-}" ]]; then
    echo "  DOCKERCOMMS_IT_DH_REPO=${DOCKERCOMMS_IT_DH_REPO}"
  fi
  if [[ -n "${DOCKERCOMMS_IT_LARGE_PAYLOAD:-}" ]]; then
    echo "  DOCKERCOMMS_IT_LARGE_PAYLOAD=${DOCKERCOMMS_IT_LARGE_PAYLOAD}"
  fi
  echo "[check] Docker daemon..."
  if ! docker info >/dev/null 2>&1; then
    echo "  FAIL: Start Docker Desktop"
    exit 1
  fi
  echo "  OK"
  echo "[check] GHCR connectivity..."
  code=$(curl -sS -o /dev/null -w "%{http_code}" --max-time 10 https://ghcr.io/v2/ 2>/dev/null) || true
  if [[ "$code" == "401" ]] || [[ "$code" == "405" ]]; then
    echo "  OK (got $code)"
  else
    echo "  FAIL: Cannot reach ghcr.io (got ${code:-timeout})"
    exit 1
  fi
  echo "[check] Auth (best-effort)..."
  if docker manifest inspect "${DOCKERCOMMS_IT_GHCR_REPO}:latest" >/dev/null 2>&1; then
    echo "  OK (auth verified)"
  else
    echo "  auth not verified (expected if not logged in)"
  fi
  echo "[check] Go test path..."
  if [[ -d "${PROJECT}/test/integration" ]]; then
    echo "  OK"
  else
    echo "  FAIL: test/integration not found"
    exit 1
  fi
  echo "All checks passed."
}

main() {
  if [[ "${1:-}" == "--check" ]]; then
    check_mode
    exit 0
  fi

  echo "== System =="
  uname -a
  echo
  echo "== Go =="
  go version
  echo
  echo "== Docker =="
  docker version --format '{{.Client.Version}}' 2>/dev/null || true
  echo
  echo "== Docker daemon check =="
  if ! docker info >/dev/null 2>&1; then
    echo "Docker daemon: NOT reachable"
    echo "Start Docker Desktop and re-run."
    exit 1
  fi
  echo "Docker daemon: OK"
  echo
  echo "== Env vars =="
  export DOCKERCOMMS_IT_GHCR_REPO="${DOCKERCOMMS_IT_GHCR_REPO:-ghcr.io/codethor0/dockercomms}"
  export DOCKERCOMMS_IT_RECIPIENT="${DOCKERCOMMS_IT_RECIPIENT:-team-b}"
  echo "DOCKERCOMMS_IT_GHCR_REPO=$DOCKERCOMMS_IT_GHCR_REPO"
  echo "DOCKERCOMMS_IT_RECIPIENT=$DOCKERCOMMS_IT_RECIPIENT"
  if [[ -n "${DOCKERCOMMS_IT_DH_REPO:-}" ]]; then
    export DOCKERCOMMS_IT_DH_REPO
    echo "DOCKERCOMMS_IT_DH_REPO=$DOCKERCOMMS_IT_DH_REPO"
  fi
  if [[ -n "${DOCKERCOMMS_IT_LARGE_PAYLOAD:-}" ]]; then
    export DOCKERCOMMS_IT_LARGE_PAYLOAD
    echo "DOCKERCOMMS_IT_LARGE_PAYLOAD=$DOCKERCOMMS_IT_LARGE_PAYLOAD"
  fi
  echo
  echo "== Run integration tests =="
  cd "${PROJECT}"
  set +e
  go test -tags=integration ./test/integration/... -run Test -count=1 -v -timeout "${GO_TEST_TIMEOUT}" 2>&1 | tee /tmp/dockercomms_it.out
  r=${PIPESTATUS[0]}
  set -e
  if [[ $r -ne 0 ]] && grep -q "context deadline exceeded" /tmp/dockercomms_it.out 2>/dev/null; then
    echo ""
    echo "This is almost certainly missing/invalid GHCR auth; run ./scripts/login-and-run-integration.sh"
  fi
  exit $r
}

main "$@"
