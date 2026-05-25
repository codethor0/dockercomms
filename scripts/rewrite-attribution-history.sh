#!/usr/bin/env bash
# Maintainer-only: rewrite main history to remove bot/agent commit attribution.
# Requires temporary allow_force_pushes on main, then restore protection after push.
# Preserves trees; rewrites author/email/message metadata only.
set -euo pipefail

if [[ "${1:-}" != "--confirm" ]]; then
  echo "Usage: $0 --confirm"
  echo "Run only from a fresh clone. This rewrites all of main and force-pushes."
  exit 1
fi

git filter-repo --force \
  --name-callback 'return b"Isaac Thor" if name in (b"dependabot[bot]",) else name' \
  --email-callback 'return b"codethor@gmail.com" if b"dependabot" in email else email' \
  --message-callback '
lines = message.split(b"\n")
out = []
for line in lines:
    if line.startswith(b"Co-authored-by: Cursor"):
        continue
    if line.startswith(b"Co-authored-by: dependabot"):
        continue
    if line.startswith(b"Made-with: Cursor"):
        continue
    if line.startswith(b"Signed-off-by: dependabot"):
        continue
    out.append(line)
body = b"\n".join(out).rstrip(b"\n")
return body + b"\n" if body else message
'

echo "Rewrite done. Verify: git log --author=dependabot; git log --format=%B | grep cursoragent"
echo "Then: git push --force origin main && git push --force origin --tags"
