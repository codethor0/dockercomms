// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	"github.com/sigstore/sigstore-go/pkg/verify"
)

// VerifyBundle verifies a Sigstore bundle against the expected digest.
// digestRef is the full digest string (e.g. "sha256:abc123...").
// trustedRootPath is optional; if empty, uses Sigstore TUF.
func VerifyBundle(bundlePath string, digestRef string, trustedRootPath string) error {
	b, err := bundle.LoadJSONFromPath(bundlePath)
	if err != nil {
		return fmt.Errorf("load bundle: %w", err)
	}
	return verifyBundle(b, digestRef, trustedRootPath)
}

// VerifyBundleBytes verifies bundle bytes against the expected digest.
func VerifyBundleBytes(bundleBytes []byte, digestRef string, trustedRootPath string) (err error) {
	tmp, err := os.CreateTemp("", "dockercomms-bundle-*.json")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	path := tmp.Name()
	defer func() {
		if rerr := os.Remove(path); rerr != nil && err == nil {
			err = fmt.Errorf("remove temp bundle: %w", rerr)
		}
	}()
	if _, err = tmp.Write(bundleBytes); err != nil {
		cerr := tmp.Close()
		if cerr != nil {
			return fmt.Errorf("write bundle: %w (close: %v)", err, cerr)
		}
		return fmt.Errorf("write bundle: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	err = VerifyBundle(path, digestRef, trustedRootPath)
	return
}

func verifyBundle(b *bundle.Bundle, digestRef string, trustedRootPath string) error {
	var trustedMaterial root.TrustedMaterial
	if trustedRootPath != "" {
		data, err := os.ReadFile(trustedRootPath) // #nosec G304 -- trustedRootPath from config, validated by caller
		if err != nil {
			return fmt.Errorf("read trusted root: %w", err)
		}
		trustedMaterial, err = root.NewTrustedRootFromJSON(data)
		if err != nil {
			return fmt.Errorf("parse trusted root: %w", err)
		}
	} else {
		opts := tuf.DefaultOptions()
		client, err := tuf.New(opts)
		if err != nil {
			return fmt.Errorf("tuf client: %w", err)
		}
		trustedMaterial, err = root.GetTrustedRoot(client)
		if err != nil {
			return fmt.Errorf("trusted root: %w", err)
		}
	}

	sev, err := verify.NewVerifier(trustedMaterial,
		verify.WithSignedCertificateTimestamps(1),
		verify.WithTransparencyLog(1),
		verify.WithObserverTimestamps(1))
	if err != nil {
		return fmt.Errorf("new verifier: %w", err)
	}

	digestHex := digestRef
	if len(digestHex) > 7 && digestHex[:7] == "sha256:" {
		digestHex = digestHex[7:]
	}
	digestBytes, err := hex.DecodeString(digestHex)
	if err != nil {
		return fmt.Errorf("decode digest: %w", err)
	}
	if len(digestBytes) != 32 {
		return fmt.Errorf("digest must be 32 bytes for sha256")
	}

	policy := verify.NewPolicy(
		verify.WithArtifactDigest("sha256", digestBytes),
		verify.WithoutIdentitiesUnsafe(),
	)

	_, err = sev.Verify(b, policy)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}
	return nil
}
