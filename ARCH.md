# DockerComms Architecture

## Components

- **cmd/dockercomms**: CLI entrypoint
- **pkg/cli**: Cobra commands and flag handling
- **pkg/config**: Configuration loading
- **pkg/oci**: OCI registry client (oras-go/v2 wrapper)
- **pkg/crypto**: Sigstore verification (sigstore-go)
- **pkg/transfer**: Chunking, compression, reassembly

## Data Flow

### Send

1. Read file, stream into chunks (tar+gzip or tar+zstd)
2. Compute digests, build manifest
3. HEAD each blob; upload missing blobs
4. Push manifest, tag with inbox-...
5. Sign with cosign (os/exec), attach bundle as artifact + tag

### Receive

1. List tags with prefix inbox-<recipient_tag>-
2. Pull manifest, validate annotations, apply --since filter
3. Fetch bundle (referrers or tag fallback)
4. Verify with sigstore-go; ensure signed digest matches manifest
5. Download layers, reassemble by chunk index
6. Write to temp file, fsync, atomic rename
7. Optionally write receipt artifact

### Verify

1. Resolve bundle (auto: referrers then tag; or explicit mode)
2. Verify bundle with sigstore-go policy
3. Ensure signed digest equals given digest

## Security

- Verify-before-materialize: never write payload until verification succeeds
- Path traversal defense: filename is basename only
- Constant-time comparisons where applicable
- Hard limits on chunks and total size
