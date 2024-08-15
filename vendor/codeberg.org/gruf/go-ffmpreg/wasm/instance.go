package wasm

import (
	"context"
	"errors"
	"io"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/sys"
)

type Args struct {
	// Optional further module configuration function.
	// (e.g. to mount filesystem dir, set env vars, etc).
	Config func(wazero.ModuleConfig) wazero.ModuleConfig

	// Standard FDs.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// CLI args.
	Args []string
}

type Instantiator struct {
	// Module ...
	Module string

	// Runtime ...
	Runtime func(context.Context) wazero.Runtime

	// Config ...
	Config func() wazero.ModuleConfig

	// Source ...
	Source []byte
}

func (inst *Instantiator) New(ctx context.Context) (*Instance, error) {
	switch {
	case inst.Module == "":
		panic("missing module name")
	case inst.Runtime == nil:
		panic("missing runtime instantiator")
	case inst.Config == nil:
		panic("missing module configuration")
	case len(inst.Source) == 0:
		panic("missing module source")
	}

	// Create new host runtime.
	rt := inst.Runtime(ctx)

	// Compile guest module from WebAssembly source.
	mod, err := rt.CompileModule(ctx, inst.Source)
	if err != nil {
		return nil, err
	}

	return &Instance{
		inst: inst,
		wzrt: rt,
		cmod: mod,
	}, nil
}

// Instance ...
//
// NOTE: Instance is NOT concurrency
// safe. One at a time please!!
type Instance struct {
	inst *Instantiator
	wzrt wazero.Runtime
	cmod wazero.CompiledModule
}

func (inst *Instance) Run(ctx context.Context, args Args) (uint32, error) {
	if inst.inst == nil {
		panic("not initialized")
	}

	// Check instance open.
	if inst.IsClosed() {
		return 0, errors.New("instance closed")
	}

	// Prefix binary name as argv0 to args.
	cargs := make([]string, len(args.Args)+1)
	copy(cargs[1:], args.Args)
	cargs[0] = inst.inst.Module

	// Create base module config.
	modcfg := inst.inst.Config()
	modcfg = modcfg.WithName(inst.inst.Module)
	modcfg = modcfg.WithArgs(cargs...)
	modcfg = modcfg.WithStdin(args.Stdin)
	modcfg = modcfg.WithStdout(args.Stdout)
	modcfg = modcfg.WithStderr(args.Stderr)

	if args.Config != nil {
		// Pass through config fn.
		modcfg = args.Config(modcfg)
	}

	// Instantiate the module from precompiled wasm module data.
	mod, err := inst.wzrt.InstantiateModule(ctx, inst.cmod, modcfg)
	switch err := err.(type) {
	case nil:
		return 0, mod.Close(ctx)
	case *sys.ExitError:
		return err.ExitCode(), nil
	default:
		return 0, err
	}
}

func (inst *Instance) IsClosed() bool {
	return (inst.wzrt == nil || inst.cmod == nil)
}

func (inst *Instance) Close(ctx context.Context) error {
	if inst.IsClosed() {
		return nil
	}
	err1 := inst.cmod.Close(ctx)
	err2 := inst.wzrt.Close(ctx)
	return errors.Join(err1, err2)
}
