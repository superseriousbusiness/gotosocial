package ffprobe

import (
	_ "embed"
	"os"
)

func init() {
	// Check for WASM source file path.
	path := os.Getenv("FFPROBE_WASM")
	if path == "" {
		return
	}

	var err error

	// Read file into memory.
	B, err = os.ReadFile(path)
	if err != nil {
		panic(err)
	}
}

//go:embed ffprobe.wasm
var B []byte
