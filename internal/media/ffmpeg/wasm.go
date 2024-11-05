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

//go:build !nowasm

package ffmpeg

import (
	"context"
	"os"
	"sync/atomic"
	"unsafe"

	"codeberg.org/gruf/go-ffmpreg/embed"
	"codeberg.org/gruf/go-ffmpreg/wasm"
	"github.com/tetratelabs/wazero"
)

// ffmpreg is a concurrency-safe pointer
// to our necessary WebAssembly runtime
// and compiled ffmpreg module instance.
var ffmpreg atomic.Pointer[struct {
	run wazero.Runtime
	mod wazero.CompiledModule
}]

// initWASM safely prepares new WebAssembly runtime
// and compiles ffmpreg module instance, if the global
// pointer has not been already. else, is a no-op.
func initWASM(ctx context.Context) error {
	if ffmpreg.Load() != nil {
		return nil
	}

	// Create new runtime config.
	cfg := wazero.NewRuntimeConfig()

	if dir := os.Getenv("GTS_WAZERO_COMPILATION_CACHE"); dir != "" {
		// Use on-filesystem compilation cache given by env.
		cache, err := wazero.NewCompilationCacheWithDir(dir)
		if err != nil {
			return err
		}

		// Update runtime config with cache.
		cfg = cfg.WithCompilationCache(cache)
	}

	var (
		run wazero.Runtime
		mod wazero.CompiledModule
		err error
		set bool
	)

	defer func() {
		if err == nil && set {
			// Drop binary.
			embed.B = nil
			return
		}

		// Close module.
		if !isNil(mod) {
			mod.Close(ctx)
		}

		// Close runtime.
		if !isNil(run) {
			run.Close(ctx)
		}
	}()

	// Initialize new runtime from config.
	run, err = wasm.NewRuntime(ctx, cfg)
	if err != nil {
		return err
	}

	// Compile ffmpreg WebAssembly into memory.
	mod, err = run.CompileModule(ctx, embed.B)
	if err != nil {
		return err
	}

	// Try set global WASM runtime and module,
	// or if beaten to it defer will handle close.
	set = ffmpreg.CompareAndSwap(nil, &struct {
		run wazero.Runtime
		mod wazero.CompiledModule
	}{
		run: run,
		mod: mod,
	})

	return nil
}

// isNil will safely check if 'v' is nil without
// dealing with weird Go interface nil bullshit.
func isNil(i interface{}) bool {
	type eface struct{ Type, Data unsafe.Pointer }
	return (*eface)(unsafe.Pointer(&i)).Data == nil
}
