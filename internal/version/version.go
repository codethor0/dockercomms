// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

// Package version provides build-time version information for DockerComms.
package version

// These are set via -ldflags at build time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
