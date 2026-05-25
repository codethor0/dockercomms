# Changelog

All notable changes to DockerComms are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0-rc4] - 2026-05-25

Source release candidate on current `main`. Not final GA: GHCR integration/signing proof (Section B) and full GitHub security UI closure (Section A) remain maintainer steps.

### Changed

- Public README: architecture diagrams on the landing page; logo uses transparent SVG
- Dependency updates: `github.com/klauspost/compress` v1.18.6, `github.com/in-toto/in-toto-golang` v0.11.0
- Maintainer tooling: git hooks and PR CI guard to block IDE co-author trailers on new commits

### Removed

- Internal operator prompts and non-public docs from the default tree (see `.gitignore`)

## [1.0.0-rc3] - 2026-03-21

Release candidate toward v1.0.0 final. See [RELEASE.md](RELEASE.md) for scope and validation notes.

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
