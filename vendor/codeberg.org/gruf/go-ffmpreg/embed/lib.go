package embed

import (
	"compress/gzip"
	_ "embed"
	"io"
	"strings"
)

func init() {
	var err error

	// Wrap bytes in reader.
	r := strings.NewReader(s)

	// Create unzipper from reader.
	gz, err := gzip.NewReader(r)
	if err != nil {
		panic(err)
	}

	// Extract gzipped binary.
	b, err := io.ReadAll(gz)
	if err != nil {
		panic(err)
	}

	// Set binary.
	s = string(b)
}

// B returns a copy of
// embedded binary data.
func B() []byte {
	if s == "" {
		panic("binary already dropped from memory")
	}
	return []byte(s)
}

// Free will drop embedded
// binary from runtime mem.
func Free() { s = "" }

//go:embed ffmpreg.wasm.gz
var s string
