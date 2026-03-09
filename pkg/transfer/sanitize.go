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
	// Strip backslash (Windows separator); filepath.Base on Unix does not split on backslash
	base = strings.ReplaceAll(base, "\\", "")
	// filepath.Base("/") and filepath.Base("//") return "/" on Unix
	if base == "" || base == "." || base == ".." || base == "/" {
		return "file"
	}
	return base
}
