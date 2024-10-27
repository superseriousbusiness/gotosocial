package wasm

import (
	"context"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// CoreFeatures are the WebAssembly Core specification
// features our embedded binaries are compiled with.
const CoreFeatures = api.CoreFeatureSIMD |
	api.CoreFeatureBulkMemoryOperations |
	api.CoreFeatureNonTrappingFloatToIntConversion |
	api.CoreFeatureMutableGlobal |
	api.CoreFeatureReferenceTypes |
	api.CoreFeatureSignExtensionOps

// NewRuntime returns a new WebAssembly wazero.Runtime compatible with go-ffmpreg.
func NewRuntime(ctx context.Context, cfg wazero.RuntimeConfig) (wazero.Runtime, error) {
	var err error

	if cfg == nil {
		// Ensure runtime config is set.
		cfg = wazero.NewRuntimeConfig()
	}

	// Set core features ffmpeg compiled with.
	cfg = cfg.WithCoreFeatures(CoreFeatures)

	// Instantiate runtime with prepared config.
	rt := wazero.NewRuntimeWithConfig(ctx, cfg)

	// Prepare default "env" host module.
	env := rt.NewHostModuleBuilder("env")

	// Register setjmp host function.
	env = env.NewFunctionBuilder().
		WithGoModuleFunction(
			api.GoModuleFunc(setjmp),
			[]api.ValueType{api.ValueTypeI32},
			[]api.ValueType{api.ValueTypeI32},
		).Export("setjmp")

	// Register longjmp host function.
	env = env.NewFunctionBuilder().
		WithGoModuleFunction(
			api.GoModuleFunc(longjmp),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32},
			[]api.ValueType{},
		).Export("longjmp")

	// Instantiate "env" module.
	_, err = env.Instantiate(ctx)
	if err != nil {
		return nil, err
	}

	// Instantiate the wasi snapshot preview 1 in runtime.
	_, err = wasi_snapshot_preview1.Instantiate(ctx, rt)
	if err != nil {
		return nil, err
	}

	return rt, nil
}
