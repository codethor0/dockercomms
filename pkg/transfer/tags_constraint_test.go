// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"regexp"
	"testing"
)

func TestRecipientTag_LengthAndCharset(t *testing.T) {
	tagSafeRe := regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)
	tests := []string{"a", "alice@example.com", "bob", ""}
	for _, r := range tests {
		got := RecipientTag(r)
		if len(got) != 26 {
			t.Errorf("RecipientTag(%q) length = %d, want 26", r, len(got))
		}
		if !tagSafeRe.MatchString(got) {
			t.Errorf("RecipientTag(%q) = %q, not tag-safe", r, got)
		}
	}
}

func TestInboxTagFormat_LengthConstraint(t *testing.T) {
	tagSafeRe := regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)
	rt := RecipientTag("alice@example.com")
	tag := "inbox-" + rt + "-20250101-abc12345-def67890"
	if len(tag) > 128 {
		t.Errorf("inbox tag length %d exceeds 128", len(tag))
	}
	if !tagSafeRe.MatchString(tag) {
		t.Errorf("inbox tag %q not tag-safe", tag)
	}
}

func TestHexDigest12_NoColon(t *testing.T) {
	got := HexDigest12("sha256:abc123def4567890123456789012345678901234567890123456789012345678")
	if len(got) != 12 {
		t.Errorf("HexDigest12 length = %d, want 12", len(got))
	}
	if got[0] == ':' || got[0] == 's' {
		t.Error("HexDigest12 must strip sha256: prefix")
	}
}
