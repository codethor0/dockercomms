// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"
	"testing"
)

func TestClassifyRecvError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"verification failed", fmt.Errorf("verification failed: bad signature"), ExitVerificationFailed},
		{"verification wrapped", fmt.Errorf("recv failed: %w", fmt.Errorf("verification failed: x")), ExitVerificationFailed},
		{"denied", fmt.Errorf("GET ... 403: denied"), ExitAuthError},
		{"unauthorized", fmt.Errorf("401 unauthorized"), ExitAuthError},
		{"auth required", fmt.Errorf("authentication required"), ExitAuthError},
		{"bundle not found", fmt.Errorf("fetch bundle: bundle not found"), ExitNotFound},
		{"not found", fmt.Errorf("list tags: not found"), ExitNotFound},
		{"generic", fmt.Errorf("some other error"), ExitGenericFailure},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyRecvError(tt.err)
			if got != tt.want {
				t.Errorf("classifyRecvError() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestClassifyRecvError_Unwrap(t *testing.T) {
	base := errors.New("bundle not found")
	wrapped := fmt.Errorf("fetch bundle: %w", base)
	got := classifyRecvError(wrapped)
	if got != ExitNotFound {
		t.Errorf("classifyRecvError(unwrapped) = %d, want %d (ExitNotFound)", got, ExitNotFound)
	}
}

func TestClassifyRecvError_VerifyFetch(t *testing.T) {
	// 403 from registry should map to auth error (used by verify and recv)
	err := fmt.Errorf("verify failed: list tags: GET ... 403: denied")
	got := classifyRecvError(err)
	if got != ExitAuthError {
		t.Errorf("classifyRecvError(403) = %d, want %d (ExitAuthError)", got, ExitAuthError)
	}
}

func TestClassifyRecvError_Send(t *testing.T) {
	// Send uses same classification; 403 should map to auth error
	err := fmt.Errorf("send failed: push blob: 403 denied")
	got := classifyRecvError(err)
	if got != ExitAuthError {
		t.Errorf("classifyRecvError(send 403) = %d, want %d (ExitAuthError)", got, ExitAuthError)
	}
}
