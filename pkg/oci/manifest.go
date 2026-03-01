// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"encoding/json"
	"fmt"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ManifestBuilder builds OCI manifest with DockerComms annotations.
type ManifestBuilder struct {
	ArtifactType string
	Subject      *ocispec.Descriptor
	Annotations  map[string]string
	Layers       []ocispec.Descriptor
}

// NewMessageManifest creates a manifest builder for a message artifact.
func NewMessageManifest(annotations map[string]string, layers []ocispec.Descriptor) *ManifestBuilder {
	return &ManifestBuilder{
		ArtifactType: ArtifactTypeMessage,
		Annotations:  annotations,
		Layers:       layers,
	}
}

// NewReceiptManifest creates a manifest builder for a receipt artifact.
func NewReceiptManifest(annotations map[string]string) *ManifestBuilder {
	return &ManifestBuilder{
		ArtifactType: ArtifactTypeReceipt,
		Annotations:  annotations,
		Layers:      nil,
	}
}

// NewBundleManifest creates a manifest builder for a bundle artifact with subject for referrers.
func NewBundleManifest(subject ocispec.Descriptor, bundleAnnotations map[string]string, layer ocispec.Descriptor) *ManifestBuilder {
	ann := make(map[string]string)
	for k, v := range bundleAnnotations {
		ann[k] = v
	}
	subj := subject
	subj.MediaType = MediaTypeManifest
	return &ManifestBuilder{
		ArtifactType: ArtifactTypeBundle,
		Subject:      &subj,
		Annotations:  ann,
		Layers:       []ocispec.Descriptor{layer},
	}
}

// Build produces the OCI manifest JSON and descriptor.
func (b *ManifestBuilder) Build() ([]byte, ocispec.Descriptor, error) {
	manifest := ocispec.Manifest{
		Versioned:    specs.Versioned{SchemaVersion: 2},
		ArtifactType: b.ArtifactType,
		Config:       EmptyConfigDescriptor,
		Subject:      b.Subject,
		Layers:       b.Layers,
		Annotations:  make(map[string]string),
	}
	for k, v := range b.Annotations {
		manifest.Annotations[k] = v
	}

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return nil, ocispec.Descriptor{}, fmt.Errorf("marshal manifest: %w", err)
	}

	d := digest.FromBytes(manifestJSON)
	desc := ocispec.Descriptor{
		MediaType:   MediaTypeManifest,
		Digest:      d,
		Size:        int64(len(manifestJSON)),
		Annotations: nil,
	}

	return manifestJSON, desc, nil
}
