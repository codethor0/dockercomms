// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"encoding/json"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestEmptyConfigDescriptor(t *testing.T) {
	if EmptyConfigDescriptor.Digest.String() != "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a" {
		t.Errorf("wrong empty config digest")
	}
	if EmptyConfigDescriptor.Size != 2 {
		t.Errorf("wrong empty config size")
	}
}

func TestManifestBuilder(t *testing.T) {
	ann := map[string]string{"dockercomms.version": "1.0"}
	layers := []ocispec.Descriptor{{Digest: "sha256:abc", Size: 10}}
	b := NewMessageManifest(ann, layers)
	manifestJSON, desc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if desc.MediaType != MediaTypeManifest {
		t.Errorf("media type = %q", desc.MediaType)
	}
	var m ocispec.Manifest
	if err := json.Unmarshal(manifestJSON, &m); err != nil {
		t.Fatal(err)
	}
	if m.ArtifactType != ArtifactTypeMessage {
		t.Errorf("artifact type = %q", m.ArtifactType)
	}
	if len(m.Layers) != 1 {
		t.Errorf("layers = %d", len(m.Layers))
	}
}

func TestReceiptManifestBuilder(t *testing.T) {
	ann := map[string]string{"dockercomms.version": "1.0", "dockercomms.receipt.for": "sha256:abc"}
	b := NewReceiptManifest(ann)
	manifestJSON, desc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if desc.MediaType != MediaTypeManifest {
		t.Errorf("media type = %q", desc.MediaType)
	}
	var m ocispec.Manifest
	if err := json.Unmarshal(manifestJSON, &m); err != nil {
		t.Fatal(err)
	}
	if m.ArtifactType != ArtifactTypeReceipt {
		t.Errorf("artifact type = %q", m.ArtifactType)
	}
	if len(m.Layers) != 0 {
		t.Errorf("receipt layers = %d, want 0", len(m.Layers))
	}
}

func TestBundleManifestBuilder(t *testing.T) {
	subject := ocispec.Descriptor{Digest: "sha256:abc123", Size: 100, MediaType: MediaTypeManifest}
	layer := ocispec.Descriptor{Digest: "sha256:bundleblob", Size: 200}
	b := NewBundleManifest(subject, map[string]string{"dockercomms.version": "1.0"}, layer)
	manifestJSON, desc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if desc.MediaType != MediaTypeManifest {
		t.Errorf("media type = %q", desc.MediaType)
	}
	var m ocispec.Manifest
	if err := json.Unmarshal(manifestJSON, &m); err != nil {
		t.Fatal(err)
	}
	if m.ArtifactType != ArtifactTypeBundle {
		t.Errorf("artifact type = %q", m.ArtifactType)
	}
	if m.Subject == nil || m.Subject.Digest != subject.Digest {
		t.Errorf("subject digest = %v", m.Subject)
	}
	if len(m.Layers) != 1 {
		t.Errorf("layers = %d", len(m.Layers))
	}
}

func TestNewClient_InvalidReference(t *testing.T) {
	_, err := NewClient("://invalid")
	if err == nil {
		t.Fatal("expected error for invalid reference")
	}
}

func TestNewClient_ValidReference(t *testing.T) {
	client, err := NewClient("localhost:5000/testrepo")
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestMessageManifest_EmptyLayers(t *testing.T) {
	ann := map[string]string{"dockercomms.version": "1.0"}
	b := NewMessageManifest(ann, nil)
	manifestJSON, desc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(manifestJSON) == 0 {
		t.Fatal("expected non-empty manifest")
	}
	if desc.Size != int64(len(manifestJSON)) {
		t.Errorf("desc size = %d, want %d", desc.Size, len(manifestJSON))
	}
}

func TestManifestBuilder_EmptyAnnotations(t *testing.T) {
	b := NewMessageManifest(nil, []ocispec.Descriptor{{Digest: "sha256:x", Size: 1}})
	manifestJSON, desc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	var m ocispec.Manifest
	if err := json.Unmarshal(manifestJSON, &m); err != nil {
		t.Fatal(err)
	}
	if m.ArtifactType != ArtifactTypeMessage {
		t.Errorf("artifact type = %q", m.ArtifactType)
	}
	if desc.Size != int64(len(manifestJSON)) {
		t.Errorf("desc size = %d, want %d", desc.Size, len(manifestJSON))
	}
}

func TestManifestBuilder_SubjectNil(t *testing.T) {
	b := NewMessageManifest(map[string]string{"k": "v"}, nil)
	if b.Subject != nil {
		t.Error("message manifest subject should be nil")
	}
	_, desc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if desc.Digest.String() == "" {
		t.Error("expected non-empty digest")
	}
}
