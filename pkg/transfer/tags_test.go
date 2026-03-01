// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"regexp"
	"testing"
)

// Tag format: [A-Za-z0-9_.-], max 128, must start with [A-Za-z0-9_].
var tagSafeRe = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)

func TestRecipientTag(t *testing.T) {
	tests := []struct {
		recipient string
		want      string
	}{
		{"alice@example.com", "vu6pg6fs1o9bu394h4n4b63u4i"},
		{"bob", "g6r3fm7sqb3dkoqpsqb324t12s"},
		{"", "seoc8gkovge196nruj49irtp4g"},
	}
	for _, tt := range tests {
		got := RecipientTag(tt.recipient)
		if got != tt.want {
			t.Errorf("RecipientTag(%q) = %q, want %q", tt.recipient, got, tt.want)
		}
		if len(got) != 26 {
			t.Errorf("RecipientTag(%q) length = %d, want 26", tt.recipient, len(got))
		}
		if !tagSafeRe.MatchString(got) {
			t.Errorf("RecipientTag(%q) = %q, not tag-safe", tt.recipient, got)
		}
	}
}

func TestRecipientTagDeterministic(t *testing.T) {
	r := "test@example.com"
	a, b := RecipientTag(r), RecipientTag(r)
	if a != b {
		t.Errorf("RecipientTag must be deterministic: %q != %q", a, b)
	}
}

func TestRecipientTagLength(t *testing.T) {
	got := RecipientTag("x")
	if len(got) != 26 {
		t.Errorf("len(RecipientTag) = %d, want 26", len(got))
	}
}

func TestHexDigest12(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"sha256:abc123def456789", "abc123def456"},
		{"sha256:abc", "abc"},
		{"abc123def456", "abc123def456"},
	}
	for _, tt := range tests {
		got := HexDigest12(tt.in)
		if got != tt.want {
			t.Errorf("HexDigest12(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
