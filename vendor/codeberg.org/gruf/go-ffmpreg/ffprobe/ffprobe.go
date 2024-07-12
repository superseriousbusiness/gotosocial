package ffprobe

import (
	"context"

	"codeberg.org/gruf/go-ffmpreg/embed/ffprobe"
	"codeberg.org/gruf/go-ffmpreg/internal"
	"codeberg.org/gruf/go-ffmpreg/util"
	"codeberg.org/gruf/go-ffmpreg/wasm"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// pool of WASM module instances.
var pool = wasm.InstancePool{
	Instantiator: wasm.Instantiator{

		// WASM module name.
		Module: "ffprobe",

		// Per-instance WebAssembly runtime (with shared cache).
		Runtime: func(ctx context.Context) wazero.Runtime {

			// Prepare config with cache.
			cfg := wazero.NewRuntimeConfig()
			cfg = cfg.WithCoreFeatures(ffprobe.CoreFeatures)
			cfg = cfg.WithCompilationCache(internal.Cache)

			// Instantiate runtime with our config.
			rt := wazero.NewRuntimeWithConfig(ctx, cfg)

			// Prepare default "env" host module.
			env := rt.NewHostModuleBuilder("env")
			env = env.NewFunctionBuilder().
				WithGoModuleFunction(
					api.GoModuleFunc(util.Wasm_Tempnam),
					[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32},
					[]api.ValueType{api.ValueTypeI32},
				).
				Export("tempnam")

			// Instantiate "env" module in our runtime.
			_, err := env.Instantiate(context.Background())
			if err != nil {
				panic(err)
			}

			// Instantiate the wasi snapshot preview 1 in runtime.
			_, err = wasi_snapshot_preview1.Instantiate(ctx, rt)
			if err != nil {
				panic(err)
			}

			return rt
		},

		// Per-run module configuration.
		Config: wazero.NewModuleConfig,

		// Embedded WASM.
		Source: ffprobe.B,
	},
}

// Precompile ensures at least compiled ffprobe
// instance is available in the global pool.
func Precompile(ctx context.Context) error {
	inst, err := pool.Get(ctx)
	if err != nil {
		return err
	}
	pool.Put(inst)
	return nil
}

// Get fetches new ffprobe instance from pool, prefering cached if available.
func Get(ctx context.Context) (*wasm.Instance, error) { return pool.Get(ctx) }

// Put places the given ffprobe instance in pool.
func Put(inst *wasm.Instance) { pool.Put(inst) }

// Run will run the given args against an ffprobe instance from pool.
func Run(ctx context.Context, args wasm.Args) (uint32, error) {
	inst, err := pool.Get(ctx)
	if err != nil {
		return 0, err
	}
	rc, err := inst.Run(ctx, args)
	pool.Put(inst)
	return rc, err
}

// Cached returns a cached instance (if any) from pool.
func Cached() *wasm.Instance { return pool.Cached() }

// Free drops all instances
// cached in instance pool.
func Free() {
	ctx := context.Background()
	for {
		inst := pool.Cached()
		if inst == nil {
			return
		}
		_ = inst.Close(ctx)
	}
}
