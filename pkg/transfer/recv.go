// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/codethor0/dockercomms/pkg/crypto"
	"github.com/codethor0/dockercomms/pkg/oci"
)

// SinceFilterPass returns true if a message with the given created_at annotation
// should be included given the since filter (clock skew tolerance 5 minutes).
func SinceFilterPass(createdStr string, sinceTime time.Time) bool {
	if createdStr == "" {
		return false
	}
	created, err := time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return true
	}
	return sinceTime.Sub(created) <= 5*time.Minute
}

// RecvOptions configures the receive operation.
type RecvOptions struct {
	Repo         string
	Me           string
	Out          string
	Since        string
	Max          int
	Verify       bool
	WriteReceipt bool
	Policy       string
	TrustedRoot  string
}

// Recv discovers messages, verifies, and materializes files.
func Recv(ctx context.Context, opts RecvOptions) (int, error) {
	recipientTag := RecipientTag(opts.Me)
	prefix := "inbox-" + recipientTag + "-"
	client, err := oci.NewClient(opts.Repo)
	if err != nil {
		return 0, err
	}

	var tags []string
	err = client.Tags(ctx, "", func(page []string) error {
		for _, t := range page {
			if strings.HasPrefix(t, prefix) {
				tags = append(tags, t)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	var sinceTime *time.Time
	if opts.Since != "" {
		t, err := time.Parse(time.RFC3339, opts.Since)
		if err != nil {
			return 0, fmt.Errorf("parse --since: %w", err)
		}
		sinceTime = &t
	}

	seen := make(map[string]bool)
	var messages []messageInfo
	for _, tag := range tags {
		desc, rc, err := client.FetchManifest(ctx, tag)
		if err != nil {
			continue
		}
		manifestBytes, err := io.ReadAll(rc)
		if cerr := rc.Close(); cerr != nil && err == nil {
			err = cerr
		}
		if err != nil {
			continue
		}
		var manifest ocispec.Manifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			continue
		}
		msgID := manifest.Annotations["dockercomms.message.id"]
		if msgID == "" || seen[msgID] {
			continue
		}
		seen[msgID] = true
		recipient := manifest.Annotations["dockercomms.recipient"]
		if recipient != opts.Me {
			continue
		}
		if sinceTime != nil {
			if !SinceFilterPass(manifest.Annotations["dockercomms.created_at"], *sinceTime) {
				continue
			}
		}
		messages = append(messages, messageInfo{
			Tag:     tag,
			Digest:  desc.Digest.String(),
			Layers:  manifest.Layers,
			Ann:     manifest.Annotations,
		})
		if len(messages) >= opts.Max {
			break
		}
	}

	count := 0
	for _, m := range messages {
		if err := recvOne(ctx, client, m, opts); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

type messageInfo struct {
	Tag    string
	Digest string
	Layers []ocispec.Descriptor
	Ann    map[string]string
}

func recvOne(ctx context.Context, client *oci.Client, m messageInfo, opts RecvOptions) error {
	if opts.Verify {
		bundleBytes, err := fetchBundle(ctx, client, m.Digest, opts.Repo)
		if err != nil {
			return fmt.Errorf("fetch bundle: %w", err)
		}
		if err := crypto.VerifyBundleBytes(bundleBytes, m.Digest, opts.TrustedRoot); err != nil {
			return fmt.Errorf("verification failed: %w", err)
		}
	}

	fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) {
		return client.FetchBlob(ctx, d)
	}
	sort.Slice(m.Layers, func(i, j int) bool {
		ai, _ := strconv.Atoi(m.Layers[i].Annotations["dockercomms.chunk.index"])
		aj, _ := strconv.Atoi(m.Layers[j].Annotations["dockercomms.chunk.index"])
		return ai < aj
	})

	tmpPath := filepath.Join(opts.Out, ".tmp-"+SanitizeFilename(m.Ann["dockercomms.filename"])+"-"+HexDigest12(m.Digest))
	outPath := filepath.Join(opts.Out, SanitizeFilename(m.Ann["dockercomms.filename"]))
	f, err := os.Create(tmpPath) // #nosec G304 -- tmpPath constructed from sanitized filename in opts.Out
	if err != nil {
		return err
	}
	totalBytes, err := Reassemble(m.Layers, fetch, f)
	if err != nil {
		if cerr := f.Close(); cerr != nil {
			err = fmt.Errorf("%w (close: %v)", err, cerr)
		}
		if rerr := os.Remove(tmpPath); rerr != nil {
			err = fmt.Errorf("%w (remove: %v)", err, rerr)
		}
		return err
	}
	if err := f.Sync(); err != nil {
		if cerr := f.Close(); cerr != nil {
			err = fmt.Errorf("%w (close: %v)", err, cerr)
		}
		if rerr := os.Remove(tmpPath); rerr != nil {
			err = fmt.Errorf("%w (remove: %v)", err, rerr)
		}
		return err
	}
	if err := f.Close(); err != nil {
		if rerr := os.Remove(tmpPath); rerr != nil {
			err = fmt.Errorf("%w (remove: %v)", err, rerr)
		}
		return err
	}
	expectedStr := m.Ann["dockercomms.total_bytes"]
	if expectedStr != "" {
		expected, _ := strconv.ParseInt(expectedStr, 10, 64)
		if totalBytes != expected {
			if rerr := os.Remove(tmpPath); rerr != nil {
				return fmt.Errorf("total_bytes mismatch: got %d want %d (remove: %v)", totalBytes, expected, rerr)
			}
			return fmt.Errorf("total_bytes mismatch: got %d want %d", totalBytes, expected)
		}
	}
	if err := os.Rename(tmpPath, outPath); err != nil {
		if rerr := os.Remove(tmpPath); rerr != nil {
			return fmt.Errorf("%w (remove: %v)", err, rerr)
		}
		return err
	}

	if opts.WriteReceipt {
		if err := writeReceipt(ctx, client, opts.Repo, m, opts.Verify); err != nil {
			return fmt.Errorf("write receipt: %w", err)
		}
	}
	return nil
}

func fetchBundle(ctx context.Context, client *oci.Client, digestRef string, repo string) ([]byte, error) {
	hex12 := HexDigest12(digestRef)
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
	}); err == nil && bundleDesc != nil {
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
		layer := manifest.Layers[0]
		blobRC, err := client.FetchBlob(ctx, layer)
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
	// Tag fallback
	{
		var foundTag string
		if err := client.Tags(ctx, "", func(tags []string) error {
			for _, t := range tags {
				if strings.HasPrefix(t, "bundle-"+hex12+"-") {
					foundTag = t
					return nil
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}
		if foundTag == "" {
			return nil, fmt.Errorf("bundle not found")
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
		layer := manifest.Layers[0]
		blobRC, err := client.FetchBlob(ctx, layer)
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
}

func writeReceipt(ctx context.Context, client *oci.Client, _ string, m messageInfo, verified bool) error {
	status := "accepted"
	ann := map[string]string{
		"dockercomms.version":        "1.0",
		"dockercomms.sender":         m.Ann["dockercomms.sender"],
		"dockercomms.recipient":      m.Ann["dockercomms.recipient"],
		"dockercomms.created_at":     time.Now().UTC().Format(time.RFC3339),
		"dockercomms.ttl_seconds":    m.Ann["dockercomms.ttl_seconds"],
		"dockercomms.receipt.for":    m.Digest,
		"dockercomms.receipt.status": status,
		"dockercomms.receipt.verified": strconv.FormatBool(verified),
	}
	builder := oci.NewReceiptManifest(ann)
	manifestJSON, desc, err := builder.Build()
	if err != nil {
		return err
	}
	if err := client.EnsureEmptyConfig(ctx); err != nil {
		return fmt.Errorf("ensure empty config: %w", err)
	}
	ts := time.Now().Unix()
	randPart := truncateUUID(uuid.New().String(), 6)
	receiptTag := fmt.Sprintf("receipt-%s-%d-%s", HexDigest12(m.Digest), ts, randPart)
	if err := client.PushManifest(ctx, desc, bytes.NewReader(manifestJSON), receiptTag); err != nil {
		return fmt.Errorf("push receipt: %w", err)
	}
	return nil
}
