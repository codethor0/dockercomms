// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"strings"
	"testing"
)

func TestVersionFormat(t *testing.T) {
	if Version == "" {
		t.Error("Version must not be empty")
	}
	if Commit == "" {
		t.Error("Commit must not be empty")
	}
	if Date == "" {
		t.Error("Date must not be empty")
	}
}

func TestVersionString(t *testing.T) {
	s := Version + " " + Commit + " " + Date
	if strings.Contains(s, "undefined") {
		t.Error("version string must not contain undefined")
	}
}
