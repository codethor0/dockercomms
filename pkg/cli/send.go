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

func newSendCmd(cfg *config.Config) *cobra.Command {
	var (
		repo, recipient, session string
		chunkBytes               int64
		parallel                 int
		compress                 string
		ttlSeconds               int
		sign                     bool
		cosignPath, identity     string
	)
	cmd := &cobra.Command{
		Use:   "send <file_path>",
		Short: "Push a file as an OCI artifact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := transfer.SendOptions{
				Repo:       repo,
				Recipient:  recipient,
				Session:    session,
				ChunkBytes: chunkBytes,
				Compress:   compress,
				TTLSeconds: ttlSeconds,
				Sign:       sign,
				CosignPath: cosignPath,
				Identity:   identity,
			}
			result, err := transfer.Send(context.Background(), args[0], opts)
			if err != nil {
				return &ExitError{Code: classifyRecvError(err), Err: fmt.Errorf("send failed: %w", err)}
			}
			if cfg.JSON {
				fmt.Printf(`{"digest":"%s","tag":"%s","bundle":"%s"}`+"\n", result.Digest, result.Tag, result.Bundle)
			} else {
				fmt.Printf("digest: %s\ntag: %s\n", result.Digest, result.Tag)
				if result.Bundle != "" {
					fmt.Printf("bundle: %s\n", result.Bundle)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "Registry/repo (required)")
	cmd.Flags().StringVar(&recipient, "recipient", "", "Recipient identity (required)")
	cmd.Flags().StringVar(&session, "session", "", "Session UUID (optional)")
	cmd.Flags().Int64Var(&chunkBytes, "chunk-bytes", 104857600, "Chunk target size in bytes")
	cmd.Flags().IntVar(&parallel, "parallel", 4, "Parallel uploads")
	cmd.Flags().StringVar(&compress, "compress", "gzip", "Compression: gzip or zstd")
	cmd.Flags().IntVar(&ttlSeconds, "ttl-seconds", 86400*7, "TTL in seconds")
	cmd.Flags().BoolVar(&sign, "sign", true, "Sign artifact")
	cmd.Flags().StringVar(&cosignPath, "cosign", "cosign", "Path to cosign binary")
	cmd.Flags().StringVar(&identity, "identity", "", "Identity for keyless signing (also used as sender if --sender not set)")
	_ = cmd.MarkFlagRequired("repo") // valid flag names
	_ = cmd.MarkFlagRequired("recipient")
	return cmd
}
