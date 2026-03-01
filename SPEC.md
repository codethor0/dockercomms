# DockerComms Protocol Specification v1.0

## Scope

DockerComms v1.0 provides secure file transport over OCI registries. It is not a messaging system; it transports files as OCI artifacts with signing and verification.

## Artifact Types

- `application/vnd.dockercomms.message.v1` - Message artifact
- `application/vnd.dockercomms.receipt.v1` - Receipt artifact
- `application/vnd.dev.sigstore.bundle.v0.3+json` - Sigstore bundle (configurable)

## Media Types

- Manifest: `application/vnd.oci.image.manifest.v1+json`
- Chunk layer (gzip): `application/vnd.dockercomms.chunk.v1.tar+gzip`
- Chunk layer (zstd): `application/vnd.dockercomms.chunk.v1.tar+zstd`

## Required Annotations (All Artifacts)

- `dockercomms.version` = 1.0
- `dockercomms.sender`
- `dockercomms.recipient`
- `dockercomms.created_at` (RFC3339)
- `dockercomms.ttl_seconds` (int)
- `dockercomms.filename` (basename only)
- `dockercomms.total_bytes`
- `dockercomms.chunk_bytes`
- `dockercomms.chunk_count`
- `dockercomms.compression` = gzip | zstd

## Message-Only Annotations

- `dockercomms.message.id` (UUID)

## Receipt Annotations

- `dockercomms.receipt.for` (full digest sha256:...)
- `dockercomms.receipt.status` = accepted | rejected
- `dockercomms.receipt.verified` = true | false
- `dockercomms.receipt.reason` (optional)

## Tag Format

- Message: `inbox-<recipient_tag>-<YYYYMMDD>-<sid8>-<mid8>`
- Receipt: `receipt-<hexDigest12>-<ts>-<rand>`
- Bundle: `bundle-<hexDigest12>-<ts>-<rand>`

`recipient_tag` = lower(base32hex(sha256(recipient)))[:26]

`hexDigest12` = first 12 hex chars of manifest digest without "sha256:"

## Empty Config Descriptor

- mediaType: `application/vnd.oci.empty.v1+json`
- digest: `sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a`
- size: 2
- content: {}

## Chunk Layer Format

Each layer is a tar stream with exactly one file: `chunk_<index>.bin`

Descriptor annotation: `dockercomms.chunk.index` = 0..N-1
