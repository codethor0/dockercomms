# Architecture and flows

DockerComms uses standard OCI registries as the transport. Blobs hold chunked payloads; manifests and tags carry metadata; Cosign bundles provide signatures verified before any payload is written to its final path.

See [ARCH.md](../ARCH.md) for component layout and [SPEC.md](../SPEC.md) for protocol details.

## High-level architecture

```mermaid
flowchart LR
  subgraph client [DockerComms CLI]
    send[send]
    recv[recv]
    verify[verify]
    ack[ack]
  end
  subgraph registry [OCI registry]
    blobs[(Content blobs)]
    manifests[(Manifests and tags)]
    bundles[(Signature bundles)]
  end
  send --> blobs
  send --> manifests
  send --> bundles
  recv --> manifests
  recv --> blobs
  recv --> bundles
  verify --> bundles
  verify --> manifests
  ack --> bundles
```

## Send flow

```mermaid
sequenceDiagram
  participant U as Operator
  participant S as send
  participant R as OCI registry
  U->>S: file path, repo, recipient
  S->>S: chunk, compress, digest
  S->>R: HEAD missing blobs
  S->>R: PUT blobs
  S->>R: PUT manifest, tag inbox-...
  S->>R: sign and attach bundle
  S-->>U: exit 0 or error code
```

## Receive and verify-before-materialize

```mermaid
sequenceDiagram
  participant U as Operator
  participant C as recv
  participant R as OCI registry
  participant FS as Local filesystem
  U->>C: repo, recipient, out dir
  C->>R: list inbox tags
  C->>R: pull manifest and bundle
  C->>C: verify bundle and digest match
  alt verification fails
    C-->>U: exit 2, no final write
  else verification ok
    C->>R: fetch chunk layers
    C->>C: reassemble to temp file
    C->>FS: fsync and atomic rename
    C-->>U: exit 0
  end
```

## Trust and verification decision

```mermaid
flowchart TD
  start([Resolve artifact]) --> bundle{Bundle found?}
  bundle -->|no| nf[Exit 5 not found]
  bundle -->|yes| sig[Verify with Sigstore policy]
  sig -->|fail| vfail[Exit 2 verification failed]
  sig -->|ok| digest{Signed digest equals manifest digest?}
  digest -->|no| vfail
  digest -->|yes| mat{Materialize payload?}
  mat -->|recv path| write[Temp file then atomic rename]
  mat -->|verify only| done[Exit 0]
  write --> done
```

## Release scope (v1.0 line)

```mermaid
flowchart LR
  subgraph in_scope [In scope for source release]
    cli[CLI send recv verify ack]
    oci[OCI push pull tags]
    crypto[Sigstore verification]
    hardening[Path limits and exit codes]
  end
  subgraph operator [Operator-provided]
    reg[Registry credentials]
    cosign[Cosign keyless or keys]
    trust[Trusted root / TUF]
  end
  cli --> reg
  cli --> cosign
  crypto --> trust
```
