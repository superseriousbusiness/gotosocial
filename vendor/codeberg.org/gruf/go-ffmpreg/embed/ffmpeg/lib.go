package ffmpeg

import (
	_ "embed"
	"os"

	"github.com/tetratelabs/wazero/api"
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

// CoreFeatures is the WebAssembly Core specification
// features this embedded binary was compiled with.
const CoreFeatures = api.CoreFeatureSIMD |
	api.CoreFeatureBulkMemoryOperations |
	api.CoreFeatureNonTrappingFloatToIntConversion |
	api.CoreFeatureMutableGlobal |
	api.CoreFeatureReferenceTypes |
	api.CoreFeatureSignExtensionOps

//go:embed ffmpeg.wasm
var B []byte
