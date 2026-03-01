// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"testing"
	"time"
)

func TestSinceFilterPass(t *testing.T) {
	since := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		createdAt string
		want      bool
	}{
		{"empty", "", false},
		{"invalid format", "not-a-date", true},
		{"within 5 min before", "2025-01-15T11:56:00Z", true},
		{"exactly 5 min before", "2025-01-15T11:55:00Z", true},
		{"over 5 min before", "2025-01-15T11:54:59Z", false},
		{"after since", "2025-01-15T12:05:00Z", true},
		{"same time", "2025-01-15T12:00:00Z", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SinceFilterPass(tt.createdAt, since)
			if got != tt.want {
				t.Errorf("SinceFilterPass(%q, since) = %v, want %v", tt.createdAt, got, tt.want)
			}
		})
	}
}
