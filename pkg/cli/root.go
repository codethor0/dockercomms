// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"

	"github.com/codethor0/dockercomms/pkg/config"
	"github.com/spf13/cobra"
)

// Exit codes per CLI contract.
const (
	ExitSuccess           = 0
	ExitGenericFailure    = 1
	ExitVerificationFailed = 2
	ExitAuthError         = 3
	ExitProtocolError     = 4
	ExitNotFound          = 5
)

// ExitError carries an exit code for main to use. Only main() may call os.Exit.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit %d", e.Code)
}

func (e *ExitError) Unwrap() error { return e.Err }

// ExitCode returns the exit code from err if it is an ExitError, else ExitGenericFailure.
func ExitCode(err error) int {
	var e *ExitError
	if errors.As(err, &e) {
		return e.Code
	}
	return ExitGenericFailure
}

// NewRootCmd creates the root cobra command.
func NewRootCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "dockercomms",
		Short:         "OCI-native secure file transport CLI",
		Long:          "DockerComms: push and pull files as OCI artifacts with verification.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().StringVar(&cfg.ConfigPath, "config", config.DefaultConfigPath(), "Config file path")
	cmd.PersistentFlags().BoolVar(&cfg.JSON, "json", false, "Output JSON")
	cmd.PersistentFlags().BoolVar(&cfg.Verbose, "verbose", false, "Verbose output")
	return cmd
}
