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

	ffmpeglib "codeberg.org/gruf/go-ffmpreg/embed/ffmpeg"
	ffprobelib "codeberg.org/gruf/go-ffmpreg/embed/ffprobe"
	"codeberg.org/gruf/go-ffmpreg/wasm"
	"github.com/tetratelabs/wazero"
)

var (
	// shared WASM runtime instance.
	runtime wazero.Runtime

	// ffmpeg / ffprobe compiled WASM.
	ffmpeg  wazero.CompiledModule
	ffprobe wazero.CompiledModule
)

// compileFfmpeg ensures the ffmpeg WebAssembly has been
// pre-compiled into memory. If already compiled is a no-op.
func compileFfmpeg(ctx context.Context) error {
	if ffmpeg != nil {
		return nil
	}

	// Ensure runtime already initialized.
	if err := initRuntime(ctx); err != nil {
		return err
	}

	// Compile the ffmpeg WebAssembly module into memory.
	cmod, err := runtime.CompileModule(ctx, ffmpeglib.B)
	if err != nil {
		return err
	}

	// Set module.
	ffmpeg = cmod
	return nil
}

// compileFfprobe ensures the ffprobe WebAssembly has been
// pre-compiled into memory. If already compiled is a no-op.
func compileFfprobe(ctx context.Context) error {
	if ffprobe != nil {
		return nil
	}

	// Ensure runtime already initialized.
	if err := initRuntime(ctx); err != nil {
		return err
	}

	// Compile the ffprobe WebAssembly module into memory.
	cmod, err := runtime.CompileModule(ctx, ffprobelib.B)
	if err != nil {
		return err
	}

	// Set module.
	ffprobe = cmod
	return nil
}

// initRuntime initializes the global wazero.Runtime,
// if already initialized this function is a no-op.
func initRuntime(ctx context.Context) (err error) {
	if runtime != nil {
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

	// Initialize new runtime from config.
	runtime, err = wasm.NewRuntime(ctx, cfg)
	return
}
