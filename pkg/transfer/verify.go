// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/codethor0/dockercomms/pkg/crypto"
	"github.com/codethor0/dockercomms/pkg/oci"
)

// bundleStore is the minimal interface for fetching bundles (referrers or tag fallback).
type bundleStore interface {
	Tags(ctx context.Context, last string, fn func(tags []string) error) error
	FetchManifest(ctx context.Context, reference string) (ocispec.Descriptor, io.ReadCloser, error)
	FetchBlob(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error)
	Referrers(ctx context.Context, subject ocispec.Descriptor, artifactType string, fn func([]ocispec.Descriptor) error) error
}

// FetchBundleForVerify fetches bundle using auto: referrers first, then tag fallback.
func FetchBundleForVerify(ctx context.Context, client bundleStore, repo, digestRef string) ([]byte, error) {
	b, err := FetchBundleReferrers(ctx, client, repo, digestRef)
	if err == nil {
		return b, nil
	}
	return FetchBundleTag(ctx, client, digestRef)
}

// FetchBundleReferrers fetches bundle via referrers API.
func FetchBundleReferrers(ctx context.Context, client bundleStore, repo, digestRef string) ([]byte, error) {
	subject := ocispec.Descriptor{
		MediaType: oci.MediaTypeManifest,
		Digest:    digest.Digest(digestRef),
	}
	var bundleDesc *ocispec.Descriptor
	if err := client.Referrers(ctx, subject, oci.ArtifactTypeBundle, func(refs []ocispec.Descriptor) error {
		if len(refs) > 0 {
			bundleDesc = &refs[0]
		}
		return nil
	}); err != nil || bundleDesc == nil {
		if err != nil {
			return nil, fmt.Errorf("referrers: %w", err)
		}
		return nil, fmt.Errorf("no bundle referrers found")
	}
	manifestRef := repo + "@" + bundleDesc.Digest.String()
	_, rc, err := client.FetchManifest(ctx, manifestRef)
	if err != nil {
		return nil, err
	}
	manifestBytes, err := io.ReadAll(rc)
	if cerr := rc.Close(); cerr != nil && err == nil {
		err = cerr
	}
	if err != nil {
		return nil, err
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, err
	}
	if len(manifest.Layers) == 0 {
		return nil, fmt.Errorf("bundle manifest has no layers")
	}
	blobRC, err := client.FetchBlob(ctx, manifest.Layers[0])
	if err != nil {
		return nil, err
	}
	bundleBytes, err := io.ReadAll(blobRC)
	if cerr := blobRC.Close(); cerr != nil && err == nil {
		err = cerr
	}
	if err != nil {
		return nil, err
	}
	return bundleBytes, nil
}

// FetchBundleTag fetches bundle by searching tags with bundle-<hex12>- prefix.
func FetchBundleTag(ctx context.Context, client bundleStore, digestRef string) ([]byte, error) {
	hex12 := HexDigest12(digestRef)
	prefix := "bundle-" + hex12 + "-"
	var foundTag string
	if err := client.Tags(ctx, "", func(tags []string) error {
		for _, t := range tags {
			if strings.HasPrefix(t, prefix) {
				foundTag = t
				return nil
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	if foundTag == "" {
		return nil, fmt.Errorf("bundle tag not found")
	}
	_, rc, err := client.FetchManifest(ctx, foundTag)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(rc)
	if cerr := rc.Close(); cerr != nil && err == nil {
		err = cerr
	}
	if err != nil {
		return nil, err
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(b, &manifest); err != nil {
		return nil, err
	}
	if len(manifest.Layers) == 0 {
		return nil, fmt.Errorf("bundle manifest has no layers")
	}
	blobRC, err := client.FetchBlob(ctx, manifest.Layers[0])
	if err != nil {
		return nil, err
	}
	bundleBytes, err := io.ReadAll(blobRC)
	if cerr := blobRC.Close(); cerr != nil && err == nil {
		err = cerr
	}
	if err != nil {
		return nil, err
	}
	return bundleBytes, nil
}

// VerifyBundleInProcess verifies bundle bytes with sigstore-go.
func VerifyBundleInProcess(bundleBytes []byte, digestRef string, trustedRoot string) error {
	return crypto.VerifyBundleBytes(bundleBytes, digestRef, trustedRoot)
}
