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

//go:build !nowasmffmpeg && !nowasm

package ffmpeg

import (
	"context"
	"os"
	"runtime"
	"sync/atomic"
	"unsafe"

	"codeberg.org/gruf/go-ffmpreg/embed"
	"codeberg.org/gruf/go-ffmpreg/wasm"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/tetratelabs/wazero"
	"golang.org/x/sys/cpu"
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

	var cfg wazero.RuntimeConfig

	// Allocate new runtime config, letting
	// wazero determine compiler / interpreter.
	cfg = wazero.NewRuntimeConfig()

	// Though still perform a check of CPU features at
	// runtime to warn about slow interpreter performance.
	if reason, supported := compilerSupported(); !supported {
		log.Warn(ctx, "!!! WAZERO COMPILER MAY NOT BE AVAILABLE !!!"+
			" Reason: "+reason+"."+
			" Wazero will likely fall back to interpreter mode,"+
			" resulting in poor performance for media processing (and SQLite, if in use)."+
			" For more info and possible workarounds, please check:"+
			" https://docs.gotosocial.org/en/latest/getting_started/releases/#supported-platforms",
		)
	}

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

func compilerSupported() (string, bool) {
	switch runtime.GOOS {
	case "linux", "android",
		"windows", "darwin",
		"freebsd", "netbsd", "dragonfly",
		"solaris", "illumos":
		break
	default:
		return "unsupported OS", false
	}
	switch runtime.GOARCH {
	case "amd64":
		// NOTE: wazero in the future may decouple the
		// requirement of simd (sse4_1) from requirements
		// for compiler support in the future, but even
		// still our module go-ffmpreg makes use of them.
		return "amd64 SSE4.1 required", cpu.X86.HasSSE41
	case "arm64":
		// NOTE: this particular check may change if we
		// later update go-ffmpreg to a version that makes
		// use of threads, i.e. v7.x.x. in that case we would
		// need to check for cpu.ARM64.HasATOMICS.
		return "", true
	default:
		return "unsupported ARCH", false
	}
}

// isNil will safely check if 'v' is nil without
// dealing with weird Go interface nil bullshit.
func isNil(i interface{}) bool {
	type eface struct{ Type, Data unsafe.Pointer }
	return (*eface)(unsafe.Pointer(&i)).Data == nil
}
