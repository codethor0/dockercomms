// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/codethor0/dockercomms/pkg/oci"
)

func TestReassemble_MissingChunkIndex(t *testing.T) {
	layers := []ocispec.Descriptor{
		{Digest: "sha256:a", Size: 1, Annotations: map[string]string{}},
	}
	fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) { return nil, nil }
	var buf bytes.Buffer
	_, err := Reassemble(layers, fetch, &buf)
	if err == nil {
		t.Fatal("expected error for missing chunk index")
	}
	if !strings.Contains(err.Error(), "dockercomms.chunk.index") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReassemble_InvalidChunkIndex(t *testing.T) {
	layers := []ocispec.Descriptor{
		{Digest: "sha256:a", Size: 1, Annotations: map[string]string{"dockercomms.chunk.index": "x"}},
	}
	fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) { return nil, nil }
	var buf bytes.Buffer
	_, err := Reassemble(layers, fetch, &buf)
	if err == nil {
		t.Fatal("expected error for invalid index")
	}
	if !strings.Contains(err.Error(), "invalid chunk index") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReassemble_MissingChunk(t *testing.T) {
	validChunk := makeValidGzipChunk(t, []byte("x"))
	layers := []ocispec.Descriptor{
		{Digest: "sha256:a", Size: int64(len(validChunk)), MediaType: oci.MediaTypeChunkGzip, Annotations: map[string]string{"dockercomms.chunk.index": "0"}},
		{Digest: "sha256:c", Size: 1, Annotations: map[string]string{"dockercomms.chunk.index": "2"}},
	}
	fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) {
		if d.Annotations["dockercomms.chunk.index"] == "0" {
			return io.NopCloser(bytes.NewReader(validChunk)), nil
		}
		return nil, nil
	}
	var buf bytes.Buffer
	_, err := Reassemble(layers, fetch, &buf)
	if err == nil {
		t.Fatal("expected error for missing chunk")
	}
	if !strings.Contains(err.Error(), "missing chunk index 1") {
		t.Errorf("unexpected error: %v", err)
	}
}

func makeValidGzipChunk(t *testing.T, data []byte) []byte {
	t.Helper()
	var gzBuf bytes.Buffer
	gz := gzip.NewWriter(&gzBuf)
	tr := tar.NewWriter(gz)
	_ = tr.WriteHeader(&tar.Header{Name: "chunk_0.bin", Size: int64(len(data))})
	_, _ = tr.Write(data)
	_ = tr.Close()
	_ = gz.Close()
	return gzBuf.Bytes()
}

func TestReassemble_FetchError(t *testing.T) {
	layers := []ocispec.Descriptor{
		{Digest: "sha256:a", Size: 1, Annotations: map[string]string{"dockercomms.chunk.index": "0"}},
	}
	fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) {
		return nil, fmt.Errorf("fetch failed")
	}
	var buf bytes.Buffer
	_, err := Reassemble(layers, fetch, &buf)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReassemble_EmptyTarChunk(t *testing.T) {
	var gzBuf bytes.Buffer
	gz := gzip.NewWriter(&gzBuf)
	tr := tar.NewWriter(gz)
	_ = tr.Close()
	_ = gz.Close()

	layers := []ocispec.Descriptor{
		{Digest: "sha256:a", Size: int64(gzBuf.Len()), MediaType: oci.MediaTypeChunkGzip, Annotations: map[string]string{"dockercomms.chunk.index": "0"}},
	}
	fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(gzBuf.Bytes())), nil
	}
	var buf bytes.Buffer
	_, err := Reassemble(layers, fetch, &buf)
	if err == nil {
		t.Fatal("expected error for empty tar")
	}
	if !strings.Contains(err.Error(), "empty tar") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReassemble_UnexpectedTarEntry(t *testing.T) {
	var gzBuf bytes.Buffer
	gz := gzip.NewWriter(&gzBuf)
	tr := tar.NewWriter(gz)
	_ = tr.WriteHeader(&tar.Header{Name: "badname.bin", Size: 0})
	_ = tr.Close()
	_ = gz.Close()

	layers := []ocispec.Descriptor{
		{Digest: "sha256:a", Size: int64(gzBuf.Len()), MediaType: oci.MediaTypeChunkGzip, Annotations: map[string]string{"dockercomms.chunk.index": "0"}},
	}
	fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(gzBuf.Bytes())), nil
	}
	var buf bytes.Buffer
	_, err := Reassemble(layers, fetch, &buf)
	if err == nil {
		t.Fatal("expected error for unexpected entry")
	}
	if !strings.Contains(err.Error(), "unexpected tar entry") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReassemble_ExtraTarEntry(t *testing.T) {
	var gzBuf bytes.Buffer
	gz := gzip.NewWriter(&gzBuf)
	tr := tar.NewWriter(gz)
	_ = tr.WriteHeader(&tar.Header{Name: "chunk_0.bin", Size: 3})
	_, _ = tr.Write([]byte("abc"))
	_ = tr.WriteHeader(&tar.Header{Name: "chunk_1.bin", Size: 0})
	_ = tr.Close()
	_ = gz.Close()

	layers := []ocispec.Descriptor{
		{Digest: "sha256:a", Size: int64(gzBuf.Len()), MediaType: oci.MediaTypeChunkGzip, Annotations: map[string]string{"dockercomms.chunk.index": "0"}},
	}
	fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(gzBuf.Bytes())), nil
	}
	var buf bytes.Buffer
	_, err := Reassemble(layers, fetch, &buf)
	if err == nil {
		t.Fatal("expected error for extra entry")
	}
	if !strings.Contains(err.Error(), "extra tar entry") {
		t.Errorf("unexpected error: %v", err)
	}
}
