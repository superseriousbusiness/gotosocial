// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package ffmpeg

import (
	"context"

	ffmpeglib "codeberg.org/gruf/go-ffmpreg/embed/ffmpeg"
	"codeberg.org/gruf/go-ffmpreg/util"
	"codeberg.org/gruf/go-ffmpreg/wasm"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// InitFfmpeg initializes the ffmpeg WebAssembly instance pool,
// with given maximum limiting the number of concurrent instances.
func InitFfmpeg(ctx context.Context, max int) error {
	initCache() // ensure compilation cache initialized
	return ffmpegPool.Init(ctx, max)
}

// Ffmpeg runs the given arguments with an instance of ffmpeg.
func Ffmpeg(ctx context.Context, args wasm.Args) (uint32, error) {
	return ffmpegPool.Run(ctx, args)
}

var ffmpegPool = wasmInstancePool{
	inst: wasm.Instantiator{

		// WASM module name.
		Module: "ffmpeg",

		// Per-instance WebAssembly runtime (with shared cache).
		Runtime: func(ctx context.Context) wazero.Runtime {

			// Prepare config with cache.
			cfg := wazero.NewRuntimeConfig()
			cfg = cfg.WithCoreFeatures(ffmpeglib.CoreFeatures)
			cfg = cfg.WithCompilationCache(cache)

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
		Source: ffmpeglib.B,
	},
}
