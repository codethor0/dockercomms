## Summary

What does this PR change and why?

## Testing

- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] `golangci-lint run ./...`
- [ ] `make coverage-gate`
- [ ] `./scripts/run-integration.sh --check` and `./scripts/login-and-run-integration.sh --check` (if touching scripts or registry behavior)

## Security / behavior

- [ ] No weakening of verify-before-materialize, path sanitization, or exit-code semantics
- [ ] If this touches crypto, registry, or trust: describe impact and what you ran (including integration/E2E if applicable)

## Docs

- [ ] README / SPEC / RELEASE docs updated if user-facing behavior changed
