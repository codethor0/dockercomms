// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyBeforeMaterialize_NoOutputOnFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()
	outDir := t.TempDir()
	opts := RecvOptions{
		Repo:         "localhost:5000/test",
		Me:           "alice@example.com",
		Out:          outDir,
		Verify:       true,
		WriteReceipt: false,
		Max:          1,
	}
	_, err := Recv(ctx, opts)
	if err == nil {
		t.Fatal("expected error (registry unreachable)")
	}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("verify-before-materialize: on failure, out dir must be empty; got %d entries", len(entries))
	}
}

func TestVerifyBeforeMaterialize_NoTempLeftover(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()
	outDir := t.TempDir()
	opts := RecvOptions{
		Repo:         "localhost:5000/test",
		Me:           "alice@example.com",
		Out:          outDir,
		Verify:       true,
		WriteReceipt: false,
		Max:          1,
	}
	_, _ = Recv(ctx, opts)
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Base(e.Name()) != e.Name() {
			t.Errorf("unexpected path in out dir: %s", e.Name())
		}
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}
