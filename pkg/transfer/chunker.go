// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	DefaultChunkBytes = 104857600 // 100 MiB
	MaxChunks         = 10000
	MaxTotalBytes     = 10 * 1024 * 1024 * 1024 * 1024 // 10 TiB
)

// Compression is gzip or zstd.
type Compression string

const (
	CompressionGzip Compression = "gzip"
	CompressionZstd Compression = "zstd"
)

// ChunkDescriptor holds a chunk layer descriptor with index.
type ChunkDescriptor struct {
	Descriptor ocispec.Descriptor
	Index      int
}

// Chunker streams a file into compressed tar chunks.
type Chunker struct {
	ChunkBytes int64
	Compress   Compression
}

// NewChunker creates a chunker with default or custom settings.
func NewChunker(chunkBytes int64, comp Compression) *Chunker {
	if chunkBytes <= 0 {
		chunkBytes = DefaultChunkBytes
	}
	if comp == "" {
		comp = CompressionGzip
	}
	return &Chunker{ChunkBytes: chunkBytes, Compress: comp}
}

// MediaType returns the layer media type for the compression.
func (c *Chunker) MediaType() string {
	if c.Compress == CompressionZstd {
		return "application/vnd.dockercomms.chunk.v1.tar+zstd"
	}
	return "application/vnd.dockercomms.chunk.v1.tar+gzip"
}

// Chunk reads from r and produces chunk descriptors and content.
// Each chunk is a tar with one entry: chunk_<index>.bin.
func (c *Chunker) Chunk(r io.Reader, totalSize int64) ([]ChunkDescriptor, []io.Reader, error) {
	if totalSize > MaxTotalBytes {
		return nil, nil, fmt.Errorf("total size %d exceeds max %d", totalSize, MaxTotalBytes)
	}
	var descriptors []ChunkDescriptor
	var readers []io.Reader
	var idx int
	remaining := totalSize
	for remaining > 0 {
		if idx >= MaxChunks {
			return nil, nil, fmt.Errorf("exceeded max chunks %d", MaxChunks)
		}
		toRead := c.ChunkBytes
		if remaining < toRead {
			toRead = remaining
		}
		chunkData := make([]byte, toRead)
		n, err := io.ReadFull(r, chunkData)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}
		if n == 0 {
			break
		}
		chunkData = chunkData[:n]
		remaining -= int64(n)

		tarBuf := &bytes.Buffer{}
		tw := tar.NewWriter(tarBuf)
		hdr := &tar.Header{
			Name: fmt.Sprintf("chunk_%d.bin", idx),
			Mode: 0644,
			Size: int64(len(chunkData)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, nil, err
		}
		if _, err := tw.Write(chunkData); err != nil {
			return nil, nil, err
		}
		if err := tw.Close(); err != nil {
			return nil, nil, err
		}
		tarBytes := tarBuf.Bytes()

		var compressedBuf bytes.Buffer
		if c.Compress == CompressionZstd {
			zw, err := zstd.NewWriter(&compressedBuf)
			if err != nil {
				return nil, nil, err
			}
			if _, err := zw.Write(tarBytes); err != nil {
				if cerr := zw.Close(); cerr != nil {
					return nil, nil, fmt.Errorf("write chunk: %w (close: %v)", err, cerr)
				}
				return nil, nil, err
			}
			if err := zw.Close(); err != nil {
				return nil, nil, err
			}
		} else {
			gw := gzip.NewWriter(&compressedBuf)
			if _, err := gw.Write(tarBytes); err != nil {
				if cerr := gw.Close(); cerr != nil {
					return nil, nil, fmt.Errorf("write chunk: %w (close: %v)", err, cerr)
				}
				return nil, nil, err
			}
			if err := gw.Close(); err != nil {
				return nil, nil, err
			}
		}
		compressed := compressedBuf.Bytes()
		d := digest.FromBytes(compressed)
		desc := ocispec.Descriptor{
			MediaType: c.MediaType(),
			Digest:    d,
			Size:      int64(len(compressed)),
			Annotations: map[string]string{
				"dockercomms.chunk.index": fmt.Sprintf("%d", idx),
			},
		}
		descriptors = append(descriptors, ChunkDescriptor{Descriptor: desc, Index: idx})
		readers = append(readers, bytes.NewReader(compressed))
		idx++
	}
	return descriptors, readers, nil
}

// ChunkFile opens a file and chunks it.
func (c *Chunker) ChunkFile(path string) (descs []ChunkDescriptor, readers []io.Reader, size int64, err error) {
	f, err := os.Open(path) // #nosec G304 -- path from CLI, validated by caller before ChunkFile
	if err != nil {
		return nil, nil, 0, err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close file: %w", cerr)
		}
	}()
	info, err := f.Stat()
	if err != nil {
		return nil, nil, 0, err
	}
	if info.IsDir() {
		return nil, nil, 0, fmt.Errorf("cannot chunk directory")
	}
	size = info.Size()
	descs, readers, err = c.Chunk(f, size)
	return descs, readers, size, err
}
