// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/codethor0/dockercomms/pkg/config"
	"github.com/spf13/cobra"
)

// AddCommands attaches all subcommands to the root.
func AddCommands(root *cobra.Command, cfg *config.Config) {
	root.AddCommand(newSendCmd(cfg))
	root.AddCommand(newRecvCmd(cfg))
	root.AddCommand(newVerifyCmd(cfg))
	root.AddCommand(newAckCmd(cfg))
	root.AddCommand(newVersionCmd(cfg))
}
