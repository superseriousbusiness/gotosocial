package ffmpeg

import (
	_ "embed"
	"os"
)

func init() {
	// Check for WASM source file path.
	path := os.Getenv("FFMPEG_WASM")
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

//go:embed ffmpeg.wasm
var B []byte
