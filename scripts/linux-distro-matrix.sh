#!/usr/bin/env bash
# Local-only Linux userland matrix: local registry on Docker network, no GHCR.
# Usage: from repo root, ./scripts/linux-distro-matrix.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

NET="dockercomms-e2e-net"
REG_NAME="dockercomms-registry"
PLATFORM="${DOCKERCOMMS_DISTRO_PLATFORM:-linux/arm64}"
case "$PLATFORM" in
linux/arm64)
  BIN="$ROOT/dist/dockercomms-linux-arm64"
  GOARCH_HINT=arm64
  ;;
linux/amd64)
  BIN="$ROOT/dist/dockercomms-linux-amd64"
  GOARCH_HINT=amd64
  ;;
*)
  echo "unsupported DOCKERCOMMS_DISTRO_PLATFORM=$PLATFORM (use linux/arm64 or linux/amd64)" >&2
  exit 1
  ;;
esac
if [[ ! -f "$BIN" ]]; then
  echo "missing $BIN — run: CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH_HINT} go build -o dist/dockercomms-linux-${GOARCH_HINT} ./cmd/dockercomms" >&2
  exit 1
fi
chmod +x "$BIN" || true

docker rm -f "$REG_NAME" 2>/dev/null || true
docker network rm "$NET" 2>/dev/null || true
docker network create "$NET"
docker run -d --name "$REG_NAME" --network "$NET" --platform "$PLATFORM" registry:2
sleep 2

# Inner runner: args = distro_slug install_cmd (single line)
run_inner() {
  local slug="$1"
  local install_cmd="$2"
  local repo="${REG_NAME}:5000/e2e-${slug}"

  docker run --rm \
    --platform "$PLATFORM" \
    --network "$NET" \
    -v "${BIN}:/usr/local/bin/dockercomms:ro" \
    "${IMAGE}" \
    sh -ec "
set -e
export PATH=\"/usr/local/bin:\$PATH\"

echo \"[${slug}] install...\"
${install_cmd}

echo \"[${slug}] registry ping...\"
curl -fsS \"http://${REG_NAME}:5000/v2/\" >/dev/null

DC=dockercomms
REPO='${repo}'
BASE=/tmp/dc-e2e-${slug}
rm -rf \"\$BASE\"
mkdir -p \"\$BASE/in\" \"\$BASE/out\" \"\$BASE/out-wrong\" \"\$BASE/logs\"

\$DC --help >/dev/null
\$DC send --help >/dev/null
\$DC recv --help >/dev/null
\$DC verify --help >/dev/null

echo hello > \"\$BASE/in/small.txt\"
dd if=/dev/urandom of=\"\$BASE/in/bin8k.bin\" bs=1024 count=8 2>/dev/null
: > \"\$BASE/in/empty.dat\"
dd if=/dev/urandom of=\"\$BASE/in/medium.bin\" bs=1048576 count=32 2>/dev/null

check_eq() {
  inf=\"\$1\"; outf=\"\$2\"
  hi=\$(sha256sum \"\$inf\" | awk '{print \$1}')
  ho=\$(sha256sum \"\$outf\" | awk '{print \$1}')
  [ \"\$hi\" = \"\$ho\" ] || { echo \"hash mismatch \$inf vs \$outf\"; exit 1; }
}

for f in small.txt bin8k.bin empty.dat medium.bin; do
  rm -f \"\$BASE/out/\$f\" 2>/dev/null || true
  \$DC send --repo \"\$REPO\" --recipient team-a --sign=false \"\$BASE/in/\$f\"
  \$DC recv --repo \"\$REPO\" --me team-a --out \"\$BASE/out\" --verify=false --write-receipt=false --max 50
  check_eq \"\$BASE/in/\$f\" \"\$BASE/out/\$f\"
done

rm -rf \"\$BASE/out-wrong\"; mkdir -p \"\$BASE/out-wrong\"
\$DC send --repo \"\$REPO\" --recipient team-a --sign=false \"\$BASE/in/small.txt\"
out=\"\$(\$DC recv --repo \"\$REPO\" --me team-b --out \"\$BASE/out-wrong\" --verify=false --write-receipt=false --max 10 2>&1)\"
echo \"\$out\" | grep -q 'received 0 message' || { echo \"wrong-recipient: expected 0 messages: \$out\"; exit 1; }
[ \"\$(ls -A \"\$BASE/out-wrong\" 2>/dev/null | wc -l | tr -d ' ')\" = \"0\" ] || { echo \"wrong-recipient: out dir not empty\"; ls -la \"\$BASE/out-wrong\"; exit 1; }

echo path > /tmp/outside-dc.txt
rm -f \"\$BASE/out/outside-dc.txt\" 2>/dev/null || true
\$DC send --repo \"\$REPO\" --recipient team-a --sign=false /tmp/outside-dc.txt
\$DC recv --repo \"\$REPO\" --me team-a --out \"\$BASE/out\" --verify=false --write-receipt=false --max 50
check_eq /tmp/outside-dc.txt \"\$BASE/out/outside-dc.txt\"
for n in \$(ls -A \"\$BASE/out\"); do
  case \"\$n\" in *[/\\\\]*) echo \"bad out name: \$n\"; exit 1 ;; esac
  [ \"\$n\" != '.' ] && [ \"\$n\" != '..' ] || exit 1
done

printf 'z' > \"\$BASE/in/foo..bar.txt\"
\$DC send --repo \"\$REPO\" --recipient team-a --sign=false \"\$BASE/in/foo..bar.txt\"
\$DC recv --repo \"\$REPO\" --me team-a --out \"\$BASE/out\" --verify=false --write-receipt=false --max 50
check_eq \"\$BASE/in/foo..bar.txt\" \"\$BASE/out/foo..bar.txt\"

REPOR=\"\${REPO}-repeat\"
rm -rf \"\$BASE/out-r\" && mkdir -p \"\$BASE/out-r\"
\$DC send --repo \"\$REPOR\" --recipient team-a --sign=false \"\$BASE/in/small.txt\"
\$DC send --repo \"\$REPOR\" --recipient team-a --sign=false \"\$BASE/in/small.txt\"
out2=\"\$(\$DC recv --repo \"\$REPOR\" --me team-a --out \"\$BASE/out-r\" --verify=false --write-receipt=false --max 10 2>&1)\"
echo \"\$out2\" | grep -q 'received 2 message' || { echo \"repeat send: want 2 msgs: \$out2\"; exit 1; }
check_eq \"\$BASE/in/small.txt\" \"\$BASE/out-r/small.txt\"

echo \"[${slug}] OK\"
"
}

declare -a FAILS=()

try_distro() {
  local name="$1"
  local image="$2"
  local install="$3"
  IMAGE="$image"
  echo "========== $name ($image) =========="
  if ! run_inner "$(echo "$name" | tr '[:upper:]' '[:lower:]' | tr '/' '-')" "$install"; then
    FAILS+=("$name")
    echo "FAIL: $name"
  else
    echo "PASS: $name"
  fi
}

try_distro "ubuntu-24.04" "ubuntu:24.04" \
  "apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq curl ca-certificates coreutils >/dev/null"

try_distro "debian-13" "debian:13" \
  "apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq curl ca-certificates coreutils >/dev/null"

try_distro "alpine-3.23" "alpine:3.23" \
  "apk add --no-cache -q curl ca-certificates coreutils"

try_distro "fedora-42" "fedora:42" \
  "(command -v microdnf >/dev/null && microdnf install -y curl coreutils ca-certificates) || dnf install -y -q curl coreutils ca-certificates"

try_distro "rockylinux-9" "rockylinux:9" \
  "true"

try_distro "opensuse-leap-16" "opensuse/leap:16.0" \
  "zypper --non-interactive install -y curl coreutils ca-certificates >/dev/null"

echo "========== SUMMARY =========="
if [ "${#FAILS[@]}" -eq 0 ]; then
  echo "All distros PASS"
  exit 0
fi
echo "FAILED: ${FAILS[*]}"
exit 1
