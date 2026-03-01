# Hardening Sprint Report

## Commands and Outputs

### go test ./...

```
?   	github.com/dockercomms/dockercomms/cmd/dockercomms	[no test files]
ok  	github.com/dockercomms/dockercomms/internal/version	0.203s
?   	github.com/dockercomms/dockercomms/pkg/cli	[no test files]
?   	github.com/dockercomms/dockercomms/pkg/config	[no test files]
?   	github.com/dockercomms/dockercomms/pkg/crypto	[no test files]
ok  	github.com/dockercomms/dockercomms/pkg/oci	0.290s
ok  	github.com/dockercomms/dockercomms/pkg/transfer	0.449s
```

### go test -race ./...

```
ok  	github.com/dockercomms/dockercomms/internal/version	1.281s
ok  	github.com/dockercomms/dockercomms/pkg/oci	(cached)
ok  	github.com/dockercomms/dockercomms/pkg/transfer	2.896s
```

### golangci-lint run ./...

```
0 issues.
```

### Coverage (pkg/transfer, pkg/oci, internal/version)

- pkg/transfer: 20.6%
- pkg/oci: 24.3%
- total: 20.8%

Sensitive paths with high coverage: RecipientTag, HexDigest12, SanitizeFilename, Chunk, Reassemble, manifest Build.

## Completed Tasks

1. **errcheck**: Enabled; all 21 findings fixed. No suppressions.
2. **internal/version**: Implemented with Version, Commit, Date; `dockercomms version` command; Makefile -ldflags.
3. **Test hardening**: Verify-before-materialize (no output on failure, no temp leftover); chunk reassembly property test; tag grammar constraints.
4. **Spec consistency**: oras-go/v2 documented; referrers only for relations.
5. **RELEASE_CHECKLIST.md**: Added with stop-ship gates.

## Artifacts

- RELEASE_CHECKLIST.md
- internal/version/
- pkg/transfer/recv_test.go, chunker_property_test.go, tags_constraint_test.go
