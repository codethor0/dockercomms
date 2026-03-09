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

func newRecvCmd(cfg *config.Config) *cobra.Command {
	var (
		repo, me, out, since string
		max                  int
		verify, writeReceipt bool
		policy, trustedRoot  string
	)
	cmd := &cobra.Command{
		Use:   "recv",
		Short: "Receive files from inbox",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := transfer.RecvOptions{
				Repo:         repo,
				Me:           me,
				Out:          out,
				Since:        since,
				Max:          max,
				Verify:       verify,
				WriteReceipt: writeReceipt,
				TrustedRoot:  trustedRoot,
			}
			count, err := transfer.Recv(context.Background(), opts)
			if err != nil {
				return &ExitError{Code: classifyRecvError(err), Err: fmt.Errorf("recv failed: %w", err)}
			}
			if cfg.JSON {
				fmt.Printf(`{"received":%d}`+"\n", count)
			} else {
				fmt.Printf("received %d message(s)\n", count)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "Registry/repo (required)")
	cmd.Flags().StringVar(&me, "me", "", "Recipient identity (required)")
	cmd.Flags().StringVar(&out, "out", "", "Output directory (required)")
	cmd.Flags().StringVar(&since, "since", "", "Filter by created_at >= RFC3339")
	cmd.Flags().IntVar(&max, "max", 100, "Max messages to process")
	cmd.Flags().BoolVar(&verify, "verify", true, "Verify before materialize")
	cmd.Flags().BoolVar(&writeReceipt, "write-receipt", true, "Write receipt artifact")
	cmd.Flags().StringVar(&policy, "policy", "", "Verification policy path")
	cmd.Flags().StringVar(&trustedRoot, "trusted-root", "", "Trusted root path")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("me")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}
