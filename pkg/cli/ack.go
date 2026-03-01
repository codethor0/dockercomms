// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"

	"github.com/codethor0/dockercomms/pkg/config"
	"github.com/codethor0/dockercomms/pkg/transfer"
	"github.com/spf13/cobra"
)

func newAckCmd(cfg *config.Config) *cobra.Command {
	var repo, forDigest, status, reason string
	var verified bool
	cmd := &cobra.Command{
		Use:   "ack",
		Short: "Write receipt artifact",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := transfer.AckOptions{
				Repo:     repo,
				For:      forDigest,
				Status:   status,
				Verified: verified,
				Reason:   reason,
			}
			if err := transfer.Ack(context.Background(), opts); err != nil {
				return &ExitError{Code: ExitGenericFailure, Err: fmt.Errorf("ack failed: %w", err)}
			}
			if cfg.JSON {
				fmt.Printf(`{"status":"%s"}`+"\n", status)
			} else {
				fmt.Printf("receipt written: %s\n", status)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "Registry/repo (required)")
	cmd.Flags().StringVar(&forDigest, "for", "", "Message digest sha256:... (required)")
	cmd.Flags().StringVar(&status, "status", "accepted", "Receipt status: accepted or rejected")
	cmd.Flags().BoolVar(&verified, "verified", false, "Whether verification succeeded")
	cmd.Flags().StringVar(&reason, "reason", "", "Optional reason")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("for")
	return cmd
}
