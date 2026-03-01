// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/codethor0/dockercomms/pkg/transfer"
)

const largePayloadBytes = 256 * 1024 * 1024 // 256 MiB

func TestGHCRRoundTrip_Smoke(t *testing.T) {
	repo := os.Getenv("DOCKERCOMMS_IT_GHCR_REPO")
	recipient := os.Getenv("DOCKERCOMMS_IT_RECIPIENT")
	outDir := os.Getenv("DOCKERCOMMS_IT_OUTDIR")
	if repo == "" || recipient == "" {
		t.Skip("DOCKERCOMMS_IT_GHCR_REPO and DOCKERCOMMS_IT_RECIPIENT required")
	}
	if outDir == "" {
		outDir = t.TempDir()
	}

	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "smoke.txt")
	if err := os.WriteFile(fpath, []byte("integration smoke test"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60)
	defer cancel()

	opts := transfer.SendOptions{
		Repo:       repo,
		Recipient:  recipient,
		ChunkBytes: 1024,
	}
	result, err := transfer.Send(ctx, fpath, opts)
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if result.Digest == "" {
		t.Fatal("expected digest")
	}

	recvOpts := transfer.RecvOptions{
		Repo:   repo,
		Me:     recipient,
		Out:    outDir,
		Verify: false,
		Max:    1,
	}
	n, err := transfer.Recv(ctx, recvOpts)
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if n < 1 {
		t.Fatalf("recv got %d messages", n)
	}
	_ = result
}

func TestGHCRRoundTrip_LargePayload(t *testing.T) {
	if os.Getenv("DOCKERCOMMS_IT_LARGE_PAYLOAD") == "" {
		t.Skip("DOCKERCOMMS_IT_LARGE_PAYLOAD required (set to 1 to enable 256 MiB round-trip)")
	}
	repo := os.Getenv("DOCKERCOMMS_IT_GHCR_REPO")
	recipient := os.Getenv("DOCKERCOMMS_IT_RECIPIENT")
	outDir := os.Getenv("DOCKERCOMMS_IT_OUTDIR")
	if repo == "" || recipient == "" {
		t.Skip("DOCKERCOMMS_IT_GHCR_REPO and DOCKERCOMMS_IT_RECIPIENT required")
	}
	if outDir == "" {
		outDir = t.TempDir()
	}

	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "large.bin")
	f, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	for written := int64(0); written < largePayloadBytes; written += 1024 * 1024 {
		chunk := make([]byte, 1024*1024)
		for i := range chunk {
			chunk[i] = byte(written + int64(i))
		}
		if _, err := f.Write(chunk); err != nil {
			f.Close()
			t.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 600)
	defer cancel()

	opts := transfer.SendOptions{
		Repo:       repo,
		Recipient:  recipient,
		ChunkBytes: 32 * 1024 * 1024,
	}
	result, err := transfer.Send(ctx, fpath, opts)
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if result.Digest == "" {
		t.Fatal("expected digest")
	}

	recvOpts := transfer.RecvOptions{
		Repo:   repo,
		Me:     recipient,
		Out:    outDir,
		Verify: false,
		Max:    1,
	}
	n, err := transfer.Recv(ctx, recvOpts)
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if n < 1 {
		t.Fatalf("recv got %d messages", n)
	}
	_ = result
}
