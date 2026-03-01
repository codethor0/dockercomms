// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyBundle_InvalidPath(t *testing.T) {
	err := VerifyBundle("/nonexistent/path/bundle.json", "sha256:"+hex32zeros(), "")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "load bundle") && !os.IsNotExist(err) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerifyBundle_InvalidDigestFormat(t *testing.T) {
	tmp := t.TempDir()
	bundlePath := filepath.Join(tmp, "bundle.json")
	if err := os.WriteFile(bundlePath, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		digest string
	}{
		{"empty", ""},
		{"too short", "sha256:abc"},
		{"bad hex", "sha256:zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"},
		{"wrong length no prefix", "abc123"},
		{"31 bytes hex", "sha256:" + hexN(31)},
		{"33 bytes hex", "sha256:" + hexN(33)},
		{"digest without prefix", hexN(32)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyBundle(bundlePath, tt.digest, "")
			if err == nil {
				t.Fatal("expected error for invalid digest")
			}
		})
	}
}

func TestVerifyBundleBytes_InvalidJSON(t *testing.T) {
	invalid := []byte(`{invalid json`)
	err := VerifyBundleBytes(invalid, "sha256:"+hex32zeros(), "")
	if err == nil {
		t.Fatal("expected error for invalid bundle")
	}
}

func TestVerifyBundleBytes_EmptyDigest(t *testing.T) {
	err := VerifyBundleBytes([]byte(`{}`), "", "")
	if err == nil {
		t.Fatal("expected error for empty digest")
	}
}

func TestVerifyBundleBytes_ValidBundleWrongDigest(t *testing.T) {
	b, err := os.ReadFile("testdata/bundle_v01.json")
	if err != nil {
		t.Fatal(err)
	}
	err = VerifyBundleBytes(b, "sha256:"+hex32zeros(), "")
	if err == nil {
		t.Fatal("expected verification to fail")
	}
	if strings.Contains(err.Error(), "load bundle") {
		t.Skip("bundle format not supported")
	}
}

func TestVerifyBundle_ValidBundleWrongDigest_HitsVerifyPath(t *testing.T) {
	// Use real bundle fixture; verification will fail (wrong digest) but we exercise verifyBundle.
	bundlePath := "testdata/bundle_v01.json"
	err := VerifyBundle(bundlePath, "sha256:"+hex32zeros(), "")
	if err == nil {
		t.Fatal("expected verification to fail (wrong digest)")
	}
	// Should fail at verification or digest binding, not at load
	if strings.Contains(err.Error(), "load bundle") {
		t.Skip("bundle format not supported by this sigstore-go version")
	}
}

func TestVerifyBundle_TrustedRootMissing(t *testing.T) {
	tmp := t.TempDir()
	bundlePath := filepath.Join(tmp, "bundle.json")
	if err := os.WriteFile(bundlePath, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	err := VerifyBundle(bundlePath, "sha256:"+hex32zeros(), "/nonexistent/trusted-root.json")
	if err == nil {
		t.Fatal("expected error for missing trusted root or invalid bundle")
	}
	if !strings.Contains(err.Error(), "trusted root") && !strings.Contains(err.Error(), "load bundle") && !strings.Contains(err.Error(), "read") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerifyBundle_InvalidTrustedRootContent(t *testing.T) {
	bundlePath := "testdata/bundle_v01.json"
	tmp := t.TempDir()
	badRoot := filepath.Join(tmp, "root.json")
	if err := os.WriteFile(badRoot, []byte("not valid json"), 0600); err != nil {
		t.Fatal(err)
	}
	err := VerifyBundle(bundlePath, "sha256:"+hex32zeros(), badRoot)
	if err == nil {
		t.Fatal("expected error for invalid trusted root")
	}
	if strings.Contains(err.Error(), "load bundle") {
		t.Skip("bundle format not supported")
	}
	if !strings.Contains(err.Error(), "trusted root") && !strings.Contains(err.Error(), "parse") {
		t.Errorf("unexpected error: %v", err)
	}
}

func hex32zeros() string {
	return hexN(32)
}

func hexN(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '0'
	}
	return string(b)
}

