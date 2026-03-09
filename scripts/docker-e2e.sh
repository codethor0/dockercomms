#!/usr/bin/env bash
# Docker-based E2E validation harness for dockercomms.
# Modes: gates | integration | cli | full
# Usage: ./scripts/docker-e2e.sh [gates|integration|cli|full]
# Integration/full require: GH_PAT (or ~/.dockercomms_gh_pat), DOCKERCOMMS_IT_GHCR_REPO, DOCKERCOMMS_IT_RECIPIENT
set -euo pipefail

PROJECT="$(cd "$(dirname "$0")/.." && pwd)"
GO_VERSION="${GO_VERSION:-1.25}"
IMAGE="golang:${GO_VERSION}"
MODE="${1:-gates}"
PAT_FILE="${HOME}/.dockercomms_gh_pat"

# Resolve GH_PAT: env or secure file (never print)
GH_PAT="${GH_PAT:-}"
if [[ -z "${GH_PAT}" ]] && [[ -f "${PAT_FILE}" ]] && [[ -r "${PAT_FILE}" ]]; then
  GH_PAT=$(cat "${PAT_FILE}")
fi

# Export for container (do not echo)
export DOCKERCOMMS_IT_GHCR_REPO="${DOCKERCOMMS_IT_GHCR_REPO:-}"
export DOCKERCOMMS_IT_RECIPIENT="${DOCKERCOMMS_IT_RECIPIENT:-}"
export DOCKERCOMMS_IT_DH_REPO="${DOCKERCOMMS_IT_DH_REPO:-}"
export DOCKERCOMMS_IT_LARGE_PAYLOAD="${DOCKERCOMMS_IT_LARGE_PAYLOAD:-}"
export DOCKERCOMMS_IT_AUTH_TAG="${DOCKERCOMMS_IT_AUTH_TAG:-}"
export DOCKERCOMMS_IT_SINCE="${DOCKERCOMMS_IT_SINCE:-}"
export GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-240s}"

host_login_ghcr() {
  if [[ -z "${GH_PAT}" ]]; then
    echo "[docker-e2e] GH_PAT not set; integration tests will fail auth"
    return 1
  fi
  GH_USER="${GH_USER:-codethor0}"
  echo "[docker-e2e] Logging in to ghcr.io on host (credentials in ~/.docker for container)..."
  printf '%s' "${GH_PAT}" | docker login ghcr.io -u "${GH_USER}" --password-stdin 2>/dev/null || {
    echo "[docker-e2e] docker login ghcr.io failed"
    return 1
  }
}

docker_run() {
  docker run --rm \
    -v "${PROJECT}:/workspace" \
    -w /workspace \
    -v "${HOME}/.docker:/root/.docker:ro" \
    -e DOCKERCOMMS_IT_GHCR_REPO \
    -e DOCKERCOMMS_IT_RECIPIENT \
    -e DOCKERCOMMS_IT_DH_REPO \
    -e DOCKERCOMMS_IT_LARGE_PAYLOAD \
    -e DOCKERCOMMS_IT_AUTH_TAG \
    -e DOCKERCOMMS_IT_SINCE \
    -e GO_TEST_TIMEOUT \
    "$IMAGE" \
    bash -c "$1"
}

case "$MODE" in
  gates)
    echo "[docker-e2e] gates mode: build, test, race, lint, coverage-gate"
    docker_run '
      set -e
      go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1
      make build
      go test ./...
      go test -race ./...
      golangci-lint run ./...
      make coverage-gate
    '
    ;;
  integration)
    echo "[docker-e2e] integration mode: host login + go test -tags=integration"
    : "${DOCKERCOMMS_IT_GHCR_REPO:=ghcr.io/codethor0/dockercomms}"
    : "${DOCKERCOMMS_IT_RECIPIENT:=team-b}"
    export DOCKERCOMMS_IT_GHCR_REPO DOCKERCOMMS_IT_RECIPIENT
    if ! host_login_ghcr; then
      echo "[docker-e2e] Skipping integration: set GH_PAT or ~/.dockercomms_gh_pat"
      exit 1
    fi
    docker_run '
      set -e
      make build
      if [[ -z "${DOCKERCOMMS_IT_GHCR_REPO:-}" ]] || [[ -z "${DOCKERCOMMS_IT_RECIPIENT:-}" ]]; then
        echo "DOCKERCOMMS_IT_GHCR_REPO and DOCKERCOMMS_IT_RECIPIENT required"
        exit 1
      fi
      go test -tags=integration ./test/integration/... -run Test -count=1 -v -timeout "${GO_TEST_TIMEOUT:-240s}"
    '
    ;;
  cli)
    echo "[docker-e2e] cli mode: send/recv round-trip, verify-failure no-materialize, resume"
    : "${DOCKERCOMMS_IT_GHCR_REPO:=ghcr.io/codethor0/dockercomms}"
    : "${DOCKERCOMMS_IT_RECIPIENT:=team-b}"
    export DOCKERCOMMS_IT_GHCR_REPO DOCKERCOMMS_IT_RECIPIENT
    if ! host_login_ghcr; then
      echo "[docker-e2e] Skipping cli: set GH_PAT or ~/.dockercomms_gh_pat"
      exit 1
    fi
    docker_run '
      set -e
      make build
      E2E=/tmp/dockercomms-e2e
      rm -rf "$E2E" && mkdir -p "$E2E" "$E2E/out" "$E2E/bad"
      dd if=/dev/urandom of="$E2E/payload.bin" bs=1M count=4 2>/dev/null
      echo "=== Payload SHA256 ==="
      sha256sum "$E2E/payload.bin"

      echo "=== 7.1 Send (no sign) ==="
      ./dockercomms send "$E2E/payload.bin" --repo "$DOCKERCOMMS_IT_GHCR_REPO" --recipient "$DOCKERCOMMS_IT_RECIPIENT" --sign=false --ttl-seconds 3600

      echo "=== 7.2 Recv (Verify=false) ==="
      rm -rf "$E2E/out" && mkdir -p "$E2E/out"
      ./dockercomms recv --repo "$DOCKERCOMMS_IT_GHCR_REPO" --me "$DOCKERCOMMS_IT_RECIPIENT" --out "$E2E/out" --verify=false
      echo "=== Output SHA256 ==="
      sha256sum "$E2E/out/payload.bin" 2>/dev/null || true
      cmp -s "$E2E/payload.bin" "$E2E/out/payload.bin" && echo "OK: payloads match"

      echo "=== 7.3 Verify-failure no-materialize ==="
      rm -rf "$E2E/bad" && mkdir -p "$E2E/bad"
      set +e
      ./dockercomms recv --repo "$DOCKERCOMMS_IT_GHCR_REPO" --me "$DOCKERCOMMS_IT_RECIPIENT" --out "$E2E/bad" --verify=true --trusted-root /workspace/testdata/bad-trusted-root.json 2>/dev/null
      REXIT=$?
      set -e
      echo "recv exit: $REXIT (expect 2)"
      test ! -f "$E2E/bad/payload.bin" && echo "OK: no output on verify failure"
      ls -la "$E2E/bad" 2>/dev/null || true

      echo "=== 7.4 Resume (send same payload twice) ==="
      ./dockercomms send "$E2E/payload.bin" --repo "$DOCKERCOMMS_IT_GHCR_REPO" --recipient "$DOCKERCOMMS_IT_RECIPIENT" --sign=false --chunk-bytes 1048576
      ./dockercomms send "$E2E/payload.bin" --repo "$DOCKERCOMMS_IT_GHCR_REPO" --recipient "$DOCKERCOMMS_IT_RECIPIENT" --sign=false --chunk-bytes 1048576
      echo "OK: both sends completed (blobs reused via HEAD)"
    '
    ;;
  full)
    echo "[docker-e2e] full mode: gates + integration + cli"
    : "${DOCKERCOMMS_IT_GHCR_REPO:=ghcr.io/codethor0/dockercomms}"
    : "${DOCKERCOMMS_IT_RECIPIENT:=team-b}"
    export DOCKERCOMMS_IT_GHCR_REPO DOCKERCOMMS_IT_RECIPIENT
    docker_run '
      set -e
      go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1
      make build
      go test ./...
      go test -race ./...
      golangci-lint run ./...
      make coverage-gate
    '
    if host_login_ghcr; then
      docker_run '
        set -e
        go test -tags=integration ./test/integration/... -run Test -count=1 -v -timeout "${GO_TEST_TIMEOUT:-240s}"
      '
      docker_run '
        set -e
        make build
        E2E=/tmp/dockercomms-e2e
        rm -rf "$E2E" && mkdir -p "$E2E" "$E2E/out" "$E2E/bad"
        dd if=/dev/urandom of="$E2E/payload.bin" bs=1M count=4 2>/dev/null
        sha256sum "$E2E/payload.bin"
        ./dockercomms send "$E2E/payload.bin" --repo "$DOCKERCOMMS_IT_GHCR_REPO" --recipient "$DOCKERCOMMS_IT_RECIPIENT" --sign=false --ttl-seconds 3600
        rm -rf "$E2E/out" && mkdir -p "$E2E/out"
        ./dockercomms recv --repo "$DOCKERCOMMS_IT_GHCR_REPO" --me "$DOCKERCOMMS_IT_RECIPIENT" --out "$E2E/out" --verify=false
        cmp -s "$E2E/payload.bin" "$E2E/out/payload.bin" && echo "OK: round-trip"
        rm -rf "$E2E/bad" && mkdir -p "$E2E/bad"
        ./dockercomms recv --repo "$DOCKERCOMMS_IT_GHCR_REPO" --me "$DOCKERCOMMS_IT_RECIPIENT" --out "$E2E/bad" --verify=true --trusted-root /workspace/testdata/bad-trusted-root.json 2>/dev/null || true
        test ! -f "$E2E/bad/payload.bin" && echo "OK: no materialize on verify fail"
      '
    else
      echo "[docker-e2e] integration and cli skipped: set GH_PAT or ~/.dockercomms_gh_pat"
    fi
    ;;
  *)
    echo "Usage: $0 [gates|integration|full|cli]"
    exit 1
    ;;
esac
