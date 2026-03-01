// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"path/filepath"
	"strings"
)

// SanitizeFilename returns a safe basename for use in annotations.
// Defends against path traversal: ensures result is basename only, no slashes or parent refs.
func SanitizeFilename(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSpace(base)
	if base == "" || base == "." || base == ".." {
		return "file"
	}
	return base
}
