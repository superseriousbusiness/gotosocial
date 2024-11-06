package embed

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"os"
)

func init() {
	var err error

	if path := os.Getenv("FFMPREG_WASM"); path != "" {
		// Read file into memory.
		B, err = os.ReadFile(path)
		if err != nil {
			panic(err)
		}
	}

	// Wrap bytes in reader.
	b := bytes.NewReader(B)

	// Create unzipper from reader.
	gz, err := gzip.NewReader(b)
	if err != nil {
		panic(err)
	}

	// Extract gzipped binary.
	B, err = io.ReadAll(gz)
	if err != nil {
		panic(err)
	}
}

//go:embed ffmpreg.wasm.gz
var B []byte
