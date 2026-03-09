// Copyright 2025 DockerComms Authors
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var fuzzTagSafeRe = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)

func FuzzRecipientTag_NoInvalidCharsetOrLength(f *testing.F) {
	f.Add("alice@example.com")
	f.Add("bob")
	f.Add("")
	f.Add(strings.Repeat("x", 1000))
	f.Fuzz(func(t *testing.T, recipient string) {
		got := RecipientTag(recipient)
		if len(got) != 26 {
			t.Errorf("RecipientTag length = %d, want 26", len(got))
		}
		if !fuzzTagSafeRe.MatchString(got) {
			t.Errorf("RecipientTag(%q) = %q, not tag-safe", recipient, got)
		}
	})
}

func FuzzInboxTag_FormatAndLength(f *testing.F) {
	f.Add("alice@example.com", "20250101", "abc12345", "def67890")
	f.Fuzz(func(t *testing.T, recipient, date, sid, mid string) {
		rt := RecipientTag(recipient)
		tag := "inbox-" + rt + "-" + date + "-" + sid + "-" + mid
		if len(tag) > 128 {
			t.Errorf("inbox tag length %d exceeds 128", len(tag))
		}
	})
}

func FuzzSanitizeFilename_NoTraversal(f *testing.F) {
	f.Add("file.txt")
	f.Add("/path/to/file")
	f.Add("../../../etc/passwd")
	f.Add("..\\evil.txt")
	f.Add("")
	f.Fuzz(func(t *testing.T, path string) {
		got := SanitizeFilename(path)
		if strings.Contains(got, "/") {
			t.Errorf("SanitizeFilename(%q) = %q, must not contain slash", path, got)
		}
		if strings.Contains(got, "\\") {
			t.Errorf("SanitizeFilename(%q) = %q, must not contain backslash", path, got)
		}
		if got == ".." || got == "" {
			t.Errorf("SanitizeFilename(%q) = %q, must not be .. or empty", path, got)
		}
	})
}

func FuzzChunkReassemble_RoundTrip(f *testing.F) {
	f.Add([]byte("hello"), int64(10))
	f.Add([]byte("x"), int64(1))
	f.Add(make([]byte, 1000), int64(100))
	f.Fuzz(func(t *testing.T, data []byte, chunkSize int64) {
		if len(data) > 100*1024 {
			t.Skip("bounded: max 100KB to avoid resource exhaustion")
		}
		if chunkSize < 1 || chunkSize > 1024*1024 {
			t.Skip("bounded: chunk size 1-1MB")
		}
		c := NewChunker(chunkSize, CompressionGzip)
		descs, readers, err := c.Chunk(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			t.Fatal(err)
		}
		layers := make([]struct {
			d ocispec.Descriptor
			b []byte
		}, len(descs))
		for i, cd := range descs {
			layers[i].d = cd.Descriptor
			var err error
			layers[i].b, err = io.ReadAll(readers[i])
			if err != nil {
				t.Fatal(err)
			}
		}
		fetch := func(d ocispec.Descriptor) (io.ReadCloser, error) {
			for _, l := range layers {
				if l.d.Digest == d.Digest {
					return io.NopCloser(bytes.NewReader(l.b)), nil
				}
			}
			return nil, fmt.Errorf("not found")
		}
		var buf bytes.Buffer
		layerDescs := make([]ocispec.Descriptor, len(layers))
		for i := range layers {
			layerDescs[i] = layers[i].d
		}
		n, err := Reassemble(layerDescs, fetch, &buf)
		if err != nil {
			t.Fatal(err)
		}
		if n != int64(len(data)) {
			t.Errorf("reassembled %d, want %d", n, len(data))
		}
		if !bytes.Equal(buf.Bytes(), data) {
			t.Errorf("reassembled content mismatch")
		}
	})
}
