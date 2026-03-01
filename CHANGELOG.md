# Changelog

All notable changes to DockerComms are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-03-01

### Added

- send: Push file as OCI artifact with chunking (gzip/zstd), cosign signing, bundle attachment
- recv: Discover inbox tags, verify bundle, reassemble, verify-before-materialize
- verify: Verify artifact digest using bundle (referrers or tag fallback)
- ack: Write receipt artifact
- OCI client (oras-go/v2): push, pull, tags, referrers, HEAD for resume
- Chunking: 100 MiB default, streaming, tar+gzip/zstd
- Sigstore verification via sigstore-go
- Path traversal defense (SanitizeFilename)
- Tag encoding: RecipientTag, HexDigest12 with test vectors

### Security

- Verify-before-materialize: temp file, fsync, atomic rename
- Bundle verification with TUF or custom trusted root
- Hard limits on chunks and total size
