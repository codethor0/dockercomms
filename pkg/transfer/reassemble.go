// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"regexp"

	"github.com/klauspost/compress/zstd"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var chunkEntryRe = regexp.MustCompile(`^chunk_\d+\.bin$`)

// Reassemble reads chunk layers in order and writes the reassembled content to w.
// Layers must be sorted by dockercomms.chunk.index.
func Reassemble(layers []ocispec.Descriptor, fetch func(ocispec.Descriptor) (io.ReadCloser, error), w io.Writer) (int64, error) {
	byIndex := make(map[int]ocispec.Descriptor)
	for _, d := range layers {
		idxStr, ok := d.Annotations["dockercomms.chunk.index"]
		if !ok {
			return 0, fmt.Errorf("layer missing dockercomms.chunk.index: %s", d.Digest)
		}
		var idx int
		if _, err := fmt.Sscanf(idxStr, "%d", &idx); err != nil {
			return 0, fmt.Errorf("invalid chunk index %q: %w", idxStr, err)
		}
		byIndex[idx] = d
	}
	var total int64
	for i := 0; i < len(byIndex); i++ {
		d, ok := byIndex[i]
		if !ok {
			return total, fmt.Errorf("missing chunk index %d", i)
		}
		rc, err := fetch(d)
		if err != nil {
			return total, err
		}
		n, err := extractChunk(rc, d.MediaType, w)
		if cerr := rc.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close chunk blob: %w", cerr)
		}
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

func extractChunk(r io.Reader, mediaType string, w io.Writer) (n int64, err error) {
	var decompressed io.Reader
	if mediaType == "application/vnd.dockercomms.chunk.v1.tar+zstd" {
		zr, e := zstd.NewReader(r)
		if e != nil {
			return 0, e
		}
		defer zr.Close()
		decompressed = zr
	} else {
		gr, e := gzip.NewReader(r)
		if e != nil {
			return 0, e
		}
		defer func() {
			if cerr := gr.Close(); cerr != nil && err == nil {
				err = fmt.Errorf("close gzip: %w", cerr)
			}
		}()
		decompressed = gr
	}
	tr := tar.NewReader(decompressed)
	hdr, err := tr.Next()
	if err == io.EOF {
		return 0, fmt.Errorf("empty tar chunk")
	}
	if err != nil {
		return 0, err
	}
	if !chunkEntryRe.MatchString(hdr.Name) {
		return 0, fmt.Errorf("unexpected tar entry %q", hdr.Name)
	}
	n, err = io.Copy(w, tr) // #nosec G110 -- chunk size bounded by manifest annotations and MaxChunks
	if err != nil {
		return n, err
	}
	_, err = tr.Next()
	if err == nil {
		return n, fmt.Errorf("extra tar entry in chunk")
	}
	if err != io.EOF {
		return n, err
	}
	return n, nil
}
