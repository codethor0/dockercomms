// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/codethor0/dockercomms/pkg/oci"
)

// SendOptions configures the send operation.
type SendOptions struct {
	Repo       string
	Recipient  string
	Sender     string
	Session    string
	ChunkBytes int64
	Parallel   int
	Compress   string
	TTLSeconds int
	Sign       bool
	CosignPath string
	Identity   string
}

// SendResult holds the result of a send.
type SendResult struct {
	Digest string
	Tag    string
	Bundle string
}

// Send pushes a file to the registry as an OCI artifact.
func Send(ctx context.Context, filePath string, opts SendOptions) (*SendResult, error) {
	recipientTag := RecipientTag(opts.Recipient)
	msgID := uuid.New().String()
	sessionID := opts.Session
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	sid8 := truncateUUID(sessionID, 8)
	mid8 := truncateUUID(msgID, 8)
	now := time.Now().UTC()
	dateStr := now.Format("20060102")

	comp := CompressionGzip
	if opts.Compress == "zstd" {
		comp = CompressionZstd
	}
	chunker := NewChunker(opts.ChunkBytes, comp)
	descs, readers, totalBytes, err := chunker.ChunkFile(filePath)
	if err != nil {
		return nil, err
	}
	chunkBytes := opts.ChunkBytes
	if totalBytes < chunkBytes {
		chunkBytes = totalBytes
	}

	client, err := oci.NewClient(opts.Repo)
	if err != nil {
		return nil, err
	}
	if err := client.EnsureEmptyConfig(ctx); err != nil {
		return nil, err
	}

	// Upload blobs (resume: skip if exists)
	layerDescs := make([]ocispec.Descriptor, len(descs))
	for i, cd := range descs {
		exists, err := client.Exists(ctx, cd.Descriptor)
		if err != nil {
			return nil, err
		}
		if !exists {
			if err := client.PushBlob(ctx, cd.Descriptor, readers[i]); err != nil {
				return nil, err
			}
		}
		layerDescs[i] = cd.Descriptor
	}

	filename := SanitizeFilename(filePath)
	sender := opts.Sender
	if sender == "" {
		sender = opts.Identity
	}
	if sender == "" {
		sender = "unknown"
	}
	annotations := map[string]string{
		"dockercomms.version":     "1.0",
		"dockercomms.sender":     sender,
		"dockercomms.recipient":   opts.Recipient,
		"dockercomms.created_at": now.Format(time.RFC3339),
		"dockercomms.ttl_seconds": fmt.Sprintf("%d", opts.TTLSeconds),
		"dockercomms.message.id":  msgID,
		"dockercomms.filename":    filename,
		"dockercomms.total_bytes": fmt.Sprintf("%d", totalBytes),
		"dockercomms.chunk_bytes": fmt.Sprintf("%d", chunkBytes),
		"dockercomms.chunk_count": fmt.Sprintf("%d", len(descs)),
		"dockercomms.compression": string(comp),
	}
	builder := oci.NewMessageManifest(annotations, layerDescs)
	manifestJSON, manifestDesc, err := builder.Build()
	if err != nil {
		return nil, err
	}

	msgTag := fmt.Sprintf("inbox-%s-%s-%s-%s", recipientTag, dateStr, sid8, mid8)
	if err := client.PushManifest(ctx, manifestDesc, bytes.NewReader(manifestJSON), msgTag); err != nil {
		return nil, err
	}

	result := &SendResult{
		Digest: manifestDesc.Digest.String(),
		Tag:    msgTag,
	}

	if opts.Sign {
		bundlePath, err := signWithCosign(ctx, opts.CosignPath, opts.Repo, manifestDesc.Digest.String(), opts.Identity)
		if err != nil {
			return nil, err
		}
		result.Bundle = bundlePath
		// Attach bundle as OCI artifact
		bundleBytes, err := os.ReadFile(bundlePath) // #nosec G304 -- bundlePath from CreateTemp
		if err != nil {
			return nil, err
		}
		bundleDigest := digest.FromBytes(bundleBytes)
		bundleDesc := ocispec.Descriptor{
			MediaType: oci.ArtifactTypeBundle,
			Digest:    bundleDigest,
			Size:      int64(len(bundleBytes)),
		}
		if err := client.PushBlob(ctx, bundleDesc, bytes.NewReader(bundleBytes)); err != nil {
			return nil, err
		}
		subject := ocispec.Descriptor{
			MediaType: oci.MediaTypeManifest,
			Digest:    manifestDesc.Digest,
			Size:      manifestDesc.Size,
		}
		hex12 := HexDigest12(manifestDesc.Digest.String())
		ts := now.Unix()
		randPart := truncateUUID(uuid.New().String(), 6)
		bundleTag := fmt.Sprintf("bundle-%s-%d-%s", hex12, ts, randPart)
		bundleAnn := map[string]string{
			"dockercomms.version": "1.0",
		}
		bundleBuilder := oci.NewBundleManifest(subject, bundleAnn, bundleDesc)
		bundleManifestJSON, bundleManifestDesc, err := bundleBuilder.Build()
		if err != nil {
			return nil, err
		}
		if err := client.PushManifest(ctx, bundleManifestDesc, bytes.NewReader(bundleManifestJSON), bundleTag); err != nil {
			return nil, err
		}
		_ = os.Remove(bundlePath) // best-effort; temp dir purged by OS
	}

	return result, nil
}

func truncateUUID(s string, n int) string {
	s = strings.ReplaceAll(s, "-", "")
	if len(s) > n {
		return s[:n]
	}
	return s
}

func signWithCosign(ctx context.Context, cosignPath, repo, digestRef, identity string) (string, error) {
	tmp, err := os.CreateTemp("", "dockercomms-bundle-*.json")
	if err != nil {
		return "", err
	}
	bundlePath := tmp.Name()
	if err := tmp.Close(); err != nil {
		if rerr := os.Remove(bundlePath); rerr != nil { // #nosec G703 -- bundlePath from CreateTemp
			return "", fmt.Errorf("close temp bundle: %w (remove: %v)", err, rerr)
		}
		return "", fmt.Errorf("close temp bundle: %w", err)
	}
	ref := fmt.Sprintf("%s@%s", repo, digestRef)
	args := []string{"sign", ref, "--bundle", bundlePath}
	if identity != "" {
		args = append(args, "--identity", identity)
	}
	cmd := exec.CommandContext(ctx, cosignPath, args...) // #nosec G204 G702 -- cosignPath from config
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if rerr := os.Remove(bundlePath); rerr != nil { // #nosec G703 -- bundlePath from CreateTemp
			return "", fmt.Errorf("cosign sign: %w (remove: %v)", err, rerr)
		}
		return "", fmt.Errorf("cosign sign: %w", err)
	}
	return bundlePath, nil
}
