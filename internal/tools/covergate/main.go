// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

// covergate checks per-package coverage against thresholds.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var thresholds = map[string]float64{
	"github.com/codethor0/dockercomms/pkg/crypto":   68,
	"github.com/codethor0/dockercomms/pkg/transfer": 37,
	"github.com/codethor0/dockercomms/pkg/oci":      54,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: covergate <coverage.out>\n")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1]) // #nosec G703 -- coverage path from make coverage-gate, not user input
	if err != nil {
		fmt.Fprintf(os.Stderr, "covergate: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "covergate: close: %v\n", cerr)
		}
	}()

	// Parse coverage profile: filename:start.end,start.end numStmts count
	// Extract package from filename (path before last /)
	type pkgStats struct {
		total   int
		covered int
	}
	stats := make(map[string]*pkgStats)

	scanner := bufio.NewScanner(f)
	first := true
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "mode:") {
			first = false
			continue
		}
		if first {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		pathRange := parts[0]
		numStmts, err1 := strconv.Atoi(parts[1])
		count, err2 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil {
			continue
		}
		colon := strings.Index(pathRange, ":")
		if colon < 0 {
			continue
		}
		path := pathRange[:colon]
		slash := strings.LastIndex(path, "/")
		if slash < 0 {
			continue
		}
		pkg := path[:slash]
		if stats[pkg] == nil {
			stats[pkg] = &pkgStats{}
		}
		stats[pkg].total += numStmts
		if count > 0 {
			stats[pkg].covered += numStmts
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "covergate: %v\n", err)
		os.Exit(1)
	}

	failed := false
	for pkg, thresh := range thresholds {
		s := stats[pkg]
		if s == nil || s.total == 0 {
			fmt.Printf("%s: no coverage data (threshold %.0f%%)\n", pkg, thresh)
			failed = true
			continue
		}
		pct := 100 * float64(s.covered) / float64(s.total)
		if pct < thresh {
			fmt.Printf("%s: %.1f%% (need %.0f%%)\n", pkg, pct, thresh)
			failed = true
		} else {
			fmt.Printf("%s: %.1f%% OK\n", pkg, pct)
		}
	}
	if failed {
		os.Exit(1)
	}
}
