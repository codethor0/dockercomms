// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"crypto/sha256"
	"encoding/base32"
	"strings"
)

// RecipientTag computes the tag-safe token for a recipient identity.
// Formula: lower(base32hex(sha256(recipient)))[:26]
// Allowed chars in tag: [A-Za-z0-9_.-], must start with [A-Za-z0-9_].
// base32hex uses 0-9 and a-v, so lower() gives valid start.
func RecipientTag(recipient string) string {
	h := sha256.Sum256([]byte(recipient))
	encoded := base32.HexEncoding.WithPadding(base32.NoPadding).EncodeToString(h[:])
	return strings.ToLower(encoded)[:26]
}

// HexDigest12 returns the first 12 hex chars of a digest without the "sha256:" prefix.
func HexDigest12(digest string) string {
	if len(digest) > 7 && digest[:7] == "sha256:" {
		digest = digest[7:]
	}
	if len(digest) > 12 {
		return digest[:12]
	}
	return digest
}
