// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/codethor0/dockercomms/pkg/oci"
)

// AckOptions configures the ack (receipt) operation.
type AckOptions struct {
	Repo     string
	For      string
	Status   string
	Verified bool
	Reason   string
}

// Ack writes a receipt artifact.
func Ack(ctx context.Context, opts AckOptions) error {
	client, err := oci.NewClient(opts.Repo)
	if err != nil {
		return err
	}
	if err := client.EnsureEmptyConfig(ctx); err != nil {
		return err
	}
	ann := map[string]string{
		"dockercomms.version":          "1.0",
		"dockercomms.sender":           "unknown",
		"dockercomms.recipient":        "unknown",
		"dockercomms.created_at":       time.Now().UTC().Format(time.RFC3339),
		"dockercomms.ttl_seconds":      "604800",
		"dockercomms.receipt.for":      opts.For,
		"dockercomms.receipt.status":   opts.Status,
		"dockercomms.receipt.verified": strconv.FormatBool(opts.Verified),
	}
	if opts.Reason != "" {
		ann["dockercomms.receipt.reason"] = opts.Reason
	}
	builder := oci.NewReceiptManifest(ann)
	manifestJSON, desc, err := builder.Build()
	if err != nil {
		return err
	}
	ts := time.Now().Unix()
	randPart := truncateUUID(uuid.New().String(), 6)
	receiptTag := fmt.Sprintf("receipt-%s-%d-%s", HexDigest12(opts.For), ts, randPart)
	return client.PushManifest(ctx, desc, bytes.NewReader(manifestJSON), receiptTag)
}
