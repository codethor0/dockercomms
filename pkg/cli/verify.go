// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"

	"github.com/codethor0/dockercomms/pkg/config"
	"github.com/codethor0/dockercomms/pkg/oci"
	"github.com/codethor0/dockercomms/pkg/transfer"
	"github.com/spf13/cobra"
)

func newVerifyCmd(cfg *config.Config) *cobra.Command {
	var repo, digest, bundleMode, policy, trustedRoot string
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify artifact digest using bundle",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := oci.NewClient(repo)
			if err != nil {
				return &ExitError{Code: ExitGenericFailure, Err: fmt.Errorf("verify failed: %w", err)}
			}
			var bundleBytes []byte
			switch bundleMode {
			case "auto":
				bundleBytes, err = transfer.FetchBundleForVerify(context.Background(), client, repo, digest)
			case "referrers":
				bundleBytes, err = transfer.FetchBundleReferrers(context.Background(), client, repo, digest)
			case "tag":
				bundleBytes, err = transfer.FetchBundleTag(context.Background(), client, digest)
			default:
				return &ExitError{Code: ExitProtocolError, Err: fmt.Errorf("unknown bundle mode: %s", bundleMode)}
			}
			if err != nil {
				return &ExitError{Code: classifyRecvError(err), Err: fmt.Errorf("verify failed: %w", err)}
			}
			if err := transfer.VerifyBundleInProcess(bundleBytes, digest, trustedRoot); err != nil {
				return &ExitError{Code: ExitVerificationFailed, Err: fmt.Errorf("verification failed: %w", err)}
			}
			if cfg.JSON {
				fmt.Printf(`{"verified":true,"digest":"%s"}`+"\n", digest)
			} else {
				fmt.Printf("verified: %s\n", digest)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "Registry/repo (required)")
	cmd.Flags().StringVar(&digest, "digest", "", "Artifact digest sha256:... (required)")
	cmd.Flags().StringVar(&bundleMode, "bundle", "auto", "Bundle source: auto, referrers, or tag")
	cmd.Flags().StringVar(&policy, "policy", "", "Verification policy path")
	cmd.Flags().StringVar(&trustedRoot, "trusted-root", "", "Trusted root path")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("digest")
	return cmd
}
