// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestChunkerReconstruct(t *testing.T) {
	original := []byte("hello world chunked content for testing reconstruction")
	c := NewChunker(20, CompressionGzip)
	descs, readers, err := c.Chunk(bytes.NewReader(original), int64(len(original)))
	if err != nil {
		t.Fatal(err)
	}
	if len(descs) != 3 {
		t.Fatalf("got %d chunks, want 3", len(descs))
	}
	for i, d := range descs {
		if d.Index != i {
			t.Errorf("chunk %d: index = %d", i, d.Index)
		}
		idxStr := d.Descriptor.Annotations["dockercomms.chunk.index"]
		if idxStr != fmt.Sprintf("%d", i) {
			t.Errorf("chunk %d: annotation = %q, want %q", i, idxStr, fmt.Sprintf("%d", i))
		}
	}
	// Build layer list and fetch func for Reassemble
	layers := make([]ocispec.Descriptor, len(descs))
	compressed := make([][]byte, len(descs))
	for i, d := range descs {
		layers[i] = d.Descriptor
		var err error
		compressed[i], err = io.ReadAll(readers[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) {
		for i, l := range layers {
			if l.Digest == d.Digest {
				return io.NopCloser(bytes.NewReader(compressed[i])), nil
			}
		}
		return nil, fmt.Errorf("not found: %s", d.Digest)
	}
	var buf bytes.Buffer
	n, err := Reassemble(layers, fetch, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != int64(len(original)) {
		t.Errorf("reassembled %d bytes, want %d", n, len(original))
	}
	if !bytes.Equal(buf.Bytes(), original) {
		t.Errorf("reconstructed = %q, want %q", buf.Bytes(), original)
	}
}

func TestChunkerZstd(t *testing.T) {
	original := []byte("zstd test data")
	c := NewChunker(10, CompressionZstd)
	descs, readers, err := c.Chunk(bytes.NewReader(original), int64(len(original)))
	if err != nil {
		t.Fatal(err)
	}
	if len(descs) != 2 {
		t.Fatalf("got %d chunks, want 2", len(descs))
	}
	if descs[0].Descriptor.MediaType != "application/vnd.dockercomms.chunk.v1.tar+zstd" {
		t.Errorf("media type = %q", descs[0].Descriptor.MediaType)
	}
	layers := make([]ocispec.Descriptor, len(descs))
	compressed := make([][]byte, len(descs))
	for i, d := range descs {
		layers[i] = d.Descriptor
		var err error
		compressed[i], err = io.ReadAll(readers[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) {
		for i, l := range layers {
			if l.Digest == d.Digest {
				return io.NopCloser(bytes.NewReader(compressed[i])), nil
			}
		}
		return nil, fmt.Errorf("not found: %s", d.Digest)
	}
	var buf bytes.Buffer
	n, err := Reassemble(layers, fetch, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != int64(len(original)) {
		t.Errorf("reassembled %d bytes, want %d", n, len(original))
	}
	if !bytes.Equal(buf.Bytes(), original) {
		t.Errorf("reconstructed = %q, want %q", buf.Bytes(), original)
	}
}

func TestChunkerChunkFile(t *testing.T) {
	tmp := t.TempDir()
	fpath := tmp + "/data.bin"
	content := []byte("file content for ChunkFile test")
	if err := os.WriteFile(fpath, content, 0600); err != nil {
		t.Fatal(err)
	}
	c := NewChunker(10, CompressionGzip)
	descs, readers, size, err := c.ChunkFile(fpath)
	if err != nil {
		t.Fatal(err)
	}
	if size != int64(len(content)) {
		t.Errorf("size = %d, want %d", size, len(content))
	}
	if len(descs) == 0 {
		t.Fatal("expected at least one chunk")
	}
	_ = readers
}

func TestChunkerChunkFile_NotExist(t *testing.T) {
	c := NewChunker(10, CompressionGzip)
	_, _, _, err := c.ChunkFile("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestChunkerChunkFile_Dir(t *testing.T) {
	tmp := t.TempDir()
	c := NewChunker(10, CompressionGzip)
	_, _, _, err := c.ChunkFile(tmp)
	if err == nil {
		t.Fatal("expected error for directory")
	}
}

func TestNewChunker_Defaults(t *testing.T) {
	c := NewChunker(0, "")
	if c.ChunkBytes != DefaultChunkBytes {
		t.Errorf("ChunkBytes = %d, want %d", c.ChunkBytes, DefaultChunkBytes)
	}
	if c.Compress != CompressionGzip {
		t.Errorf("Compress = %q, want gzip", c.Compress)
	}
}

func TestChunker_ExceedsMaxTotalBytes(t *testing.T) {
	c := NewChunker(1024, CompressionGzip)
	_, _, err := c.Chunk(bytes.NewReader([]byte("x")), MaxTotalBytes+1)
	if err == nil {
		t.Fatal("expected error for exceeding max total bytes")
	}
}

func TestChunker_ExceedsMaxChunks(t *testing.T) {
	c := NewChunker(1, CompressionGzip)
	data := make([]byte, MaxChunks+1)
	for i := range data {
		data[i] = 'x'
	}
	_, _, err := c.Chunk(bytes.NewReader(data), int64(len(data)))
	if err == nil {
		t.Fatal("expected error for exceeding max chunks")
	}
}
