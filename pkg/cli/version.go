// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/codethor0/dockercomms/internal/version"
	"github.com/codethor0/dockercomms/pkg/config"
	"github.com/spf13/cobra"
)

func newVersionCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.JSON {
				fmt.Printf(`{"version":"%s","commit":"%s","date":"%s"}`+"\n",
					version.Version, version.Commit, version.Date)
			} else {
				fmt.Printf("dockercomms version %s\ncommit: %s\ndate: %s\n",
					version.Version, version.Commit, version.Date)
			}
			return nil
		},
	}
	return cmd
}
