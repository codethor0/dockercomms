// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/codethor0/dockercomms/pkg/oci"
)

type mockBundleStore struct {
	tags          []string
	manifestJSON  []byte
	bundleContent []byte
}

func (m *mockBundleStore) Tags(ctx context.Context, last string, fn func(tags []string) error) error {
	return fn(m.tags)
}

func (m *mockBundleStore) FetchManifest(ctx context.Context, ref string) (ocispec.Descriptor, io.ReadCloser, error) {
	return ocispec.Descriptor{}, io.NopCloser(bytes.NewReader(m.manifestJSON)), nil
}

func (m *mockBundleStore) FetchBlob(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(m.bundleContent)), nil
}

func (m *mockBundleStore) Referrers(ctx context.Context, subject ocispec.Descriptor, artifactType string, fn func([]ocispec.Descriptor) error) error {
	return nil
}

type mockBundleStoreWithReferrers struct {
	*mockBundleStore
	referrers []ocispec.Descriptor
}

func (m *mockBundleStoreWithReferrers) Referrers(ctx context.Context, subject ocispec.Descriptor, artifactType string, fn func([]ocispec.Descriptor) error) error {
	return fn(m.referrers)
}

type mockBundleStoreReferrersError struct {
	*mockBundleStore
}

func (m *mockBundleStoreReferrersError) Referrers(ctx context.Context, subject ocispec.Descriptor, artifactType string, fn func([]ocispec.Descriptor) error) error {
	return fmt.Errorf("referrers error")
}

type mockBundleStoreEmptyReferrers struct {
	*mockBundleStore
}

func (m *mockBundleStoreEmptyReferrers) Referrers(ctx context.Context, subject ocispec.Descriptor, artifactType string, fn func([]ocispec.Descriptor) error) error {
	return fn(nil)
}

func TestFetchBundleTag_Success(t *testing.T) {
	hex12 := HexDigest12("sha256:abc123def456789012345678901234567890123456789012345678901234")
	manifest := ocispec.Manifest{
		Layers: []ocispec.Descriptor{{Digest: "sha256:blob1", Size: 10}},
	}
	manifestJSON, _ := json.Marshal(manifest)
	mock := &mockBundleStore{
		tags:          []string{"bundle-" + hex12 + "-123456-xyz"},
		manifestJSON:  manifestJSON,
		bundleContent: []byte(`{"mediaType":"application/vnd.dev.sigstore.bundle+json"}`),
	}
	ctx := context.Background()
	b, err := FetchBundleTag(ctx, mock, "sha256:abc123def456789012345678901234567890123456789012345678901234")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, mock.bundleContent) {
		t.Errorf("got bundle %q", b)
	}
}

func TestFetchBundleForVerify_ReferrersFirst(t *testing.T) {
	hex12 := HexDigest12("sha256:abc123def456789012345678901234567890123456789012345678901234")
	manifest := ocispec.Manifest{
		Layers: []ocispec.Descriptor{{Digest: "sha256:blob1", Size: 10}},
	}
	manifestJSON, _ := json.Marshal(manifest)
	base := &mockBundleStore{
		tags:          []string{"bundle-" + hex12 + "-123456-xyz"},
		manifestJSON:  manifestJSON,
		bundleContent: []byte(`{"mediaType":"bundle"}`),
	}
	mock := &mockBundleStoreWithReferrers{
		mockBundleStore: base,
		referrers:       []ocispec.Descriptor{{Digest: "sha256:ref1", Size: 100}},
	}
	ctx := context.Background()
	b, err := FetchBundleForVerify(ctx, mock, "repo", "sha256:abc123def456789012345678901234567890123456789012345678901234")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, base.bundleContent) {
		t.Errorf("got bundle %q", b)
	}
}

func TestFetchBundleTag_NotFound(t *testing.T) {
	mock := &mockBundleStore{tags: []string{"other-tag"}}
	ctx := context.Background()
	_, err := FetchBundleTag(ctx, mock, "sha256:abc123def456789012345678901234567890123456789012345678901234")
	if err == nil {
		t.Fatal("expected error")
	}
	if err == nil || !strings.Contains(err.Error(), "bundle tag not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerifyBundleInProcess_InvalidDigest(t *testing.T) {
	err := VerifyBundleInProcess([]byte(`{}`), "", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyBundleInProcess_InvalidBundle(t *testing.T) {
	err := VerifyBundleInProcess([]byte(`{invalid`), "sha256:0000000000000000000000000000000000000000000000000000000000000000", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchBundleReferrers_EmptyReferrers(t *testing.T) {
	base := &mockBundleStore{tags: []string{"other-tag"}}
	mock := &mockBundleStoreEmptyReferrers{mockBundleStore: base}
	ctx := context.Background()
	_, err := FetchBundleReferrers(ctx, mock, "repo", "sha256:abc123def456789012345678901234567890123456789012345678901234")
	if err == nil {
		t.Fatal("expected error for empty referrers")
	}
	if err != nil && !strings.Contains(err.Error(), "no bundle referrers") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFetchBundleForVerify_ReferrersError(t *testing.T) {
	base := &mockBundleStore{tags: []string{"other-tag"}}
	mock := &mockBundleStoreReferrersError{mockBundleStore: base}
	ctx := context.Background()
	_, err := FetchBundleForVerify(ctx, mock, "repo", "sha256:abc123def456789012345678901234567890123456789012345678901234")
	if err == nil {
		t.Fatal("expected error when referrers fails and tag fallback finds no bundle")
	}
}

func TestFetchBundleForVerify_UnreachableRegistry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2)
	defer cancel()
	client, err := oci.NewClient("localhost:5000/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	_, err = FetchBundleForVerify(ctx, client, "localhost:5000/nonexistent", "sha256:abc123def456789012345678901234567890123456789012345678901234")
	if err == nil {
		t.Fatal("expected error (registry unreachable)")
	}
}
