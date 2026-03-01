# Release Checklist

Stop-ship gates. All must pass before release.

## Spec Compliance Checklist (SPEC.md)

- [ ] Digest binding: signed digest == pulled manifest digest (verify + recv --verify)
- [ ] Verify-before-materialize: temp file, fsync, atomic rename; no final output on failure
- [ ] Tag discovery primary; referrers only for bundle/receipt
- [ ] Clock skew +/-5 min for --since
- [ ] Tag grammar: [A-Za-z0-9_.-], max 128, start with [A-Za-z0-9_]
- [ ] Empty config descriptor canonical
- [ ] Chunk format: tar with chunk_<index>.bin
- [ ] Bundle attached with tag AND referrers
- [ ] Resume via HEAD (skip upload if blob exists)
- [ ] Dedup by message id (unordered tag listing)

## Static Analysis

```bash
golangci-lint run ./...
```

Must be clean with errcheck enabled.

## Tests

```bash
go test ./...
go test -race ./...
```

## Coverage (sensitive packages)

```bash
make coverage-gate
```

Enforces per-package thresholds: pkg/crypto >= 66%, pkg/transfer >= 36%, pkg/oci >= 54%.

## Smoke Test

Document at least one end-to-end run:

1. send a file to a registry
2. recv the file
3. verify the digest

## Integration Tests (optional)

```bash
# GHCR round-trip (requires DOCKERCOMMS_IT_GHCR_REPO, DOCKERCOMMS_IT_RECIPIENT)
DOCKERCOMMS_IT_GHCR_REPO=ghcr.io/user/repo DOCKERCOMMS_IT_RECIPIENT=alice@example.com go test -tags=integration ./test/integration/...

# Docker Hub tag listing (requires DOCKERCOMMS_IT_DH_REPO)
DOCKERCOMMS_IT_DH_REPO=docker.io/user/repo go test -tags=integration ./test/integration/...
```

1GB file round-trip: manual only; create 1GB file and run send/recv/verify. Document in release notes.

## Build

```bash
make build
./dockercomms version
./dockercomms send --help
```
