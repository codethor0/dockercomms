// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestChunkReassembleProperty_RandomSizes(t *testing.T) {
	sizes := []int{0, 1, 100, 1024, 10240, 100000}
	chunkSizes := []int64{100, 1024, 65536, 1048576}
	for _, totalSize := range sizes {
		for _, chunkSize := range chunkSizes {
			if totalSize == 0 && chunkSize > 0 {
				continue
			}
			t.Run(fmt.Sprintf("size%d_chunk%d", totalSize, chunkSize), func(t *testing.T) {
				data := make([]byte, totalSize)
				if totalSize > 0 {
					if _, err := rand.Read(data); err != nil {
						t.Fatal(err)
					}
				}
				c := NewChunker(chunkSize, CompressionGzip)
				descs, readers, err := c.Chunk(bytes.NewReader(data), int64(totalSize))
				if err != nil {
					t.Fatal(err)
				}
				layers := make([]ocispec.Descriptor, len(descs))
				blobs := make([][]byte, len(descs))
				for i, d := range descs {
					layers[i] = d.Descriptor
					blobs[i], err = io.ReadAll(readers[i])
					if err != nil {
						t.Fatal(err)
					}
				}
				fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) {
					for i, l := range layers {
						if l.Digest == d.Digest {
							return io.NopCloser(bytes.NewReader(blobs[i])), nil
						}
					}
					return nil, fmt.Errorf("blob not found")
				}
				var out bytes.Buffer
				n, err := Reassemble(layers, fetch, &out)
				if err != nil {
					t.Fatal(err)
				}
				if n != int64(totalSize) {
					t.Errorf("reassembled %d bytes, want %d", n, totalSize)
				}
				if !bytes.Equal(out.Bytes(), data) {
					t.Error("reassembled content != original")
				}
			})
		}
	}
}

func TestChunkExactlyOneTarEntry(t *testing.T) {
	data := []byte("x")
	c := NewChunker(10, CompressionGzip)
	descs, readers, err := c.Chunk(bytes.NewReader(data), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(descs) != 1 {
		t.Fatalf("want 1 chunk, got %d", len(descs))
	}
	idxStr := descs[0].Descriptor.Annotations["dockercomms.chunk.index"]
	if idxStr != "0" {
		t.Errorf("chunk index = %q, want 0", idxStr)
	}
	blob, err := io.ReadAll(readers[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(blob) == 0 {
		t.Error("chunk blob empty")
	}
}
