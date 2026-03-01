// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
)

// Config holds DockerComms configuration.
type Config struct {
	ConfigPath string
	Verbose    bool
	JSON       bool
}

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() string {
	if p := os.Getenv("DOCKERCOMMS_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "dockercomms", "config.yaml")
}
