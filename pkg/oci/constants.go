// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Artifact and media types per SPEC.md.
const (
	ArtifactTypeMessage = "application/vnd.dockercomms.message.v1"
	ArtifactTypeReceipt = "application/vnd.dockercomms.receipt.v1"
	ArtifactTypeBundle  = "application/vnd.dev.sigstore.bundle.v0.3+json"

	MediaTypeManifest     = ocispec.MediaTypeImageManifest
	MediaTypeChunkGzip    = "application/vnd.dockercomms.chunk.v1.tar+gzip"
	MediaTypeChunkZstd    = "application/vnd.dockercomms.chunk.v1.tar+zstd"
	MediaTypeEmptyConfig  = "application/vnd.oci.empty.v1+json"
	MediaTypeBundleConfig = ArtifactTypeBundle
)

// Empty config descriptor per spec (J).
var EmptyConfigDescriptor = ocispec.Descriptor{
	MediaType: MediaTypeEmptyConfig,
	Digest:    digest.Digest("sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"),
	Size:      2,
}
