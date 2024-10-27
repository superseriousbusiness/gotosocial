package wasm

import (
	"context"
	"io"
	"unsafe"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/sys"
)

// Args encompasses a common set of
// configuration options often passed to
// wazero.Runtime on module instantiation.
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

// Run will run given compiled WebAssembly module
// within the given runtime, with given arguments.
// Returns the exit code, or error.
func Run(
	ctx context.Context,
	runtime wazero.Runtime,
	module wazero.CompiledModule,
	args Args,
) (rc uint32, err error) {

	// Prefix arguments with module name.
	cargs := make([]string, len(args.Args)+1)
	cargs[0] = module.Name()
	copy(cargs[1:], args.Args)

	// Prepare new module configuration.
	modcfg := wazero.NewModuleConfig()
	modcfg = modcfg.WithArgs(cargs...)
	modcfg = modcfg.WithStdin(args.Stdin)
	modcfg = modcfg.WithStdout(args.Stdout)
	modcfg = modcfg.WithStderr(args.Stderr)

	if args.Config != nil {
		// Pass through config fn.
		modcfg = args.Config(modcfg)
	}

	// Enable setjmp longjmp.
	ctx = withSetjmpLongjmp(ctx)

	// Instantiate the module from precompiled wasm module data.
	mod, err := runtime.InstantiateModule(ctx, module, modcfg)

	if !isNil(mod) {
		// Ensure closed.
		_ = mod.Close(ctx)
	}

	// Try extract exit code.
	switch err := err.(type) {
	case *sys.ExitError:
		return err.ExitCode(), nil
	default:
		return 0, err
	}
}

// isNil will safely check if 'v' is nil without
// dealing with weird Go interface nil bullshit.
func isNil(i interface{}) bool {
	type eface struct{ Type, Data unsafe.Pointer }
	return (*(*eface)(unsafe.Pointer(&i))).Data == nil
}
