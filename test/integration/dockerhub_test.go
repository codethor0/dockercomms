// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/codethor0/dockercomms/pkg/oci"
)

func TestDockerHubTagListing_Smoke(t *testing.T) {
	repo := os.Getenv("DOCKERCOMMS_IT_DH_REPO")
	if repo == "" {
		t.Skip("DOCKERCOMMS_IT_DH_REPO required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()

	client, err := oci.NewClient(repo)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	var count int
	err = client.Tags(ctx, "", func(tags []string) error {
		count += len(tags)
		return nil
	})
	if err != nil {
		t.Fatalf("Tags: %v", err)
	}
	t.Logf("tag listing: %d tags", count)
}
