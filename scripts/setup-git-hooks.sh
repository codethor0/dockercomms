#!/usr/bin/env bash
# Point this repository at tracked hooks under .githooks/ (strips Cursor co-author trailers).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
chmod +x .githooks/prepare-commit-msg
git config core.hooksPath .githooks
echo "core.hooksPath=$(git config core.hooksPath)"
