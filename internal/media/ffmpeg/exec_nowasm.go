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

//go:build nowasm

package ffmpeg

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os/exec"

	"codeberg.org/gruf/go-ffmpreg/wasm"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/sys"
)

func init() {
	fmt.Println("!! you are using an unsupported build configuration of gotosocial with WebAssembly disabled !!")
	fmt.Println("!! please do not file bug reports regarding media processing with this configuration !!")
	fmt.Println("!! it is also less secure; this does not enforce version checks on ffmpeg / ffprobe versions !!")
}

// runCmd will run 'name' with the given arguments, returning exit code or error.
func runCmd(ctx context.Context, name string, args wasm.Args) (uint32, error) {
	cmd := exec.CommandContext(ctx, name, args.Args...) //nolint:gosec

	// Set provided std files.
	cmd.Stdin = args.Stdin
	cmd.Stdout = args.Stdout
	cmd.Stderr = args.Stderr

	if args.Config != nil {
		// Gather some information
		// from module config func.
		var cfg falseModuleConfig
		_ = args.Config(&cfg)

		// Extract from conf.
		cmd.Env = cfg.env
	}

	// Run prepared command, catching err type.
	switch err := cmd.Run(); err := err.(type) {

	// Extract code from
	// any exit error type.
	case *exec.ExitError:
		rc := err.ExitCode()
		return uint32(rc), err

	default:
		return 0, err
	}
}

type falseModuleConfig struct{ env []string }

func (cfg *falseModuleConfig) WithArgs(...string) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithEnv(key string, value string) wazero.ModuleConfig {
	cfg.env = append(cfg.env, key+"="+value)
	return cfg // noop
}

func (cfg *falseModuleConfig) WithFS(fs.FS) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithFSConfig(wazero.FSConfig) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithName(string) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithStartFunctions(...string) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithStderr(io.Writer) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithStdin(io.Reader) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithStdout(io.Writer) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithWalltime(sys.Walltime, sys.ClockResolution) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithSysWalltime() wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithNanotime(sys.Nanotime, sys.ClockResolution) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithSysNanotime() wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithNanosleep(sys.Nanosleep) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithOsyield(sys.Osyield) wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithSysNanosleep() wazero.ModuleConfig {
	return cfg // noop
}

func (cfg *falseModuleConfig) WithRandSource(io.Reader) wazero.ModuleConfig {
	return cfg // noop
}
