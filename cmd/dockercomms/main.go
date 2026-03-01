// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/codethor0/dockercomms/pkg/cli"
	"github.com/codethor0/dockercomms/pkg/config"
)

func main() {
	cfg := &config.Config{}
	root := cli.NewRootCmd(cfg)
	cli.AddCommands(root, cfg)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(cli.ExitCode(err))
	}
}
