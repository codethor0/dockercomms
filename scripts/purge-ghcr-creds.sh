#!/usr/bin/env bash
# Purge bad cached GHCR creds from Docker config. Run if login shows "denied".
# Usage: ./scripts/purge-ghcr-creds.sh
set -euo pipefail

DOCKERCFG="${HOME}/.docker/config.json"

echo "[1/3] docker logout ghcr.io"
docker logout ghcr.io >/dev/null 2>&1 || true

echo "[2/3] Removing ghcr.io entry from Docker config (if present)"
if [[ -f "${DOCKERCFG}" ]]; then
  cp "${DOCKERCFG}" "${DOCKERCFG}.bak.$(date +%s)"
  python3 - <<'PY'
import json, os
p = os.path.expanduser("~/.docker/config.json")
with open(p) as f:
    d = json.load(f)
d.get("auths", {}).pop("ghcr.io", None)
with open(p, "w") as f:
    json.dump(d, f, indent=2)
print("Updated:", p)
PY
else
  echo "  Config not found; nothing to purge."
fi

echo "[3/3] Done."
echo ""
echo "Next: provide PAT and run integration:"
echo "  export GH_PAT='ghp_...'"
echo "  ./scripts/login-and-run-integration.sh"
echo ""
echo "If credential helper (credsStore) still returns old creds, manually:"
echo "  - Docker Desktop: docker-credential-desktop erase ghcr.io"
echo "  - macOS Keychain: delete 'ghcr.io' entry in login keychain"
echo "  - Then: ./scripts/login-and-run-integration.sh"
