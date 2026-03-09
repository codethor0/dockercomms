// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"file.txt", "file.txt"},
		{"/path/to/file.txt", "file.txt"},
		{"..", "file"},
		{".", "file"},
		{"", "file"},
		{"  ", "file"},
		{"../../../etc/passwd", "passwd"},
		{"/", "file"},
		{"//", "file"},
		{"..\\evil.txt", "..evil.txt"},
		{"..\\\\..\\\\evil.txt", "....evil.txt"},
		{"/absolute/path/like.txt", "like.txt"},
		{"  spaces  .txt  ", "spaces  .txt"},
	}
	for _, tt := range tests {
		got := SanitizeFilename(tt.in)
		if got != tt.want {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
