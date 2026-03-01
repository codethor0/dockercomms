// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"bytes"
	"context"
	"io"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry/remote"
)

// Client wraps oras-go remote repository for OCI operations.
type Client struct {
	repo *remote.Repository
}

// NewClient creates an OCI client for the given registry/repo reference.
// Reference format: registry/repo (e.g. ghcr.io/user/repo).
func NewClient(reference string) (*Client, error) {
	repo, err := remote.NewRepository(reference)
	if err != nil {
		return nil, err
	}
	return &Client{repo: repo}, nil
}

// EnsureEmptyConfig pushes the canonical empty config blob if not present.
func (c *Client) EnsureEmptyConfig(ctx context.Context) error {
	exists, err := c.repo.Exists(ctx, EmptyConfigDescriptor)
	if err != nil || exists {
		return err
	}
	return c.repo.Push(ctx, EmptyConfigDescriptor, bytes.NewReader([]byte("{}")))
}

// Exists returns true if the blob exists in the repository.
func (c *Client) Exists(ctx context.Context, desc ocispec.Descriptor) (bool, error) {
	return c.repo.Exists(ctx, desc)
}

// PushBlob uploads a blob. Skips if Exists returns true (caller responsibility for resume).
func (c *Client) PushBlob(ctx context.Context, desc ocispec.Descriptor, content io.Reader) error {
	return c.repo.Push(ctx, desc, content)
}

// FetchBlob downloads a blob.
func (c *Client) FetchBlob(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	return c.repo.Fetch(ctx, desc)
}

// Resolve resolves a tag or digest to a manifest descriptor.
func (c *Client) Resolve(ctx context.Context, reference string) (ocispec.Descriptor, error) {
	return c.repo.Resolve(ctx, reference)
}

// FetchManifest fetches manifest by tag or digest.
func (c *Client) FetchManifest(ctx context.Context, reference string) (ocispec.Descriptor, io.ReadCloser, error) {
	return c.repo.FetchReference(ctx, reference)
}

// PushManifest pushes a manifest and tags it.
func (c *Client) PushManifest(ctx context.Context, desc ocispec.Descriptor, content io.Reader, tag string) error {
	return c.repo.PushReference(ctx, desc, content, tag)
}

// Tag tags an existing manifest.
func (c *Client) Tag(ctx context.Context, desc ocispec.Descriptor, tag string) error {
	return c.repo.Tag(ctx, desc, tag)
}

// Tags lists tags with pagination. fn is called for each page.
func (c *Client) Tags(ctx context.Context, last string, fn func(tags []string) error) error {
	return c.repo.Tags(ctx, last, fn)
}

// ReferrersSupported returns true if the registry supports the Referrers API.
// Best-effort: attempts a referrers call with a valid digest; on 404/unsupported returns false.
func (c *Client) ReferrersSupported(ctx context.Context, subject ocispec.Descriptor) bool {
	err := c.repo.Referrers(ctx, subject, "", func([]ocispec.Descriptor) error { return nil })
	return err == nil
}

// Referrers lists referrers for the given subject. artifactType filters by type (empty = all).
func (c *Client) Referrers(ctx context.Context, subject ocispec.Descriptor, artifactType string, fn func([]ocispec.Descriptor) error) error {
	return c.repo.Referrers(ctx, subject, artifactType, fn)
}
